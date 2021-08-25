package runners_test

import (
    "io/ioutil"
    "os"
    "os/exec"
    "testing"

    "github.com/iplay88keys/watchtower/pkg/runners"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Process", func() {
    It("starts a background using the start command", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:     "test",
            Type:     "background",
            StartCmd: "start_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err = proc.Start()
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' start command: 'start_command'\n\n"))
    })

    It("only starts a background process once", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:     "test",
            Type:     "background",
            StartCmd: "start_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err = proc.Start()
        Expect(err).ToNot(HaveOccurred())

        err = proc.Start()
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' start command: 'start_command'\n\nProcess is already running\n"))
    })

    It("starts a task using the start command", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:     "test",
            Type:     "task",
            StartCmd: "start_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err = proc.Start()
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' start command: 'start_command'\n\n"))
    })

    It("stops a process using the stop command", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:     "test",
            Type:     "background",
            StartCmd: "start_command",
            StopCmd:  "stop_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err = proc.Start()
        Expect(err).ToNot(HaveOccurred())

        err = proc.Stop()
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' start command: 'start_command'\n\nRunning 'test' stop command: 'stop_command'\n\n"))
    })

    It("stops a process by killing it", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:     "test",
            Type:     "background",
            StartCmd: "watch ls",
        }

        err = proc.Start()
        Expect(err).ToNot(HaveOccurred())

        err = proc.Stop()
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' start command: 'watch ls'\n\nKilling process\n"))
    })

    It("restarts a process by stopping and starting it", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:       "test",
            Type:       "background",
            StartCmd:   "start_command",
            StopCmd:    "stop_command",
            CleanupCmd: "cleanup_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err = proc.Restart(false)
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' stop command: 'stop_command'\n\nRunning 'test' start command: 'start_command'\n\n"))
    })

    It("restarts a process and runs the cleanup command", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:       "test",
            Type:       "background",
            StartCmd:   "start_command",
            StopCmd:    "stop_command",
            CleanupCmd: "cleanup_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err = proc.Restart(true)
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' stop command: 'stop_command'\n\nRunning 'test' cleanup command: 'cleanup_command'\n\nRunning 'test' start command: 'start_command'\n\n"))
    })

    It("restarts a process using the restart command", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        proc := runners.Process{
            Name:       "test",
            Type:       "background",
            StartCmd:   "start_command",
            RestartCmd: "restart_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err = proc.Restart(false)
        Expect(err).ToNot(HaveOccurred())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(Equal("Running 'test' restart command: 'restart_command'\n\n"))
    })

    It("returns an error if the process type is invalid", func() {
        osStdout := os.Stdout
        osStderr := os.Stderr

        os.Stdout = nil
        os.Stderr = nil

        proc := runners.Process{
            Name:     "test",
            Type:     "invalid",
            StartCmd: "start_command",
        }

        proc.UpdateExecContext(fakeExecCommandSuccess)

        err := proc.Start()
        Expect(err).To(HaveOccurred())

        os.Stdout = osStdout
        os.Stderr = osStderr
    })
})

// https://jamiethompson.me/posts/Unit-Testing-Exec-Command-In-Golang/
func TestShellProcessSuccess(t *testing.T) {
    RegisterTestingT(t)
    if os.Getenv("GO_TEST_PROCESS") != "1" {
        return
    }

    os.Exit(0)
}

func fakeExecCommandSuccess(command string, args ...string) *exec.Cmd {
    cs := []string{"-test.run=TestShellProcessSuccess", "--", command}
    cs = append(cs, args...)

    cmd := exec.Command(os.Args[0], cs...)

    cmd.Env = []string{"GO_TEST_PROCESS=1"}

    return cmd
}
