package services

import (
	"encoding/json"
	"fmt"
	"nikki-noceps/serviceCatalogue/pkg/context"
	"nikki-noceps/serviceCatalogue/pkg/database"
	"nikki-noceps/serviceCatalogue/pkg/logger/tag"

	"github.com/google/uuid"
)

var keyServiceId = "serviceId.keyword"
var keyParentId = "parentId.keyword"
var keyVersionId = "versionId.keyword"
var NoDocumentFoundErr error = fmt.Errorf("DOCUMENT_NOT_FOUND")

// ListAllServicees queries the elasticsearch for hits and returns back serviceCatalogue objects
// Returns an error incase of any issues
func (svc *Service) ListAllServices(cctx context.CustomContext, listParams *ListParameters) ([]*ServiceCatalogue, error) {
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

	hitMap, err := svc.searchAndFetchServiceCatalogue(cctx, body)
	if err != nil {
		return nil, err
	}
	return svc.mapToServiceCatalogue(cctx, hitMap)
}

func (svc *Service) DeleteService(cctx context.CustomContext, serviceId string, userId string) error {
	body := &database.Body{
		Query: &database.Query{
			Term: &database.TermQuery{
				keyServiceId: {
					Value: serviceId,
				},
			},
		},
	}

	hitMap, err := svc.searchAndFetchServiceCatalogue(cctx, body)
	if err != nil {
		return err
	}

	svcCat, err := svc.mapToServiceCatalogue(cctx, hitMap)
	if err != nil {
		return err
	}

	svcCatVersion := &ServiceCatalogueVersion{
		ParentId:        svcCat.ServiceId,
		VersionId:       uuid.NewString(),
		Name:            svcCat.Name,
		Description:     svcCat.Description,
		Version:         svcCat.Version,
		CreatedAt:       svcCat.UpdatedAt,
		CreatedBy:       svcCat.CreatedBy,
		DecomissionedBy: userId,
	}

	_, err = svc.CreateServiceCatalogueVersion(cctx, svcCatVersion)
	if err != nil {
		return err
	}

	err = svc.deleteServiceCatalogue(cctx, hitMap["_id"].(string))
	return err
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

func (svc *Service) CreateServiceCatalogue(cctx context.CustomContext, input *ServiceCatalogue) (*ServiceCatalogue, error) {
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

// UpdateServiceCatalogue: updates fields in the service catalogue and creates new version in the service catalogue versions index
// corresponding to the service being updated. Ideally we should use database transactions to ensure both the steps are collectively
// atomic.
func (svc *Service) UpdateServiceCatalogue(cctx context.CustomContext, input *ServiceCatalogue) (*ServiceCatalogue, error) {
	body := &database.Body{
		Query: &database.Query{
			Term: &database.TermQuery{
				keyServiceId: {
					Value: input.ServiceId,
				},
			},
		},
	}

	hitMap, err := svc.searchAndFetchServiceCatalogue(cctx, body)
	if err != nil {
		return nil, err
	}

	svcCat, err := svc.mapToServiceCatalogue(cctx, hitMap)
	if err != nil {
		return nil, err
	}

	// create version
	svcVersion := &ServiceCatalogueVersion{
		ParentId:        svcCat.ServiceId,
		VersionId:       uuid.NewString(),
		Name:            svcCat.Name,
		Description:     svcCat.Description,
		Version:         svcCat.Version,
		CreatedAt:       svcCat.UpdatedAt,
		CreatedBy:       svcCat.UpdatedBy,
		DecomissionedBy: input.UpdatedBy,
	}

	svcVersionResp, err := svc.CreateServiceCatalogueVersion(cctx, svcVersion)
	if err != nil {
		return nil, err
	}

	cctx.Logger().DEBUG("version created", tag.NewAnyTag("svc_version", svcVersionResp))

	// update document
	id := hitMap["_id"].(string)
	// increment version
	svcCat.Version++
	input.Version = svcCat.Version

	updateMap, err := structToMap(input)
	if err != nil {
		cctx.Logger().DEBUG("error struct to map", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to generate update doc: %w", err)
	}
	updateDoc := &database.UpdateBody{
		Doc: updateMap,
	}
	updateBytes, err := json.Marshal(updateDoc)
	if err != nil {
		cctx.Logger().DEBUG("failed to convert struct to map", tag.NewErrorTag(err))
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	_, err = svc.esClient.UpdateDocument(cctx, updateBytes, database.ServiceCatalogueIndex, id)
	if err != nil {
		return nil, err
	}
	return svcCat, nil
}
