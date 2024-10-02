package interfaces

import (
	"nikki-noceps/serviceCatalogue/pkg/context"
	"nikki-noceps/serviceCatalogue/pkg/database"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type CatalogueStore interface {
	SearchAndGetHits(cctx context.CustomContext, query *database.Body, index string) ([]any, error)
	CreateDocument(cctx context.CustomContext, docBytes []byte, index string) (*esapi.Response, error)
	UpdateDocument(cctx context.CustomContext, docBytes []byte, index string, docId string) (*esapi.Response, error)
	DeleteDocument(cctx context.CustomContext, index string, docId string) error
}
