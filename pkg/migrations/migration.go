package migrations

import (
	"bytes"
	"context"
	"fmt"
	"nikki-noceps/serviceCatalogue/config"
	"nikki-noceps/serviceCatalogue/pkg/database"
	"nikki-noceps/serviceCatalogue/pkg/logger"
	"nikki-noceps/serviceCatalogue/pkg/logger/tag"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func RunMigrations(ctx context.Context, cfg *config.Configuration) error {
	esConfig := es.Config{
		Addresses: []string{fmt.Sprintf("%s:%s", cfg.ElasticSearch.Host, cfg.ElasticSearch.Port)},
		Username:  cfg.ElasticSearch.Username,
		Password:  cfg.ElasticSearch.Password,
	}

	esClient, err := es.NewClient(esConfig)
	if err != nil {
		return fmt.Errorf("failed to create new elasticsearch client: %w", err)
	}

	// Create servicecatalogue index
	err = createIndex(ctx, esClient, database.ServiceCatalogueIndex, serviceCatalogueMapping)
	if err != nil {
		logger.ERROR("failed to create index", tag.NewAnyTag("index", database.ServiceCatalogueIndex), tag.NewErrorTag(err))
	}

	// Create servicecatalogueversions index
	err = createIndex(ctx, esClient, database.ServiceCatalogueVersionIndex, serviceCatalogueVersionsMapping)
	if err != nil {
		logger.ERROR("failed to create index", tag.NewAnyTag("index", database.ServiceCatalogueVersionIndex), tag.NewErrorTag(err))
	}

	return nil
}

func createIndex(ctx context.Context, es *es.Client, index string, mapping []byte) error {
	req := esapi.IndicesCreateRequest{
		Index: index,
		Body:  bytes.NewReader(mapping),
	}
	res, err := req.Do(ctx, es)
	if err != nil {
		return fmt.Errorf("error creating or updating index: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error creating or updating index: %s", res.Status())
	}

	logger.INFO("Index Created !!", tag.NewAnyTag("index", index))
	return nil
}
