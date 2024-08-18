package services

import (
	"encoding/json"
	"fmt"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
	"nikki-noceps/serviceCatalogue/logger/tag"
	"time"

	"github.com/mitchellh/mapstructure"
)

var keyServiceId = "serviceId.keyword"
var NoDocumentFoundErr error = fmt.Errorf("DOCUMENT_NOT_FOUND")

// ListAllServicees queries the elasticsearch for hits and returns back serviceCatalogue objects
// Returns an error incase of any issues
func (svc *Service) ListAllService(cctx context.CustomContext, listParams *ListParameters) ([]*ServiceCatalogue, error) {
	body := &database.Body{
		Query: &database.Query{
			Range: &database.RangeQuery{
				listParams.TimeStampField: {
					Gte: listParams.TimeWindow.After,
					Lte: listParams.TimeWindow.Before,
				},
			},
		},
		Sort: listParams.Sort,
		From: *listParams.From,
		Size: *listParams.Size,
	}

	return svc.searchAndFetchServiceCatalogueList(cctx, body)
}

func (svc *Service) FetchServiceById(cctx context.CustomContext, serviceId string) (*ServiceCatalogue, error) {
	body := &database.Body{
		Query: &database.Query{
			Term: &database.TermQuery{
				keyServiceId: {
					Value: serviceId,
				},
			},
		},
	}

	return svc.searchAndFetchServiceCatalogue(cctx, body)
}

func (svc *Service) FuzzySearchService(cctx context.CustomContext, searchParams *SearchParameters) ([]*ServiceCatalogue, error) {
	body := &database.Body{
		Query: &database.Query{
			MultiMatch: &database.MultiMatch{
				Fields:    []string{database.NameField, database.DescriptionField},
				Query:     searchParams.Search,
				Fuzziness: "AUTO",
			},
		},
		From: *searchParams.From,
		Size: *searchParams.Size,
	}

	return svc.searchAndFetchServiceCatalogueList(cctx, body)
}

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
func (svc *Service) searchAndFetchServiceCatalogue(cctx context.CustomContext, body *database.Body) (*ServiceCatalogue, error) {
	hits, err := svc.esClient.SearchAndGetHits(cctx, body, database.ServiceCatalogueIndex)
	if err != nil {
		cctx.Logger().DEBUG("failed to search", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	if len(hits) == 0 {
		return nil, NoDocumentFoundErr
	}
	svcCat := &ServiceCatalogue{}
	hitmap, ok := hits[0].(map[string]any)
	if !ok {
		cctx.Logger().DEBUG("failed to parse hit", tag.NewAnyTag("hit", hits))
		return nil, fmt.Errorf("document parsing failed")
	}
	err = mapstructure.Decode(hitmap["_source"], svcCat)
	if err != nil {
		cctx.Logger().DEBUG("failed to decode hit", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to decode hit: %w", err)
	}
	return svcCat, nil
}

func (svc *Service) CreateServiceCatalogue(cctx context.CustomContext, input *ServiceCatalogue) (*ServiceCatalogue, error) {
	input.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	input.UpdatedAt = input.CreatedAt
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}
	_, err = svc.esClient.CreateDocument(cctx, inputBytes, database.ServiceCatalogueIndex)
	if err != nil {
		return nil, err
	}

	return input, nil
}

func (svc *Service) UpdateServiceCatalogue(cctx context.CustomContext) (*ServiceCatalogue, error) {
	return nil, nil
}
