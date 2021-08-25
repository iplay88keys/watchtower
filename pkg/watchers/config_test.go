package watchers_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/iplay88keys/watchtower/pkg/watchers"
)

var _ = Describe("Config", func() {
	It("properly unmarshals path watcher configs", func() {
		var watcherConfig watchers.Config
		err := json.Unmarshal([]byte(`{"paths": ["."]}`), &watcherConfig)
		Expect(err).ToNot(HaveOccurred())
		Expect(watcherConfig).To(Equal(watchers.Config{Config: &watchers.Path{
			Paths: []string{"."},
		}}))
	})

	It("returns an error if the config type is unknown", func() {
		var watcherConfig watchers.Config
		err := json.Unmarshal([]byte(`{"unknown": "something"}`), &watcherConfig)
		Expect(err).To(HaveOccurred())
	})
})
