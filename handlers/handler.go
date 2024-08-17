package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/logger/tag"
	"nikki-noceps/serviceCatalogue/services"
	"strconv"

	"github.com/gin-gonic/gin"
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
func (h *Handler) SearchSvcCatalogue(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())
	searchText := c.Query("search")
	if searchText == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("text cannot be empty"))
		return
	}
	resp, err := h.Svc.FuzzySearchService(cctx, searchText)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, resp)
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
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	listParams.AddDefaultsIfEmpty()
	if err := listParams.Validate(); err != nil {
		cctx.Logger().ERROR("VALIDATION_FAILED", tag.NewErrorTag(err))
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := h.Svc.ListAllService(cctx, listParams)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Parses the query parameters and generates struct object for list all services function
func getListQueryParams(queryParams url.Values, listParams *services.ListParameters) error {
	for key, values := range queryParams {
		if len(values) > 0 {
			var err error
			switch key {
			case "sort":
				sortField := []*services.SortField{}
				err = json.Unmarshal([]byte(values[0]), &sortField)
				listParams.Sort = sortField
			case "from":
				from := 0
				from, err = strconv.Atoi(values[0])
				listParams.From = from
			case "size":
				size := 10
				size, err = strconv.Atoi(values[0])
				listParams.Size = size
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
