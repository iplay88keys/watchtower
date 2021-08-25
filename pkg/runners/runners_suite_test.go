package runners_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRunners(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runners Suite")
}
