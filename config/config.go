package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Configuration struct {
		App           App           `yaml:"App"`
		Server        Server        `yaml:"Server"`
		ElasticSearch ElasticSearch `yaml:"Database"`
	}

	Server struct {
		Port string `yaml:"Port"`
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
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	addDefaults(config)
	return config, nil
}

func addDefaults(config *Configuration) {
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.ElasticSearch.Host == "" {
		config.ElasticSearch.Host = "localhost"
	}
	if config.ElasticSearch.Port == "" {
		config.ElasticSearch.Port = "9200"
	}
}
