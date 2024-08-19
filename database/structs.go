package database

var (
	NameField                    = "name"
	DescriptionField             = "description"
	ServiceCatalogueIndex        = "servicecatalogue"
	ServiceCatalogueVersionIndex = "servicecatalogueversions"
)

var (
	Asc  SortOrder = "asc"
	Desc SortOrder = "desc"
)

type (
	// SortOrder a custom type for sort in elasticsearch
	SortOrder string

	// Body Represents the body to be sent to elasticsearch api request
	Body struct {
		Query *Query       `json:"query,omitempty"`
		Sort  []*SortField `json:"sort,omitempty"`
		From  int          `json:"from,omitempty"`
		Size  int          `json:"size,omitempty"`
	}

	// Query represents a generic Elasticsearch query structure
	Query struct {
		Match         *MatchQuery         `json:"match,omitempty"`
		MultiMatch    *MultiMatch         `json:"multi_match,omitempty"`
		Term          *TermQuery          `json:"term,omitempty"`
		Range         *RangeQuery         `json:"range,omitempty"`
		Bool          *BoolQuery          `json:"bool,omitempty"`
		FunctionScore *FunctionScoreQuery `json:"function_score,omitempty"`
		Script        *ScriptQuery        `json:"script,omitempty"`
	}

	// Represents a match query
	MatchQuery struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}

	// Represents a multimatch query
	MultiMatch struct {
		Fields    []string `json:"fields"`
		Query     string   `json:"query"`
		Fuzziness string   `json:"fuzziness"`
	}

	// TermQuery represents a term query
	TermQuery map[string]*FieldValue

	FieldValue struct {
		Value any `json:"value"`
	}

	// RangeQuery represents a range query
	RangeQuery map[string]struct {
		Gte any `json:"gte,omitempty"`
		Lte any `json:"lte,omitempty"`
	}

	// BoolQuery represents a boolean query
	BoolQuery struct {
		Must    []Query `json:"must,omitempty"`
		Should  []Query `json:"should,omitempty"`
		MustNot []Query `json:"must_not,omitempty"`
	}

	// FunctionScoreQuery represents a function score query
	FunctionScoreQuery struct {
		Query     Query      `json:"query"`
		Functions []Function `json:"functions,omitempty"`
	}

	// Function represents a function for function score queries
	Function struct {
		Filter Query   `json:"filter"`
		Weight float64 `json:"weight"`
	}

	// ScriptQuery represents a script-based query
	ScriptQuery struct {
		Script string `json:"script"`
	}

	SortField map[string]SortOrder

	UpdateBody struct {
		Doc map[string]any `json:"doc,omitempty"`
	}
)
