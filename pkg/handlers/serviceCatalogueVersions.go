package handlers

import (
	"net/http"
	"nikki-noceps/serviceCatalogue/internal/services"
	"nikki-noceps/serviceCatalogue/pkg/context"
	"nikki-noceps/serviceCatalogue/pkg/logger/tag"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListServiceCatalogueVersions(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	parentId := c.Param(keyServiceIdPathParam)

	resp, err := h.Svc.ListAllServiceVersions(cctx, parentId)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		c.Status(http.StatusInternalServerError)
		_ = c.Error(err)
		return
	}

	searchResponse := generateVersionListFromDBResponse(resp)

	c.JSON(http.StatusOK, searchResponse)
}

func (h *Handler) FetchServiceCatalogueVersionById(c *gin.Context) {
	cctx := context.CustomContextFromContext(c.Request.Context())

	versionId := c.Param(keyVersionIdPathParam)

	resp, err := h.Svc.FetchServiceCatalogueVersionById(cctx, versionId)
	if err != nil {
		cctx.Logger().ERROR("SERVICE_ERROR", tag.NewErrorTag(err))
		c.Status(http.StatusInternalServerError)
		_ = c.Error(err)
		return
	}

	searchResponse := &services.ServiceCatalogueVersionResponse{
		ParentId:        resp.ParentId,
		VersionId:       resp.VersionId,
		Name:            resp.Name,
		Description:     resp.Description,
		Version:         resp.Version,
		CreatedAt:       resp.CreatedAt,
		DecomissionedAt: resp.DecomissionedAt,
		CreatedBy:       resp.CreatedBy,
		DecomissionedBy: resp.DecomissionedBy,
	}

	c.JSON(http.StatusOK, searchResponse)
}
