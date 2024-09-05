package services

import (
	"context"
	"nikki-noceps/serviceCatalogue/database"
)

type Service struct {
	esClient database.ESClient
}

func NewService(ctx context.Context, esClient *database.ESClient) (*Service, error) {
	return &Service{
		esClient: *esClient,
	}, nil
}
