package watchers

import (
    "errors"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "sync"

    "github.com/fsnotify/fsnotify"

    "github.com/iplay88keys/watchtower/pkg/runners"
)

const SHOULD_UPDATE_EVENT = uint32(fsnotify.Remove) | uint32(fsnotify.Rename)| uint32(fsnotify.Create)

type Path struct {
    Paths      []string `json:"paths"`
    Recursive  bool     `json:"recursive"`
    Exclusions []string `json:"exclusions"`
    Events     []string `json:"events"`
}

type PathWatcher struct {
    watcher *fsnotify.Watcher

    paths []pathConfig
    done  chan struct{}
    quit  chan struct{}

    mu            sync.RWMutex
    handlingEvent bool
}

type pathConfig struct {
    Path

    name          string
    foundPaths    map[string]bool
    desiredEvents uint32
    runners       []*runners.Config
}

func NewPathWatcher() (*PathWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    return &PathWatcher{
        watcher: watcher,
        done:    make(chan struct{}, 1),
        quit:    make(chan struct{}, 1),
    }, nil
}

func (w *PathWatcher) Add(path Path, runnerConfigs []*runners.Config, name string) error {
    fmt.Printf("Adding path watchers for '%s'\n", name)

    events, err := desiredEvents(path.Events)
    if err != nil {
        return err
    }

    pc := pathConfig{
        Path:          path,
        desiredEvents: events,
        runners:       runnerConfigs,
        name:          name,
    }

    for _, exclusion := range path.Exclusions {
        fmt.Println("Excluding:", exclusion)
    }

    foundPaths, err := w.updatePathsAndWatchers(path.Paths, path.Exclusions, path.Recursive, nil)
    if err != nil {
        return err
    }

    pc.foundPaths = foundPaths

    w.paths = append(w.paths, pc)

    fmt.Println()

    return nil
}

func (w *PathWatcher) Watch() (func(), chan struct{}) {
    events := make(chan fsnotify.Event)
    errors := make(chan error)

    go w.watch(events, errors)
    go w.handleEvents(events, errors)

    return func() {
        w.stop()
    }, w.quit
}

func (w *PathWatcher) watch(events chan fsnotify.Event, errors chan error) {
    defer close(w.quit)
    fmt.Println("Awaiting Events...")

    for {
        select {
        case <-w.done:
            close(events)
            close(errors)

            return
        case event, ok := <-w.watcher.Events:
            if !ok {
                return
            }

            w.mu.RLock()
            handling := w.handlingEvent
            w.mu.RUnlock()

            if handling {
                continue
            }

            w.updateHandlingEvent(true)

            events <- event
        case err, ok := <-w.watcher.Errors:
            if !ok {
                return
            }

            fmt.Println(err)
            return
        }
    }
}

func (w *PathWatcher) handleEvents(events chan fsnotify.Event, errors chan error) {
    for {
        select {
        case event, ok := <-events:
            if !ok {
                return
            }

            absFileLoc, err := filepath.Abs(event.Name)
            if err != nil {
                fmt.Printf("could not get absolute path for '%s'", event.Name)
                return
            }

            basePath, err := os.Getwd()
            if err != nil {
                fmt.Printf("could not get base path for '%s'", event.Name)
                return
            }

            for configInd, config := range w.paths {
                var cont bool
                for _, exclusion := range config.Exclusions {
                    relativePath := strings.TrimPrefix(event.Name, basePath+string(filepath.Separator))
                    matched, err := regexp.MatchString(exclusion, relativePath)
                    if err != nil {
                        fmt.Printf("Error testing exclusion '%s': %s", exclusion, err.Error())
                        return
                    }

                    if matched {
                        cont = true
                        break
                    }
                }

                if cont {
                    continue
                }

                var found, foundExact, shouldUpdate bool
                for foundPath := range config.foundPaths {
                    if foundPath == absFileLoc {
                        if config.desiredEvents&uint32(event.Op) != 0 {
                            if SHOULD_UPDATE_EVENT&uint32(event.Op) != 0 {
                                shouldUpdate = true
                            }

                            found = true
                            foundExact = true

                            err := w.executeRunners(config, absFileLoc, event.Op)
                            if err != nil {
                                fmt.Println("Error running: ", err.Error())

                                w.updateHandlingEvent(false)
                                return
                            }
                        }
                    }
                }

                if !found {
                    eventDepth := len(strings.Split(absFileLoc, string(filepath.Separator)))
                    for foundPath := range config.foundPaths {
                        foundDepth := len(strings.Split(foundPath, string(filepath.Separator)))
                        if (!config.Recursive && strings.Contains(absFileLoc, foundPath) && foundDepth == eventDepth-1) || (strings.Contains(absFileLoc, foundPath) && config.Recursive) {
                            if config.desiredEvents&uint32(event.Op) != 0 {
                                if SHOULD_UPDATE_EVENT&uint32(event.Op) != 0 {
                                    shouldUpdate = true
                                }

                                err := w.executeRunners(config, absFileLoc, event.Op)
                                if err != nil {
                                    fmt.Println("Error running: ", err.Error())

                                    w.updateHandlingEvent(false)
                                    return
                                }

                                break
                            }
                        }
                    }
                }

                if !foundExact || shouldUpdate {
                    foundPaths, err := w.updatePathsAndWatchers(config.Paths, config.Exclusions, config.Recursive, config.foundPaths)
                    if err != nil {
                        fmt.Printf("Error updating paths for '%s': %s", config.name, err.Error())

                        return
                    }

                    w.paths[configInd].foundPaths = foundPaths
                }
            }

            w.updateHandlingEvent(false)
        case err, ok := <-errors:
            if !ok {
                return
            }

            fmt.Println(err)
            return
        }
    }
}

