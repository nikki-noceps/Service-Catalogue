package services

import (
	"fmt"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
	"nikki-noceps/serviceCatalogue/logger/tag"

	"github.com/mitchellh/mapstructure"
)

// searchAndFetchServiceCatalogueList searches es with body provided.
// Parses the response and fetches _source from hits.hits and tranforms to service catalogue list
func (svc *Service) searchAndFetchServiceCatalogueList(cctx context.CustomContext, body *database.Body) ([]*ServiceCatalogue, error) {
	hits, err := svc.esClient.SearchAndGetHits(cctx, body, database.ServiceCatalogueIndex)
	if err != nil {
		cctx.Logger().DEBUG("failed to search", tag.NewErrorTag(err))
		return nil, err
	}
	svcCatalogue := []*ServiceCatalogue{}
	for _, hit := range hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			cctx.Logger().DEBUG("failed to parse hit", tag.NewAnyTag("hit", hit))
			continue
		}
		var svcCat ServiceCatalogue
		err := mapstructure.Decode(hitMap["_source"], &svcCat)
		if err != nil {
			cctx.Logger().DEBUG("failed to decode hit", tag.NewErrorTag(err))
			continue
		}
		svcCatalogue = append(svcCatalogue, &svcCat)
	}
	return svcCatalogue, nil
}

// searchAndFetchServiceCatalogue searches es with body provided.
// Parses the response and fetches _source from hits.hits and tranforms to service catalogue object
func (svc *Service) searchAndFetchServiceCatalogue(cctx context.CustomContext, body *database.Body) (map[string]any, error) {
	hits, err := svc.esClient.SearchAndGetHits(cctx, body, database.ServiceCatalogueIndex)
	if err != nil {
		cctx.Logger().DEBUG("failed to search", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	if len(hits) == 0 {
		return nil, NoDocumentFoundErr
	}
	hitmap, ok := hits[0].(map[string]any)
	if !ok {
		cctx.Logger().DEBUG("failed to parse hit", tag.NewAnyTag("hit", hits))
		return nil, fmt.Errorf("document parsing failed")
	}

	return hitmap, nil
}

// searchAndFetchServiceCatalogueVersions searches for all versions of a given serviceId
// Parses the response and fetches _source from hits.hits and tranforms to service catalogue versions list
func (svc *Service) searchAndFetchServiceCatalogueVersions(cctx context.CustomContext, body *database.Body) ([]*ServiceCatalogueVersion, error) {
	hits, err := svc.esClient.SearchAndGetHits(cctx, body, database.ServiceCatalogueVersionIndex)
	if err != nil {
		cctx.Logger().DEBUG("failed to search", tag.NewErrorTag(err))
		return nil, err
	}
	svcCatalogue := []*ServiceCatalogueVersion{}
	for _, hit := range hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			cctx.Logger().DEBUG("failed to parse hit", tag.NewAnyTag("hit", hit))
			continue
		}
		var svcCat ServiceCatalogueVersion
		err := mapstructure.Decode(hitMap["_source"], &svcCat)
		if err != nil {
			cctx.Logger().DEBUG("failed to decode hit", tag.NewErrorTag(err))
			continue
		}
		svcCatalogue = append(svcCatalogue, &svcCat)
	}
	return svcCatalogue, nil
}

// fetchServiceCatalogueVersion looks up a specific versionId provided in body
// Parses the response and fetches _source from hits.hits and tranforms to service catalogue versions list
func (svc *Service) fetchServiceCatalogueVersion(cctx context.CustomContext, body *database.Body) (*ServiceCatalogueVersion, error) {
	hits, err := svc.esClient.SearchAndGetHits(cctx, body, database.ServiceCatalogueVersionIndex)
	if err != nil {
		cctx.Logger().DEBUG("failed to search", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	if len(hits) == 0 {
		return nil, NoDocumentFoundErr
	}
	hitmap, ok := hits[0].(map[string]any)
	if !ok {
		cctx.Logger().DEBUG("failed to parse hit", tag.NewAnyTag("hit", hits))
		return nil, fmt.Errorf("document parsing failed")
	}

	svcCat := &ServiceCatalogueVersion{}
	err = mapstructure.Decode(hitmap["_source"], svcCat)
	if err != nil {
		cctx.Logger().DEBUG("failed to decode hit", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to decode hit: %w", err)
	}
	return svcCat, nil
}

func (svc *Service) deleteServiceCatalogue(cctx context.CustomContext, docId string) error {
	return svc.esClient.DeleteDocument(cctx, database.ServiceCatalogueIndex, docId)
}

// parses hits["_source"] received from es to service catalogue response
func (svc *Service) mapToServiceCatalogue(cctx context.CustomContext, hitmap map[string]any) (*ServiceCatalogue, error) {
	svcCat := &ServiceCatalogue{}
	err := mapstructure.Decode(hitmap["_source"], svcCat)
	if err != nil {
		cctx.Logger().DEBUG("failed to decode hit", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to decode hit: %w", err)
	}
	return svcCat, nil
}

func structToMap(v any) (map[string]any, error) {
	var result map[string]any
	err := mapstructure.Decode(v, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
