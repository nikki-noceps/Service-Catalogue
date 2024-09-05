package services

import (
	"context"
	"fmt"
	"nikki-noceps/serviceCatalogue/config"
	"nikki-noceps/serviceCatalogue/pkg/database"
)

type Service struct {
	esClient database.ESClient
}

func NewService(ctx context.Context, cfg *config.Configuration) (*Service, error) {
	esClient, err := database.InitESClient(cfg.ElasticSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to setup database: %w", err)
	}

	return &Service{
		esClient: *esClient,
	}, nil
}
