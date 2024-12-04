package entitlements

import (
	"fmt"
	"sort"
	"strings"
)

const metadataFormat = "metadata['%s']:'%s'"

const clausesPerQuery = 10

type QueryBuilder struct {
	operator string
	keys     map[string]string
	OrgCust  *OrganizationCustomer
}

func NewQueryBuilder(options ...func(*QueryBuilder)) *QueryBuilder {
	qb := &QueryBuilder{
		keys: make(map[string]string),
	}

	for _, option := range options {
		option(qb)
	}
	return qb
}

func WithOperator(operator string) func(*QueryBuilder) {
	return func(qb *QueryBuilder) {
		qb.operator = operator
	}
}

func WithKeys(keys map[string]string) func(*QueryBuilder) {
	return func(qb *QueryBuilder) {
		qb.keys = keys
	}
}

func WithOrganizationCustomer(orgCust *OrganizationCustomer) func(*QueryBuilder) {
	return func(qb *QueryBuilder) {
		qb.OrgCust = orgCust
	}
}

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
