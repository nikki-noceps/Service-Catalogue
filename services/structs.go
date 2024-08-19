package services

import (
	"fmt"
	"nikki-noceps/serviceCatalogue/context"
	"nikki-noceps/serviceCatalogue/database"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
)

type (
	SortOrder string

	ServiceCatalogue struct {
		ServiceId   string `json:"serviceId,omitempty" mapstructure:"serviceId,omitempty"`
		Name        string `json:"name,omitempty" mapstructure:"name,omitempty"`
		Description string `json:"description,omitempty" mapstructure:"description,omitempty"`
		Version     int    `json:"version,omitempty" mapstructure:"version,omitempty"`
		CreatedAt   string `json:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
		UpdatedAt   string `json:"updatedAt,omitempty" mapstructure:"updatedAt,omitempty"`
		CreatedBy   string `json:"createdBy,omitempty" mapstructure:"createdBy,omitempty"`
		UpdatedBy   string `json:"updatedBy,omitempty" mapstructure:"updatedBy,omitempty"`
	}

	ServiceCatalogueVersion struct {
		ParentId        string `json:"parentId,omitempty" mapstructure:"parentId,omitempty"`
		VersionId       string `json:"versionId,omitempty" mapstructure:"versionId,omitempty"`
		Name            string `json:"name,omitempty" mapstructure:"name,omitempty"`
		Description     string `json:"description,omitempty" mapstructure:"description,omitempty"`
		Version         int    `json:"version,omitempty" mapstructure:"version,omitempty"`
		CreatedAt       string `json:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
		DecomissionedAt string `json:"decomissionedAt,omitempty" mapstructure:"decomissionedAt,omitempty"`
		CreatedBy       string `json:"createdBy,omitempty" mapstructure:"createdBy,omitempty"`
		DecomissionedBy string `json:"decomissionedBy,omitempty" mapstructure:"decomissionedBy,omitempty"`
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

	UpdateServiceCatalogueRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		UpdatedBy   string `json:"-"`
		ServiceId   string `json:"-"`
	}

	CreateServiceCatalogueResponse struct {
		*ServiceCatalogueResponse `json:"serviceCatalogue"`
		TimeStamp                 string `json:"timestamp"`
	}
	UpdateServiceCatalogueResponse struct {
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

	ServiceCatalogueVersionResponse struct {
		ParentId        string `json:"parentId,omitempty"`
		VersionId       string `json:"versionId,omitempty"`
		Name            string `json:"name,omitempty"`
		Description     string `json:"description,omitempty"`
		Version         int    `json:"version,omitempty"`
		CreatedAt       string `json:"createdAt,omitempty"`
		DecomissionedAt string `json:"decomissionedAt,omitempty"`
		CreatedBy       string `json:"createdBy,omitempty"`
		DecomissionedBy string `json:"decomissionedBy,omitempty"`
	}

	ListServiceCatalogueResponse struct {
		ServiceList []*ServiceCatalogueResponse `json:"serviceList"`
		TimeStamp   string                      `json:"timestamp"`
	}

	ListServiceCatalogueVersionsResponse struct {
		ServiceVersionsList []*ServiceCatalogueVersionResponse `json:"versions"`
		TimeStamp           string                             `json:"timestamp"`
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
		validation.Field(&c.CreatedBy, validation.Required.Error("missing `x-user-id` header")),
	)
}

func (c *UpdateServiceCatalogueRequest) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.When(c.Description == "", validation.Required)),
		validation.Field(&c.Description, validation.When(c.Name == "", validation.Required)),
		validation.Field(&c.UpdatedBy, validation.Required.Error("missing `x-user-id` header")),
		validation.Field(&c.ServiceId, validation.Required),
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

func (svcReq *UpdateServiceCatalogueRequest) RequestStructToServiceStruct(cctx context.CustomContext) *ServiceCatalogue {
	now := time.Now().UTC().Format(time.RFC3339)
	return &ServiceCatalogue{
		ServiceId:   svcReq.ServiceId,
		Name:        svcReq.Name,
		Description: svcReq.Description,
		UpdatedAt:   now,
		UpdatedBy:   svcReq.UpdatedBy,
	}
}
