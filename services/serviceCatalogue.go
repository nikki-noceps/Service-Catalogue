package services

import (
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
	"nikki-noceps/serviceCatalogue/logger/tag"

	"github.com/mitchellh/mapstructure"
)

// ListAllServicees queries the elasticsearch for hits and returns back serviceCatalogue objects
// Returns an error incase of any issues
func (svc *Service) ListAllService(cctx context.CustomContext, listParams *ListParameters) ([]*ServiceCatalogue, error) {
	query := &database.Query{}
	return nil, nil
}

func (svc *Service) FuzzySearchService(cctx context.CustomContext, searchText string) ([]*ServiceCatalogue, error) {
	query := &database.Query{
		MultiMatch: &database.MultiMatch{
			Fields:    []string{database.NameField, database.DescriptionField},
			Query:     searchText,
			Fuzziness: "AUTO",
		},
	}
	hits, err := svc.esClient.SearchAndGetHits(cctx, query, database.ServiceCatalogueIndex)
	if err != nil {
		cctx.Logger().ERROR("failed to search", tag.NewErrorTag(err))
		return nil, err
	}
	svcCatalogue := []*ServiceCatalogue{}
	for _, hit := range hits {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			cctx.Logger().ERROR("failed to parse hit", tag.NewAnyTag("hit", hit))
			continue
		}
		var svcCat ServiceCatalogue
		err := mapstructure.Decode(hitMap, &svcCat)
		if err != nil {
			cctx.Logger().ERROR("failed to decode hit", tag.NewErrorTag(err))
			continue
		}
		svcCatalogue = append(svcCatalogue, &svcCat)
	}
	return svcCatalogue, nil
}
