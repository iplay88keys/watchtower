package config

import (
	"encoding/json"
	"os"

	"sigs.k8s.io/yaml"

	"github.com/iplay88keys/watchtower/pkg/runners"
	"github.com/iplay88keys/watchtower/pkg/watchers"
)

type Config struct {
	Watches   []Watch           `json:"watches"`
	Processes []runners.Process `json:"processes"`
}

type Watch struct {
	Name      string           `json:"name"`
	Config    watchers.Config  `json:"config"`
	OnTrigger []runners.Config `json:"onTrigger"`
}

func Load(path string) (*Config, error) {
	var cfg Config

	yamlConfig, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jsonConfig, err := yaml.YAMLToJSON(yamlConfig)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonConfig, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
