package database

import (
	"fmt"

	es "github.com/elastic/go-elasticsearch/v8"
)

type ESClient struct {
	client *es.Client
}

// Creates a new elasticsearch client. Tests the connection and returns it.
// Returns a wrapped error in case of any issues
func InitESClient(cfg es.Config) (*ESClient, error) {
	es, err := es.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new elasticsearch client: %w", err)
	}

	info, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer info.Body.Close()

	return &ESClient{
		client: es,
	}, nil
}
