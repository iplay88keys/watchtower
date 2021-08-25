package runners

import (
	"fmt"
	"strings"
)

type Run struct {
	Run             []string `yaml:"run"`
	ContinueOnError bool     `yaml:"continueOnError"`
}

var NameTemplate = "{{.Name}}"

func (r *Run) Execute(triggeringFileName string) error {
	for _, command := range r.Run {
		if strings.Contains(command, NameTemplate) {
			command = strings.ReplaceAll(command, NameTemplate, triggeringFileName)
		}

		proc := Process{
			Type:     "task",
			StartCmd: command,
		}

		err := proc.Start()
		if err != nil {
			if !r.ContinueOnError {
				return err
			}

			fmt.Println(err.Error())
		}
	}

	return nil
}
