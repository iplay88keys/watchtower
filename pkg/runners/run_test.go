package runners_test

import (
	"io/ioutil"
	"os"

	"github.com/iplay88keys/watchtower/pkg/runners"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Run", func() {
	It("runs the specified processes", func() {
		stdout := os.Stdout
		r, w, err := os.Pipe()
		Expect(err).ToNot(HaveOccurred())
		os.Stdout = w

		runner := runners.Run{
			Run: []string{
				"echo 'first'",
				"echo 'second'",
			},
			ContinueOnError: false,
		}
		err = runner.Execute("")
		Expect(err).ToNot(HaveOccurred())

		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		out, err := ioutil.ReadAll(r)
		Expect(err).ToNot(HaveOccurred())

		os.Stdout = stdout

		Eventually(string(out)).Should(Equal("Running: 'echo 'first''\nfirst\n\nRunning: 'echo 'second''\nsecond\n\n"))
	})

	It("replaces template variables with their values", func() {
		stdout := os.Stdout
		r, w, err := os.Pipe()
		Expect(err).ToNot(HaveOccurred())
		os.Stdout = w

		runner := runners.Run{
			Run: []string{
				"echo '{{.Name}}'",
			},
			ContinueOnError: false,
		}
		err = runner.Execute("test")
		Expect(err).ToNot(HaveOccurred())

		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		out, err := ioutil.ReadAll(r)
		Expect(err).ToNot(HaveOccurred())

		os.Stdout = stdout

		Eventually(string(out)).Should(Equal("Running: 'echo 'test''\ntest\n\n"))
	})

	It("continues running processes even on failure if continue is true", func() {
		stdout := os.Stdout
		r, w, err := os.Pipe()
		Expect(err).ToNot(HaveOccurred())
		os.Stdout = w

		runner := runners.Run{
			Run: []string{
				"nonexistent-command",
				"echo 'ran'",
			},
			ContinueOnError: true,
		}
		err = runner.Execute("")
		Expect(err).ToNot(HaveOccurred())

		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		out, err := ioutil.ReadAll(r)
		Expect(err).ToNot(HaveOccurred())

		os.Stdout = stdout

		Eventually(string(out)).Should(Equal("Running: 'nonexistent-command'\nbash: nonexistent-command: command not found\nexit status 127\nRunning: 'echo 'ran''\nran\n\n"))
	})

	It("stops running processes on failure if continue is false", func() {
		stdout := os.Stdout
		r, w, err := os.Pipe()
		Expect(err).ToNot(HaveOccurred())
		os.Stdout = w

		runner := runners.Run{
			Run: []string{
				"nonexistent-command",
				"echo 'stop'",
			},
			ContinueOnError: false,
		}
		err = runner.Execute("")
		Expect(err).To(HaveOccurred())

		err = w.Close()
		Expect(err).ToNot(HaveOccurred())

		out, err := ioutil.ReadAll(r)
		Expect(err).ToNot(HaveOccurred())

		os.Stdout = stdout

		Eventually(string(out)).Should(Equal("Running: 'nonexistent-command'\nbash: nonexistent-command: command not found\n"))
	})
})
