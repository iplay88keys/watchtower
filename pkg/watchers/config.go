package watchers

import (
	"encoding/json"
	"errors"
	"fmt"
)

type WatcherConfig interface{}

type Config struct {
	Config WatcherConfig
}

func (c *Config) UnmarshalJSON(data []byte) error {
	watcherLookup := make(map[string]func() WatcherConfig)
	watcherLookup["paths"] = func() WatcherConfig { return &Path{} }

	var rawWatchConfig map[string]*json.RawMessage
	err := json.Unmarshal(data, &rawWatchConfig)
	if err != nil {
		return err
	}

	for key, newWatcher := range watcherLookup {
		_, found := rawWatchConfig[key]
		if !found {
			continue
		}

		watcher := newWatcher()
		err := json.Unmarshal(data, watcher)
		if err != nil {
			return err
		}

		c.Config = watcher
	}

	if c.Config == nil {
		return errors.New(fmt.Sprint("unknown watcher config:", string(data)))
	}

	return nil
}
