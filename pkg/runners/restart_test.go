package runners_test

import (
	"os"

	"github.com/iplay88keys/watchtower/pkg/runners"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restart", func() {
	var (
		osStdout *os.File
		osStderr *os.File
	)

	BeforeEach(func() {
		osStdout = os.Stdout
		osStderr = os.Stderr

		os.Stdout = nil
		os.Stderr = nil
	})

	AfterEach(func() {
		os.Stdout = osStdout
		os.Stderr = osStderr
	})

	It("restarts a process", func() {
		proc := restartMock{}

		restartRunner := runners.Restart{}
		restartRunner.Setup(&proc)
		err := restartRunner.Execute("")
		Expect(err).ToNot(HaveOccurred())

		Expect(proc.called).To(BeTrue())
	})
})

type restartMock struct {
	called bool
}

func (r *restartMock) Restart(runCleanup bool) error {
	r.called = true

	return nil
}
