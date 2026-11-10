package genhelpers

import (
	"slices"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"github.com/vektah/gqlparser/v2/ast"
)

// droppedWhereOps lists, per GraphQL scalar, the predicate suffixes that are pruned from the
// generated where inputs. this must stay in sync with the guard in templates/entgql/gql_where.tmpl,
// which prunes the matching Go struct fields; entgql builds the GraphQL schema from its own code rather than that template,
// so both sides need the same rule
var droppedWhereOps = map[string][]string{
	"ID":       {"GT", "GTE", "LT", "LTE"}, // strings should never need GT/LT comparison
	"String":   {"GT", "GTE", "LT", "LTE"}, // strings should never need GT/LT comparison
	"Time":     {"NEQ", "In", "NotIn"},     // time uses GT/LT comparison, not specific or in logic
	"DateTime": {"NEQ", "In", "NotIn"},     // time uses GT/LT comparison, not specific or in logic
	"Int":      {"In", "NotIn"},            // Ints use comparison, not in logic
}

// PruneWhereInputOps removes the predicate fields listed in droppedWhereOps from every generated
// where input. A field is only removed when the same input also declares its base field, so a real
// schema field whose name happens to end in an op suffix is left alone
func PruneWhereInputOps() entgql.SchemaHook {
	return func(_ *gen.Graph, s *ast.Schema) error {
		for _, t := range s.Types {
			if t.Kind != ast.InputObject || !strings.HasSuffix(t.Name, "WhereInput") {
				continue
			}

			base := make(map[string]bool, len(t.Fields))
			for _, f := range t.Fields {
				base[f.Name] = true
			}

			kept := make(ast.FieldList, 0, len(t.Fields))

			for _, f := range t.Fields {
				if !prunedOp(f, base) {
					kept = append(kept, f)
				}
			}

			t.Fields = kept
		}

		return nil
	}
}

// historyWhereFields are the only history schema fields exposed as where filters, history is an audit trail filtered by object, actor, and time
var historyWhereFields = []string{"id", "ref", "operation", "history_time", "created_at", "updated_at", "created_by", "updated_by", "updated_by_impersonator", "owner_id", "revision"}

// whereOpSuffixes are the predicate suffixes entgql appends to a where input base field
var whereOpSuffixes = []string{"NEQ", "In", "NotIn", "GT", "GTE", "LT", "LTE", "Contains", "HasPrefix", "HasSuffix", "EqualFold", "ContainsFold", "IsNil", "NotNil"}

// IsHistoryWhereField reports whether a history schema field, by its ent name, is exposed as a where filter
func IsHistoryWhereField(name string) bool {
	return slices.Contains(historyWhereFields, name)
}

// PruneHistoryWhereInputs removes every history where input field that does not filter on historyWhereFields,
// gql_where.tmpl applies the same rule to the Go structs through IsHistoryWhereField
func PruneHistoryWhereInputs() entgql.SchemaHook {
	camel := gen.Funcs["camel"].(func(string) string)

	keep := map[string]bool{"and": true, "or": true, "not": true}
	for _, name := range historyWhereFields {
		keep[camel(name)] = true
	}

	return func(_ *gen.Graph, s *ast.Schema) error {
		for _, t := range s.Types {
			if t.Kind != ast.InputObject || !strings.HasSuffix(t.Name, "HistoryWhereInput") {
				continue
			}

			kept := make(ast.FieldList, 0, len(t.Fields))

			for _, f := range t.Fields {
				if keptHistoryWhereField(f.Name, keep) {
					kept = append(kept, f)
				}
			}

			t.Fields = kept
		}

		return nil
	}
}

// keptHistoryWhereField reports whether the where input field is a kept base field or a predicate on one
func keptHistoryWhereField(name string, keep map[string]bool) bool {
	if keep[name] {
		return true
	}

	for _, op := range whereOpSuffixes {
		if strings.HasSuffix(name, op) && keep[strings.TrimSuffix(name, op)] {
			return true
		}
	}

	return false
}

// prunedOp reports whether the field is a predicate this codebase does not expose
func prunedOp(f *ast.FieldDefinition, base map[string]bool) bool {
	for _, op := range droppedWhereOps[f.Type.Name()] {
		if !strings.HasSuffix(f.Name, op) {
			continue
		}

		// only prune when the base field is present, which marks this as a generated predicate
		if base[strings.TrimSuffix(f.Name, op)] {
			return true
		}
	}

	return false
}
