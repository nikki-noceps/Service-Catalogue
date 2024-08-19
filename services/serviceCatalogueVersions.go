package services

import (
	"encoding/json"
	"fmt"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
	"time"
)

func (svc *Service) CreateServiceCatalogueVersion(cctx context.CustomContext, input *ServiceCatalogueVersion) (*ServiceCatalogueVersion, error) {
	input.DecomissionedAt = time.Now().UTC().Format(time.RFC3339)
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}
	_, err = svc.esClient.CreateDocument(cctx, inputBytes, database.ServiceCatalogueVersionIndex)
	if err != nil {
		return nil, err
	}

	return input, nil
}

func (svc *Service) ListAllServiceVersions(cctx context.CustomContext, parentId string) ([]*ServiceCatalogueVersion, error) {
	body := &database.Body{
		Query: &database.Query{
			Term: &database.TermQuery{
				keyParentId: {
					Value: parentId,
				},
			},
		},
	}

	return svc.searchAndFetchServiceCatalogueVersions(cctx, body)
}

func (svc *Service) FetchServiceCatalogueVersionById(cctx context.CustomContext, versionId string) (*ServiceCatalogueVersion, error) {
	body := &database.Body{
		Query: &database.Query{
			Term: &database.TermQuery{
				keyVersionId: {
					Value: versionId,
				},
			},
		},
	}

	return svc.fetchServiceCatalogueVersion(cctx, body)
}
