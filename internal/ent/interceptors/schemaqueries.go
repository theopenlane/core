package interceptors

import (
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

// SchemaQuery holds the graphql root query field names generated for an ent schema
type SchemaQuery struct {
	// Single is the query returning one object, e.g. trustCenterDoc
	Single string
	// List is the connection query, e.g. trustCenterDocs, or documentDataSlice when the name has no distinct plural
	List string
}

// SchemaQueries returns the graphql query names for every ent schema keyed by schema name, built once from the entityops registry
var SchemaQueries = sync.OnceValue(func() map[string]SchemaQuery {
	queries := make(map[string]SchemaQuery, len(entityops.AllSchemas()))

	for _, schema := range entityops.AllSchemas() {
		single := lowerFirst(schema.Name)

		// like ent, a name with no distinct plural gets a slice suffix for its list query
		list := strcase.LowerCamelCase(schema.Plural)
		if list == single {
			list += "Slice"
		}

		queries[schema.Name] = SchemaQuery{Single: single, List: list}
	}

	return queries
})

// listQueryName returns the graphql list query name for an ent schema type
func listQueryName(schemaType string) string {
	return SchemaQueries()[schemaType].List
}

// lowerFirst lowercases the first rune, turning a schema name into its graphql field name
func lowerFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}

	return string(unicode.ToLower(r)) + s[size:]
}
