package runners_test

import (
	"encoding/json"

	"github.com/iplay88keys/watchtower/pkg/runners"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	It("properly unmarshals run configs", func() {
		var runnerConfig runners.Config
		err := json.Unmarshal([]byte(`{"run": ["cmd", "cmd2"]}`), &runnerConfig)
		Expect(err).ToNot(HaveOccurred())
		Expect(runnerConfig).To(Equal(runners.Config{Config: &runners.Run{
			Run:             []string{"cmd", "cmd2"},
			ContinueOnError: false,
		}}))
	})

	It("properly unmarshals restart configs", func() {
		var runnerConfig runners.Config
		err := json.Unmarshal([]byte(`{"restart": "proc"}`), &runnerConfig)
		Expect(err).ToNot(HaveOccurred())
		Expect(runnerConfig).To(Equal(runners.Config{Config: &runners.Restart{
			Restart:    "proc",
			RunCleanup: false,
		}}))
	})

	It("returns an error if the config type is unknown", func() {
		var runnerConfig runners.Config
		err := json.Unmarshal([]byte(`{"unknown": "something"}`), &runnerConfig)
		Expect(err).To(HaveOccurred())
	})
})
