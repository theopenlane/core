package hooks

import (
	"context"
	"errors"
	"strings"

	"entgo.io/ent"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	gqlgenerated "github.com/theopenlane/core/v2/internal/graphapi/generated"
)

var (
	gqlSchema = gqlgenerated.NewExecutableSchema(gqlgenerated.Config{}).Schema()
)

// HookReport makes sure to validate the provided query and variables that they can be executed
// against the server
func HookReport() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ReportFunc(func(ctx context.Context, m *generated.ReportMutation) (generated.Value, error) {

			query, ok := m.Query()
			variables, variablesExist := m.Variables()

			if m.Op().Is(ent.OpUpdateOne) {
				if !ok && !variablesExist && !m.VariablesCleared() {
					return next.Mutate(ctx, m)
				}

				var err error
				if !ok {
					query, err = m.OldQuery(ctx)
					if err != nil {
						return nil, err
					}
				}

				if !variablesExist && !m.VariablesCleared() {
					variables, err = m.OldVariables(ctx)
					if err != nil {
						return nil, err
					}
				}
			}

			if err := validateReportingQuery(query, variables); err != nil {
				return nil, err
			}

			// clients can always reformat if they need to
			m.SetQuery(strings.TrimSpace(query))

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func validateReportingQuery(query string, variables map[string]any) error {
	document, errs := gqlparser.LoadQueryWithRules(gqlSchema, query, nil)
	if len(errs) > 0 {
		return errs
	}

	// only queries can ever be allowed
	if len(document.Operations) != 1 || document.Operations[0].Operation != ast.Query {
		return errors.New("reporting must contain exactly one query operation")
	}

	_, err := validator.VariableValues(gqlSchema, document.Operations[0], variables)
	return err
}
