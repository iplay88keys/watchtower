package runners

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "os/exec"
    "strings"
)

type execContext = func(name string, arg ...string) *exec.Cmd

type Process struct {
    Name       string `json:"name"`
    Type       string `json:"type"`
    StartCmd   string `json:"start"`
    StopCmd    string `json:"stop"`
    RestartCmd string `json:"restart"`
    CleanupCmd string `json:"cleanup"`

    execContext execContext
    process     *exec.Cmd
}

func (p *Process) UpdateExecContext(context execContext) {
    p.execContext = context
}

func (p *Process) Start() error {
    if p.execContext == nil {
        p.execContext = exec.Command
    }

    if p.process != nil && p.process.Process != nil {
        fmt.Println("Process is already running")
        return nil
    }

    err := p.execute(p.Type, p.StartCmd, "start")
    if err != nil {
        return err
    }

    return nil
}

func (p *Process) Stop() error {
    if p.execContext == nil {
        p.execContext = exec.Command
    }

    if p.StopCmd != "" {
        err := p.execute("task", p.StopCmd, "stop")
        if err != nil {
            return err
        }
    }

    if p.process != nil && p.process.Process != nil {
        fmt.Println("Killing process")
        err := p.process.Process.Kill()
        if err != nil {
            return err
        }
    }

    p.process = nil

    return nil
}

func (p *Process) Cleanup() error {
    if p.execContext == nil {
        p.execContext = exec.Command
    }

    if p.CleanupCmd != "" {
        err := p.execute("task", p.CleanupCmd, "cleanup")
        if err != nil {
            return err
        }
    }

    return nil
}

func (p *Process) Restart(runCleanup bool) error {
    if p.execContext == nil {
        p.execContext = exec.Command
    }

    if p.RestartCmd != "" {
        err := p.execute("task", p.RestartCmd, "restart")
        if err != nil {
            return err
        }

        return nil
    }

    err := p.Stop()
    if err != nil {
        return err
    }

    if runCleanup {
        err = p.Cleanup()
        if err != nil {
            return err
        }
    }

    err = p.Start()
    if err != nil {
        return err
    }

    return nil
}

func (p *Process) execute(commandType, command, commandUse string) error {
    var stdBuffer bytes.Buffer
    mw := io.MultiWriter(os.Stdout, &stdBuffer)

    args := []string{"-c", command}
    cmd := p.execContext("bash", args...)

    cmd.Stdout = mw
    cmd.Stderr = mw

    p.process = cmd

    var nameInfo string
    if p.Name != "" {
        nameInfo = fmt.Sprintf(" '%s' %s command", p.Name, commandUse)
    }

    fmt.Printf("Running%s: '%s'\n", nameInfo, strings.Join(args[1:], " "))

    switch commandType {
    case "background":
        err := p.process.Start()
        if err != nil {
            return err
        }
    case "task":
        //todo: timeout
        err := p.process.Run()
        if err != nil {
            return err
        }

        p.process = nil
    default:
        return fmt.Errorf("valid process types are: 'background' and 'task'")
    }

    fmt.Println()

    return nil
}
