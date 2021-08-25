package watchers_test

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/iplay88keys/watchtower/pkg/runners"
    "github.com/iplay88keys/watchtower/pkg/watchers"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Path", func() {
    It("watches a path for file changes", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        tmpDir, err := ioutil.TempDir("", "*")
        Expect(err).ToNot(HaveOccurred())

        f, err := os.Create(filepath.Join(tmpDir, "test"))
        Expect(err).ToNot(HaveOccurred())

        pw, err := watchers.NewPathWatcher()
        Expect(err).ToNot(HaveOccurred())

        p := watchers.Path{
            Paths: []string{
                f.Name(),
            },
            Recursive:  false,
            Exclusions: nil,
            Events: []string{
                "create",
                "write",
                "remove",
                "rename",
                "chmod",
            },
        }

        runner := []*runners.Config{{
            Config: &runners.Run{
                Run:             []string{"echo 'called'"},
                ContinueOnError: false,
            },
        }}

        err = pw.Add(p, runner, "")
        Expect(err).ToNot(HaveOccurred())

        stop, quit := pw.Watch()

        _, err = f.WriteString("update")
        Expect(err).ToNot(HaveOccurred())

        err = f.Sync()
        Expect(err).ToNot(HaveOccurred())

        err = f.Close()
        Expect(err).ToNot(HaveOccurred())

        stop()

        Eventually(quit, 15).Should(BeClosed())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Expect(string(out)).To(ContainSubstring(fmt.Sprintf("Added: %s", f.Name())))
        Expect(string(out)).To(ContainSubstring("Running: 'echo 'called''"))
    })

    It("recursively watches a path for file changes", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        tmpDir, err := ioutil.TempDir("", "*")
        Expect(err).ToNot(HaveOccurred())

        err = os.Mkdir(filepath.Join(tmpDir, "tmp"), os.ModePerm)
        Expect(err).ToNot(HaveOccurred())

        pw, err := watchers.NewPathWatcher()
        Expect(err).ToNot(HaveOccurred())

        p := watchers.Path{
            Paths: []string{
                tmpDir,
            },
            Recursive:  true,
            Exclusions: nil,
            Events: []string{
                "create",
                "write",
                "remove",
                "rename",
                "chmod",
            },
        }

        runner := []*runners.Config{{
            Config: &runners.Run{
                Run:             []string{"echo 'called'"},
                ContinueOnError: false,
            },
        }}

        err = pw.Add(p, runner, "")
        Expect(err).ToNot(HaveOccurred())

        stop, quit := pw.Watch()

        f, err := os.Create(filepath.Join(tmpDir, "tmp", "test"))
        Expect(err).ToNot(HaveOccurred())

        err = f.Close()
        Expect(err).ToNot(HaveOccurred())

        time.Sleep(200 * time.Millisecond)

        stop()

        Eventually(quit, 15).Should(BeClosed())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(ContainSubstring(fmt.Sprintf("Added: %s", tmpDir)))
        Eventually(string(out)).Should(ContainSubstring("Running: 'echo 'called''"))
    })

    It("adds sub directors to a path that are created when recursive is set", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        tmpDir, err := ioutil.TempDir("", "*")
        Expect(err).ToNot(HaveOccurred())

        pw, err := watchers.NewPathWatcher()
        Expect(err).ToNot(HaveOccurred())

        p := watchers.Path{
            Paths: []string{
                tmpDir,
            },
            Recursive:  true,
            Exclusions: nil,
            Events: []string{
                "create",
                "write",
                "remove",
                "rename",
                "chmod",
            },
        }

        runner := []*runners.Config{{
            Config: &runners.Run{
                Run:             []string{"echo 'called'"},
                ContinueOnError: false,
            },
        }}

        err = pw.Add(p, runner, "")
        Expect(err).ToNot(HaveOccurred())

        stop, quit := pw.Watch()

        err = os.Mkdir(filepath.Join(tmpDir, "tmp"), os.ModePerm)
        Expect(err).ToNot(HaveOccurred())

        time.Sleep(200 * time.Millisecond)

        f, err := os.Create(filepath.Join(tmpDir, "tmp", "test"))
        Expect(err).ToNot(HaveOccurred())

        err = f.Close()
        Expect(err).ToNot(HaveOccurred())

        time.Sleep(200 * time.Millisecond)

        stop()

        Eventually(quit, 15).Should(BeClosed())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(ContainSubstring(fmt.Sprintf("Added: %s", tmpDir)))
        Eventually(string(out)).Should(ContainSubstring("Running: 'echo 'called''"))
        Expect(strings.Count(string(out), "Running: 'echo 'called''")).To(Equal(2))
    })

    It("ignores excluded files", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        tmpDir, err := ioutil.TempDir("", "*")
        Expect(err).ToNot(HaveOccurred())

        f, err := os.Create(filepath.Join(tmpDir, "test"))
        Expect(err).ToNot(HaveOccurred())

        f2, err := os.Create(filepath.Join(tmpDir, "another"))
        Expect(err).ToNot(HaveOccurred())

        pw, err := watchers.NewPathWatcher()
        Expect(err).ToNot(HaveOccurred())

        p := watchers.Path{
            Paths: []string{
                tmpDir,
            },
            Recursive: true,
            Exclusions: []string{
                f.Name(),
                f2.Name(),
            },
            Events: []string{
                "create",
                "write",
            },
        }

        runner := []*runners.Config{{
            Config: &runners.Run{
                Run:             []string{"echo '{{.Name}}'"},
                ContinueOnError: false,
            },
        }}

        err = pw.Add(p, runner, "watcher1")
        Expect(err).ToNot(HaveOccurred())

        stop, quit := pw.Watch()

        _, err = f.WriteString("test")
        Expect(err).ToNot(HaveOccurred())

        err = f.Sync()
        Expect(err).ToNot(HaveOccurred())

        err = f.Close()
        Expect(err).ToNot(HaveOccurred())

        _, err = f2.WriteString("test2")
        Expect(err).ToNot(HaveOccurred())

        err = f2.Sync()
        Expect(err).ToNot(HaveOccurred())

        err = f2.Close()
        Expect(err).ToNot(HaveOccurred())

        time.Sleep(200 * time.Millisecond)

        stop()

        Eventually(quit, 15).Should(BeClosed())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Eventually(string(out)).Should(ContainSubstring(fmt.Sprintf("Added: %s", tmpDir)))
        Eventually(string(out)).Should(ContainSubstring(fmt.Sprintf("Excluding: %s", f.Name())))
        Eventually(string(out)).Should(ContainSubstring(fmt.Sprintf("Excluding: %s", f2.Name())))
        Eventually(string(out)).ShouldNot(ContainSubstring(fmt.Sprintf("Running: 'echo '%s''", f.Name())))
        Eventually(string(out)).ShouldNot(ContainSubstring(fmt.Sprintf("Running: 'echo '%s''", filepath.Join(tmpDir, "another"))))
    })

    It("skips handling a event if another event is being handled", func() {
        stdout := os.Stdout
        r, w, err := os.Pipe()
        Expect(err).ToNot(HaveOccurred())
        os.Stdout = w

        tmpDir, err := ioutil.TempDir("", "*")
        Expect(err).ToNot(HaveOccurred())

        pw, err := watchers.NewPathWatcher()
        Expect(err).ToNot(HaveOccurred())

        p := watchers.Path{
            Paths: []string{
                tmpDir,
            },
            Recursive:  true,
            Exclusions: nil,
            Events: []string{
                "create",
                "write",
                "remove",
                "rename",
                "chmod",
            },
        }

        runner := []*runners.Config{{
            Config: &runners.Run{
                Run:             []string{"for i in 1 2 3 4 5; do echo \"Run #${i}\"; sleep 1; done"},
                ContinueOnError: false,
            },
        }}

        err = pw.Add(p, runner, "")
        Expect(err).ToNot(HaveOccurred())

        stop, quit := pw.Watch()

        go func() {
            GinkgoRecover()
            err = os.Mkdir(filepath.Join(tmpDir, "new"), os.ModePerm)
            Expect(err).ToNot(HaveOccurred())
        }()

        err = os.Mkdir(filepath.Join(tmpDir, "another"), os.ModePerm)
        Expect(err).ToNot(HaveOccurred())

        time.Sleep(200 * time.Millisecond)

        stop()

        Eventually(quit, 15).Should(BeClosed())

        err = w.Close()
        Expect(err).ToNot(HaveOccurred())

        out, err := ioutil.ReadAll(r)
        Expect(err).ToNot(HaveOccurred())

        os.Stdout = stdout

        Expect(string(out)).To(ContainSubstring(fmt.Sprintf("Added: %s", tmpDir)))
        Expect(string(out)).To(ContainSubstring("Running"))
        Expect(strings.Count(string(out), "Running")).To(Equal(1))
    })

    It("returns an error if a path doesn't exist", func() {
        osStdout := os.Stdout
        osStderr := os.Stderr

        os.Stdout = nil
        os.Stderr = nil

        tmpDir, err := ioutil.TempDir("", "*")
        Expect(err).ToNot(HaveOccurred())

        pw, err := watchers.NewPathWatcher()
        Expect(err).ToNot(HaveOccurred())

        p := watchers.Path{
            Paths: []string{
                filepath.Join(tmpDir, "test"),
            },
            Recursive:  false,
            Exclusions: nil,
            Events: []string{
                "create",
                "write",
                "remove",
                "rename",
                "chmod",
            },
        }

        runner := []*runners.Config{{
            Config: &runners.Run{
                Run:             []string{"echo 'called'"},
                ContinueOnError: false,
            },
        }}

        err = pw.Add(p, runner, "")
        Expect(err).To(HaveOccurred())

        os.Stdout = osStdout
        os.Stderr = osStderr
    })

    It("returns an error if an unknown event is provided", func() {
        osStdout := os.Stdout
        osStderr := os.Stderr

        os.Stdout = nil
        os.Stderr = nil

        tmpDir, err := ioutil.TempDir("", "*")
        Expect(err).ToNot(HaveOccurred())

        pw, err := watchers.NewPathWatcher()
        Expect(err).ToNot(HaveOccurred())

        p := watchers.Path{
            Paths: []string{
                tmpDir,
            },
            Recursive:  true,
            Exclusions: nil,
            Events: []string{
                "unknown",
                "write",
                "remove",
                "rename",
                "chmod",
            },
        }

        err = pw.Add(p, nil, "")
        Expect(err).To(HaveOccurred())

        os.Stdout = osStdout
        os.Stderr = osStderr
    })
})
