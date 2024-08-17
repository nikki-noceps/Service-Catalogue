package services

import (
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

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

	ListParameters struct {
		From           int          `json:"from"`
		Size           int          `json:"size"`
		Sort           []*SortField `json:"sort"`
		TimeStampField string       `json:"timestampfield"`
		TimeWindow     *TimeWindow  `json:"timewindow"`
	}

	SortField struct {
		Field string
		Value string
	}
	TimeWindow struct {
		Start string
		End   string
	}
)

func (l *ListParameters) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.From, validation.Required, validation.Min(0), validation.Max(1000)),
		validation.Field(&l.Size, validation.Required, validation.Max(50), validation.Min(10)),
		validation.Field(&l.Sort, validation.Required),
		validation.Field(&l.TimeStampField, validation.In("createdAt", "updatedAt")),
		validation.Field(&l.TimeWindow, validation.By(validateTimeDifference)),
	)
}

// Parses the struct and adds default values if empty
func (l *ListParameters) AddDefaultsIfEmpty() {
	if len(l.Sort) == 0 {
		l.Sort = append(l.Sort, &SortField{
			Field: "updatedAt",
			Value: "desc",
		})
	}
	if l.TimeWindow == nil {
		now := time.Now().UTC()
		l.TimeWindow = &TimeWindow{
			Start: now.Format(time.RFC3339),
			End:   now.AddDate(0, 0, -30).Format(time.RFC3339),
		}
	}
	if l.TimeStampField == "" {
		l.TimeStampField = "updatedAt"
	}
}

// Filter on time slice should not be greater than 30 days
func validateTimeDifference(times any) error {
	timeWindow, _ := times.(*TimeWindow)
	tStart, err := time.Parse(time.RFC3339, timeWindow.Start)
	if err != nil {
		return fmt.Errorf("start time in incorrect format: Valid format is RFC3339")
	}
	tEnd, err := time.Parse(time.RFC3339, timeWindow.End)
	if err != nil {
		return fmt.Errorf("end time in incorrect format: Valid format is RFC3339")
	}
	if tEnd.After(tStart) {
		return fmt.Errorf("end time must after start time")
	}
	return nil
}
