package main

import (
    "errors"
    "flag"
    "fmt"
    "github.com/iplay88keys/watchtower/pkg/runners"
    "github.com/iplay88keys/watchtower/pkg/watchers"
    "os"
    "os/signal"
    "syscall"

    "github.com/iplay88keys/watchtower/pkg/config"
)

var configFile string

func main() {
    flag.StringVar(&configFile, "config-file", "", "a string var")

    flag.Parse()

    if configFile == "" {
        panic(errors.New("-config-file param required"))
    }

    cfg, err := config.Load(configFile)
    if err != nil {
        panic(err)
    }

    stop, quit, err := setupWatchers(cfg)
    defer stop()
    if err != nil {
        panic(err)
    }

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)

    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        for {
            select {
            case sig := <-sigs:
                fmt.Println()
                fmt.Println(sig)
                done <- true
            case <-quit:
                done <- true
            }
        }
    }()

    <-done
    fmt.Println("Exiting")
}

func setupWatchers(cfg *config.Config) (func(), chan struct{}, error) {
    pathWatcher, err := watchers.NewPathWatcher()
    if err != nil {
        return nil, nil, err
    }

    for _, watch := range cfg.Watches {
        var triggers []*runners.Config
        for _, trigger := range watch.OnTrigger {
            restartRunnerConfig, ok := trigger.Config.(*runners.Restart)
            if ok {
                for _, process := range cfg.Processes {
                    if process.Name == restartRunnerConfig.Restart {
                        store := process
                        restartRunnerConfig.Setup(&store)
                    }
                }
            }

            trig := trigger
            triggers = append(triggers, &trig)
        }

        pathWatcherConfig, ok := watch.Config.Config.(*watchers.Path)
        if ok {
            err = pathWatcher.Add(*pathWatcherConfig, triggers, watch.Name)
            if err != nil {
                return nil, nil, err
            }
        }
    }

    fmt.Println("Running startup processes:")
    for _, process := range cfg.Processes {
        err := process.Start()
        if err != nil {
            return nil, nil, err
        }
    }

    stop, quit := pathWatcher.Watch()

    return stop, quit, nil
}
