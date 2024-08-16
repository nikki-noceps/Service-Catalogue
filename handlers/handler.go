package handlers

import "nikki-noceps/serviceCatalogue/services"

type Handler struct {
	Svc *services.Service
}

func NewHandler(service *services.Service) *Handler {
	return &Handler{
		Svc: service,
	}
}
