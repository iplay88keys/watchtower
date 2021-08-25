package runners

import (
	"encoding/json"
	"errors"
	"fmt"
)

type RunnerConfig interface {
	Execute(triggeringFileName string) error
}

type Config struct {
	Config RunnerConfig
}

func (c *Config) UnmarshalJSON(data []byte) error {
	runnerLookup := make(map[string]func() RunnerConfig)
	runnerLookup["run"] = func() RunnerConfig { return &Run{} }
	runnerLookup["restart"] = func() RunnerConfig { return &Restart{} }

	var rawTrigger map[string]*json.RawMessage
	err := json.Unmarshal(data, &rawTrigger)
	if err != nil {
		return err
	}

	for key, newTrigger := range runnerLookup {
		_, found := rawTrigger[key]
		if !found {
			continue
		}

		trigger := newTrigger()
		err := json.Unmarshal(data, trigger)
		if err != nil {
			return err
		}

		c.Config = trigger
	}

	if c.Config == nil {
		return errors.New(fmt.Sprint("unknown runner config:", string(data)))
	}

	return nil
}
