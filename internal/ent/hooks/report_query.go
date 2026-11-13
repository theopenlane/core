//go:build !codegen

package hooks

import (
	"errors"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"

	gqlgenerated "github.com/theopenlane/core/v2/internal/graphapi/generated"
)

var gqlSchema = gqlgenerated.NewExecutableSchema(gqlgenerated.Config{}).Schema()

// isQueryValid checks if the provided query string is valid against the server. it also validates
// the variables too
func isQueryValid(query string, variables map[string]any) error {
	document, documentErr := gqlparser.LoadQueryWithRules(gqlSchema, query, nil)
	if len(documentErr) > 0 {
		return documentErr
	}

	// only queries should ever be allowed + one query per report
	if len(document.Operations) != 1 || document.Operations[0].Operation != ast.Query {
		return errors.New("reporting must contain exactly one query operation") //nolint:err113
	}

	_, err := validator.VariableValues(gqlSchema, document.Operations[0], variables)
	return err
}
