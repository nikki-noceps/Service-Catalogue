package services

import (
	"fmt"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

type (
	SortOrder string

	ServiceCatalogue struct {
		ServiceId   string `json:"serviceId" mapstructure:"serviceId"`
		Name        string `json:"name" mapstructure:"name"`
		Description string `json:"description" mapstructure:"description"`
		Version     int    `json:"version" mapstructure:"version"`
		CreatedAt   string `json:"createdAt" mapstructure:"createdAt"`
		UpdatedAt   string `json:"updatedAt" mapstructure:"updatedAt"`
		CreatedBy   string `json:"createdBy" mapstructure:"createdBy"`
		UpdatedBy   string `json:"updatedBy" mapstructure:"updatedBy"`
	}

	ServiceCatalogueVersions struct {
		ParentId    string `json:"parentId"`
		VersionId   string `json:"versionId"`
		Name        string `json:"name" mapstructure:"name"`
		Description string `json:"description" mapstructure:"description"`
		Version     int    `json:"version" mapstructure:"version"`
		CreatedAt   string `json:"createdAt" mapstructure:"createdAt"`
		UpdatedAt   string `json:"updatedAt" mapstructure:"updatedAt"`
		CreatedBy   string `json:"createdBy" mapstructure:"createdBy"`
		UpdatedBy   string `json:"updatedBy" mapstructure:"updatedBy"`
	}

	Query struct {
	}

	ListParameters struct {
		From           *int                  `json:"from"`
		Size           *int                  `json:"size"`
		Sort           []*database.SortField `json:"sort"`
		TimeStampField string                `json:"timestampfield"`
		TimeWindow     *TimeWindow           `json:"timewindow"`
	}

	SearchParameters struct {
		Search string `json:"search"`
		From   *int   `json:"from"`
		Size   *int   `json:"size"`
	}

	TimeWindow struct {
		Before string `json:"before"`
		After  string `json:"after"`
	}

	CreateServiceCatalogueRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedBy   string `json:"-"`
	}

	CreateServiceCatalogueResponse struct {
		*ServiceCatalogueResponse `json:"serviceCatalogue"`
		TimeStamp                 string `json:"timestamp"`
	}

	ListServiceCatalogueRequest struct {
	}

	ServiceCatalogueResponse struct {
		ServiceId   string `json:"serviceId"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Version     int    `json:"version"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
		CreatedBy   string `json:"createdBy"`
		UpdatedBy   string `json:"updatedBy"`
	}

	ListServiceCatalogueResponse struct {
		ServiceList []*ServiceCatalogueResponse `json:"serviceList"`
		TimeStamp   string                      `json:"timestamp"`
	}
)

func (l *ListParameters) Validate() error {
	return validation.ValidateStruct(l,
		validation.Field(&l.From, validation.NotNil, validation.Min(0), validation.Max(1000)),
		validation.Field(&l.Size, validation.NotNil, validation.Min(10), validation.Max(50)),
		validation.Field(&l.Sort, validation.Required),
		validation.Field(&l.TimeStampField, validation.Required, validation.In("createdAt", "updatedAt")),
		validation.Field(&l.TimeWindow, validation.Required, validation.By(validateTimeDifference)),
	)
}

func (s *SearchParameters) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.From, validation.NotNil, validation.Min(0), validation.Max(1000)),
		validation.Field(&s.Size, validation.NotNil, validation.Min(10), validation.Max(50)),
		validation.Field(&s.Search, validation.Required),
	)
}

// Parses the struct and adds default values if empty
func (l *ListParameters) AddDefaultsIfEmpty() {
	if len(l.Sort) == 0 {
		l.Sort = append(l.Sort, &database.SortField{
			"updatedAt": database.Desc,
		})
	}
	if l.TimeWindow == nil {
		now := time.Now().UTC()
		l.TimeWindow = &TimeWindow{
			Before: now.Format(time.RFC3339),
			After:  now.AddDate(0, 0, -30).Format(time.RFC3339),
		}
	}
	if l.TimeStampField == "" {
		l.TimeStampField = "updatedAt"
	}
}

// Filter on time slice should not be greater than 30 days
func validateTimeDifference(times any) error {
	timeWindow, _ := times.(*TimeWindow)
	tStart, err := time.Parse(time.RFC3339, timeWindow.Before)
	if err != nil {
		return fmt.Errorf("start time in incorrect format: Valid format is RFC3339")
	}
	tEnd, err := time.Parse(time.RFC3339, timeWindow.After)
	if err != nil {
		return fmt.Errorf("end time in incorrect format: Valid format is RFC3339")
	}
	if tEnd.After(tStart) {
		return fmt.Errorf("end time must after start time")
	}
	return nil
}

func (c *CreateServiceCatalogueRequest) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.Length(4, 20)),
		validation.Field(&c.Description, validation.Required, validation.Length(20, 200)),
		validation.Field(&c.CreatedBy, validation.Required.Error("missing `x-user-id` header ")),
	)
}

func (svcReq *CreateServiceCatalogueRequest) RequestStructToServiceStruct(cctx context.CustomContext) *ServiceCatalogue {
	now := time.Now().UTC().Format(time.RFC3339)
	return &ServiceCatalogue{
		ServiceId:   uuid.NewString(),
		Name:        svcReq.Name,
		Description: svcReq.Description,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   svcReq.CreatedBy,
		UpdatedBy:   svcReq.CreatedBy,
	}
}