func (w *PathWatcher) executeRunners(config pathConfig, filePath string, op fsnotify.Op) error {
    fmt.Printf("\n---------------------------------------\n")
    fmt.Printf("Event matched for '%s': %s, %v\n\n", config.name, filePath, op)

    for _, runner := range config.runners {
        err := runner.Config.Execute(filePath)
        if err != nil {
            return err
        }
    }

    return nil
}

func (w *PathWatcher) updateHandlingEvent(status bool) {
    w.mu.Lock()
    w.handlingEvent = status
    w.mu.Unlock()
}

func (w *PathWatcher) stop() {
    close(w.done)
}

func (w *PathWatcher) updatePathsAndWatchers(roots []string, exclusions []string, recursive bool, prevFoundPaths map[string]bool) (map[string]bool, error) {
    foundPaths := make(map[string]bool)

    for _, root := range roots {
        maxDepth := -1

        absRoot, err := filepath.Abs(root)
        if err != nil {
            return nil, fmt.Errorf("could not get absolute path for '%s' during walk: %s", root, err.Error())
        }

        if !recursive {
            maxDepth = len(strings.Split(absRoot, string(filepath.Separator))) + 1
        }

        err = filepath.Walk(root, func(fileLoc string, info fs.FileInfo, err error) error {
            if err != nil {
                return errors.New(fmt.Sprintf("walk error for '%s': %s", fileLoc, err))
            }

            absFileLoc, err := filepath.Abs(fileLoc)
            if err != nil {
                return fmt.Errorf("could not get absolute path for '%s' during walk: %s", err, err.Error())
            }

            depth := len(strings.Split(absFileLoc, string(filepath.Separator)))

            if maxDepth != -1 {
                if depth > maxDepth {
                    return nil
                }
            }

            for _, exclusion := range exclusions {
                matched, err := regexp.MatchString(exclusion, fileLoc)
                if err != nil {
                    return fmt.Errorf("exclusion '%s' is an invalid regular expression: %s", exclusion, err.Error())
                }

                if matched {
                    return nil
                }
            }

            err = w.watcher.Add(absFileLoc)
            if err != nil {
                fmt.Printf("Failed to add '%s': %s\n", absFileLoc, err.Error())
                return nil
            }

            foundPaths[absFileLoc] = true
            if _, inPrevFound := prevFoundPaths[absFileLoc]; !inPrevFound {
                fmt.Println("Added:", absFileLoc)
            }

            return nil
        })

        if err != nil {
            return nil, err
        }
    }

    for prevPath := range prevFoundPaths {
        if _, ok := foundPaths[prevPath]; !ok {
            fmt.Println("Removed:", prevPath)
        }
    }

    return foundPaths, nil
}

func desiredEvents(events []string) (uint32, error) {
    var desiredEvents uint32

    if len(events) == 0 {
        return uint32(fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename | fsnotify.Chmod), nil
    }

    for _, event := range events {
        switch strings.ToLower(event) {
        case "create":
            desiredEvents = desiredEvents | uint32(fsnotify.Create)
        case "write":
            desiredEvents = desiredEvents | uint32(fsnotify.Write)
        case "remove":
            desiredEvents = desiredEvents | uint32(fsnotify.Remove)
        case "rename":
            desiredEvents = desiredEvents | uint32(fsnotify.Rename)
        case "chmod":
            desiredEvents = desiredEvents | uint32(fsnotify.Chmod)
        default:
            return 0, errors.New("event must be one of: 'Create', 'Write', 'Remove', 'Rename', or 'Chmod'")
        }
    }

    return desiredEvents, nil
}
