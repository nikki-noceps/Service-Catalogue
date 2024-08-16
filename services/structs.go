package services

import "time"

type (
	ServiceCatalogue struct {
		DocumentId string `json:"documentId"`
		CoreService
	}

	ServiceCatalogueVersions struct {
		ParentId  string `json:"parentId"`
		VersionId string `json:"versionId"`
		CoreService
	}

	CoreService struct {
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Version     string    `json:"version"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
		CreatedBy   string    `json:"createdBy"`
		UpdatedBy   string    `json:"updatedBy"`
	}

	Query struct {
	}
)
