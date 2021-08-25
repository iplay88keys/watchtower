package config_test

import (
	"io/ioutil"

	"github.com/iplay88keys/watchtower/pkg/config"
	"github.com/iplay88keys/watchtower/pkg/runners"
	"github.com/iplay88keys/watchtower/pkg/watchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	It("loads the config from a file", func() {
		f, err := ioutil.TempFile("", "config.yml")
		Expect(err).ToNot(HaveOccurred())

		_, err = f.WriteString(basicConfig)
		Expect(err).ToNot(HaveOccurred())

		cfg, err := config.Load(f.Name())
		Expect(err).ToNot(HaveOccurred())
		Expect(cfg).To(Equal(&config.Config{
			Watches: []config.Watch{{
				Name: "test",
				Config: watchers.Config{
					Config: &watchers.Path{
						Paths:      []string{"test.yml"},
						Recursive:  true,
						Exclusions: []string{"exc/*"},
						Events:     []string{"create", "write"},
					},
				},
				OnTrigger: []runners.Config{{
					Config: &runners.Run{
						Run:             []string{"pwd"},
						ContinueOnError: true,
					},
				}, {
					Config: &runners.Restart{
						Restart:    "list",
						RunCleanup: false,
					},
				}},
			}},
			Processes: []runners.Process{{
				Name:       "list",
				Type:       "task",
				StartCmd:   "ls",
				StopCmd:    "",
				RestartCmd: "",
				CleanupCmd: "",
			}, {
				Name:       "echo",
				Type:       "task",
				StartCmd:   "echo 'hello'",
				StopCmd:    "",
				RestartCmd: "",
				CleanupCmd: "",
			}},
		}))
	})

	It("returns an error if the file doesn't exist", func() {
		_, err := config.Load("non-existent.yml")
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if the file has invalid yaml", func() {
		f, err := ioutil.TempFile("", "invalidConfig.yml")
		Expect(err).ToNot(HaveOccurred())

		_, err = f.WriteString(invalidConfig)
		Expect(err).ToNot(HaveOccurred())

		_, err = config.Load(f.Name())
		Expect(err).To(HaveOccurred())
	})
})

const basicConfig = `
watches:
  - name: "test"
    config:
      paths:
        - "test.yml"
      recursive: true
      exclusions:
        - "exc/*"
      events:
        - "create"
        - "write"
    onTrigger:
      - run:
        - "pwd"
        continueOnError: true
      - restart: "list"
processes:
  - name: "list"
    type: "task"
    start: "ls"
  - name: "echo"
    type: "task"
    start: "echo 'hello'"
`

const invalidConfig = `:-`
