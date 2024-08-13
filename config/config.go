package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type (
	Configuration struct {
		App           App           `yaml:"App"`
		ElasticSearch ElasticSearch `yaml:"Database"`
	}

	App struct {
		LogLevel    string `yaml:"LogLevel"`
		Environment string `yaml:"Environment"`
	}

	ElasticSearch struct {
		Host     string `yaml:"Host"`
		Port     string `yaml:"Port"`
		Username string `yaml:"Username"`
		Password string `yaml:"Password"`
	}
)

// Reads yaml file specified in location, parses the config and returns a Configuration object
// returns error will nil in case of any issues
func Load(location string) (*Configuration, error) {
	data, err := os.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	config := &Configuration{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return config, nil
}
