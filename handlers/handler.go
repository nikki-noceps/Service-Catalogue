package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
	"nikki-noceps/serviceCatalogue/logger/tag"
	"nikki-noceps/serviceCatalogue/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	userRequestHeader     = "x-user-id"
	keyServiceIdPathParam = "serviceId"
	keyVersionIdPathParam = "versionId"
)

type Handler struct {
	Svc *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{
		Svc: service,
	}
}

// SearchSvcCatalogue handler parses the request query to do a fuzzy search on name and description field on elasticsearch.
// Returns list of serviceCatalogues which match the search and error in case of an issues.
// Sets default from 0 and size as 20 if not specified
func (h *Handler) SearchSvcCatalogue(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	queryParams := c.Request.URL.Query()
	searchParams := &services.SearchParameters{}
	if err := getSearchQueryParams(queryParams, searchParams); err != nil {
		cctx.Logger().ERROR("QUERY_PARSING_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	if err := searchParams.Validate(); err != nil {
		cctx.Logger().ERROR("VALIDATION_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	resp, err := h.Svc.FuzzySearchService(cctx, searchParams)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(err)
		return
	}

	searchResponse := generateListResponseFromDBResponse(resp)

	c.JSON(http.StatusOK, searchResponse)
}

// ListSvcCatalogue handler parses request query, validates the request query and handles pagination and sorting
// of serviceCatalogue search request to elasticsearch
// Returns list of serviceCatalogues and error in case of an issues.
func (h *Handler) ListSvcCatalogue(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	queryParams := c.Request.URL.Query()
	listParams := &services.ListParameters{}
	if err := getListQueryParams(queryParams, listParams); err != nil {
		cctx.Logger().ERROR("QUERY_PARSING_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}
	listParams.AddDefaultsIfEmpty()
	if err := listParams.Validate(); err != nil {
		cctx.Logger().ERROR("VALIDATION_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	resp, err := h.Svc.ListAllServices(cctx, listParams)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		c.Status(http.StatusInternalServerError)
		_ = c.Error(err)
		return
	}

	searchResponse := generateListResponseFromDBResponse(resp)

	c.JSON(http.StatusOK, searchResponse)
}

// CreateSvcCatalogue creates a service document in servicecatalogue index
func (h *Handler) CreateSvcCatalogue(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	serviceCatalogueReq := &services.CreateServiceCatalogueRequest{}
	err := c.ShouldBindJSON(serviceCatalogueReq)
	if err != nil {
		cctx.Logger().ERROR("REQUEST_PARSING_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	serviceCatalogueReq.CreatedBy = c.GetHeader(userRequestHeader)

	if err := serviceCatalogueReq.Validate(); err != nil {
		cctx.Logger().ERROR("VALIDATION_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	svcCatalogue := serviceCatalogueReq.RequestStructToServiceStruct(cctx)

	resp, err := h.Svc.CreateServiceCatalogue(cctx, svcCatalogue)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	serviceCatalogueResp := &services.CreateServiceCatalogueResponse{
		ServiceCatalogueResponse: &services.ServiceCatalogueResponse{
			ServiceId:   resp.ServiceId,
			Name:        resp.Name,
			Description: resp.Description,
			Version:     resp.Version,
			CreatedAt:   resp.CreatedAt,
			UpdatedAt:   resp.UpdatedAt,
			CreatedBy:   resp.CreatedBy,
			UpdatedBy:   resp.UpdatedBy,
		},
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, serviceCatalogueResp)
}

// UpdateSvcCatalogue updates the elasticsearch document and stores the older document as a version in another
// index called servicecatalogueversion
func (h *Handler) UpdateSvcCatalogue(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	serviceId := c.Param(keyServiceIdPathParam)

	serviceCatalogueReq := &services.UpdateServiceCatalogueRequest{
		ServiceId: serviceId,
	}
	err := c.ShouldBindJSON(serviceCatalogueReq)
	if err != nil {
		cctx.Logger().ERROR("REQUEST_PARSING_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	serviceCatalogueReq.UpdatedBy = c.GetHeader(userRequestHeader)

	if err := serviceCatalogueReq.Validate(); err != nil {
		cctx.Logger().ERROR("VALIDATION_FAILED", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	svcCatalogue := serviceCatalogueReq.RequestStructToServiceStruct(cctx)
	resp, err := h.Svc.UpdateServiceCatalogue(cctx, svcCatalogue)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		c.Status(http.StatusBadRequest)
		_ = c.Error(err)
		return
	}

	serviceCatalogueResp := &services.UpdateServiceCatalogueResponse{
		ServiceCatalogueResponse: &services.ServiceCatalogueResponse{
			ServiceId:   resp.ServiceId,
			Name:        resp.Name,
			Description: resp.Description,
			Version:     resp.Version,
			CreatedAt:   resp.CreatedAt,
			UpdatedAt:   resp.UpdatedAt,
			CreatedBy:   resp.CreatedBy,
			UpdatedBy:   serviceCatalogueReq.UpdatedBy,
		},
		TimeStamp: time.Now().Format(time.RFC3339),
	}
	if serviceCatalogueReq.Name != "" {
		serviceCatalogueResp.Name = serviceCatalogueReq.Name
	}
	if serviceCatalogueReq.Description != "" {
		serviceCatalogueResp.Description = serviceCatalogueReq.Description
	}
	c.JSON(http.StatusOK, serviceCatalogueResp)
}

// DeleteService deletes document from servicecatalogue index but also stores a copy in the servicecatalogueversion index
func (h *Handler) DeleteService(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	serviceId := c.Param(keyServiceIdPathParam)
	userId := c.GetHeader(userRequestHeader)

	err := h.Svc.DeleteService(cctx, serviceId, userId)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		if errors.Is(err, services.NoDocumentFoundErr) {
			c.Status(http.StatusBadRequest)
			_ = c.Error(err)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusAccepted, nil)
}

// FetchServiceById fetches a single document from servicecatalogue index
func (h *Handler) FetchServiceById(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	serviceId := c.Param(keyServiceIdPathParam)

	resp, err := h.Svc.FetchServiceById(cctx, serviceId)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		if errors.Is(err, services.NoDocumentFoundErr) {
			c.Status(http.StatusBadRequest)
			_ = c.Error(err)
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(err)
		return
	}
	serviceCatalogueResp := &services.ServiceCatalogueResponse{
		ServiceId:   resp.ServiceId,
		Name:        resp.Name,
		Description: resp.Description,
		Version:     resp.Version,
		CreatedAt:   resp.CreatedAt,
		UpdatedAt:   resp.UpdatedAt,
		CreatedBy:   resp.CreatedBy,
		UpdatedBy:   resp.UpdatedBy,
	}
	c.JSON(http.StatusOK, serviceCatalogueResp)
}

// Parses the query parameters and generates struct object for list all services function
func getListQueryParams(queryParams url.Values, listParams *services.ListParameters) error {
	for key, values := range queryParams {
		if len(values) > 0 {
			var err error
			switch key {
			case "sort":
				sortField := []*database.SortField{}
				err = json.Unmarshal([]byte(values[0]), &sortField)
				listParams.Sort = sortField
			case "from":
				from := 0
				from, err = strconv.Atoi(values[0])
				listParams.From = &from
			case "size":
				size := 10
				size, err = strconv.Atoi(values[0])
				listParams.Size = &size
			case "timewindow":
				timeWindow := &services.TimeWindow{}
				err = json.Unmarshal([]byte(values[0]), timeWindow)
				listParams.TimeWindow = timeWindow
			case "timestampfield":
				listParams.TimeStampField = values[0]
			}

			if err != nil {
				return fmt.Errorf("invalid query params: %w", err)
			}
		}
	}
	return nil
}

func getSearchQueryParams(queryParams url.Values, searchParams *services.SearchParameters) error {
	for key, values := range queryParams {
		if len(values) > 0 {
			var err error
			switch key {
			case "search":
				searchParams.Search = values[0]
			case "from":
				from := 0
				from, err = strconv.Atoi(values[0])
				searchParams.From = &from
			case "size":
				size := 10
				size, err = strconv.Atoi(values[0])
				searchParams.Size = &size
			}

			if err != nil {
				return fmt.Errorf("invalid query params: %w", err)
			}
		}
	}
	return nil
}

func generateListResponseFromDBResponse(resp []*services.ServiceCatalogue) *services.ListServiceCatalogueResponse {
	serviceList := []*services.ServiceCatalogueResponse{}
	for _, serviceCatalogue := range resp {
		serviceList = append(serviceList, &services.ServiceCatalogueResponse{
			ServiceId:   serviceCatalogue.ServiceId,
			Name:        serviceCatalogue.Name,
			Description: serviceCatalogue.Description,
			Version:     serviceCatalogue.Version,
			CreatedAt:   serviceCatalogue.CreatedAt,
			UpdatedAt:   serviceCatalogue.UpdatedAt,
			CreatedBy:   serviceCatalogue.CreatedBy,
			UpdatedBy:   serviceCatalogue.UpdatedBy,
		})
	}
	searchRespones := &services.ListServiceCatalogueResponse{
		ServiceList: serviceList,
		TimeStamp:   time.Now().UTC().Format(time.RFC3339),
	}
	return searchRespones
}

func generateVersionListFromDBResponse(resp []*services.ServiceCatalogueVersion) *services.ListServiceCatalogueVersionsResponse {
	versionsList := []*services.ServiceCatalogueVersionResponse{}
	for _, serviceCatVersion := range resp {
		versionsList = append(versionsList, &services.ServiceCatalogueVersionResponse{
			ParentId:        serviceCatVersion.ParentId,
			VersionId:       serviceCatVersion.VersionId,
			Name:            serviceCatVersion.Name,
			Description:     serviceCatVersion.Description,
			Version:         serviceCatVersion.Version,
			CreatedAt:       serviceCatVersion.CreatedAt,
			DecomissionedAt: serviceCatVersion.DecomissionedAt,
			CreatedBy:       serviceCatVersion.CreatedBy,
			DecomissionedBy: serviceCatVersion.DecomissionedBy,
		})
	}
	searchResponse := &services.ListServiceCatalogueVersionsResponse{
		ServiceVersionsList: versionsList,
		TimeStamp:           time.Now().UTC().Format(time.RFC3339),
	}
	return searchResponse
}
