package services

import (
	"context"
	"nikki-noceps/serviceCatalogue/internal/interfaces"
)

type Service struct {
	esClient interfaces.CatalogueStore
}

func NewService(ctx context.Context, client interfaces.CatalogueStore) (*Service, error) {
	return &Service{
		esClient: client,
	}, nil
}
