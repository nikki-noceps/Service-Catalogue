package database

type (
	Body struct {
		Query *Query       `json:"query,omitempty"`
		Sort  []*TermQuery `json:"sort,omitempty"`
	}
	// Query represents a generic Elasticsearch query structure
	Query struct {
		Match         *MatchQuery         `json:"match,omitempty"`
		MultiMatch    *MultiMatch         `json:"multimatch,omitempty"`
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
	TermQuery struct {
		Field string `json:"field"`
		Value any    `json:"value"`
	}

	// RangeQuery represents a range query
	RangeQuery struct {
		Field string `json:"field"`
		Gte   any    `json:"gte,omitempty"`
		Lte   any    `json:"lte,omitempty"`
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
)
