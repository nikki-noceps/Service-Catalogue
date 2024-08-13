package services

import (
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
)

type Service struct {
	esClient database.ESClient
}

func NewService(cctx context.CustomContext) {

}
