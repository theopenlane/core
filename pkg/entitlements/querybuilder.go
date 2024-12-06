package entitlements

import (
	"fmt"
	"sort"
	"strings"
)

// metaDataFormat is the format for the metadata query to stripe during a search
const metadataFormat = "metadata['%s']:'%s'"

// clausesPerQuery is the number of clauses per query
const clausesPerQuery = 10

// QueryBuilder is a struct that holds the query builder
type QueryBuilder struct {
	// AND or OR
	operator string
	// keys is a map of keys and values that go into the metadataFormat
	keys map[string]string
}

// NewQueryBuilder creates a new query builder with included options
func NewQueryBuilder(options ...func(*QueryBuilder)) *QueryBuilder {
	qb := &QueryBuilder{
		keys: make(map[string]string),
	}

	for _, option := range options {
		option(qb)
	}

	return qb
}

// WithOperator sets the operator for the query builder (AND or OR)
func WithOperator(operator string) func(*QueryBuilder) {
	return func(qb *QueryBuilder) {
		qb.operator = operator
	}
}

// WithKeys sets the keys for the query builder
func WithKeys(keys map[string]string) func(*QueryBuilder) {
	return func(qb *QueryBuilder) {
		qb.keys = keys
	}
}

// BuildQuery builds the query from the keys
func (qb *QueryBuilder) BuildQuery() []string {
	var queries []string
	var currentQuery strings.Builder

	clausesCount := 0

	// Sort the keys
	keys := make([]string, 0, len(qb.keys))
	for k := range qb.keys {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for i, k := range keys {
		v := qb.keys[k]

		if clausesCount > 0 {
			currentQuery.WriteString(qb.operator)
		}

		currentQuery.WriteString(fmt.Sprintf(metadataFormat, k, v))

		clausesCount++

		if clausesCount == clausesPerQuery || i == len(keys)-1 {
			queries = append(queries, currentQuery.String())

			currentQuery.Reset()

			clausesCount = 0
		}
	}

	return queries
}
