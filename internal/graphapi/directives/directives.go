package directives

import (
	"strconv"

	"entgo.io/contrib/entgql"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	// Hidden is used to mark a directive as hidden so it will be ignored in documentation
	Hidden = "hidden"
	// Skip if true, the decorated field or fragment in an operation is not resolved by the GraphQL server.
	Skip = "skip"
	// Include if false, the decorated field or fragment in an operation is not resolved by the GraphQL server.
	Include = "include"
)

// NewUndocumentedDirective returns a new hidden directive with the value set
// TODO (sfunk): this is a custom directive that needs to be implemented before we can use it
func NewHiddenDirective(v bool) entgql.Directive {
	return entgql.NewDirective(Hidden, argsWithIf(v))
}

// NewSkipDirective returns a new skip directive with the value set
func NewSkipDirective(v bool) entgql.Directive {
	return entgql.NewDirective(Skip, argsWithIf(v))
}

// NewIncludeDirective returns a new include directive with the value set
func NewIncludeDirective(v bool) entgql.Directive {
	return entgql.NewDirective(Include, argsWithIf(v))
}

// argsWithReason returns an argument with the given reason set as the value.
func argsWithReason(reason string) *ast.Argument { //nolint:unused
	return &ast.Argument{
		Name: "reason",
		Value: &ast.Value{
			Raw:  reason,
			Kind: ast.StringValue,
		},
	}
}

// argsWithReason returns an argument with the given reason set as the value.
func argsWithIf(value bool) *ast.Argument {
	return &ast.Argument{
		Name: "if",
		Value: &ast.Value{
			Raw:  strconv.FormatBool(value),
			Kind: ast.BooleanValue,
		},
	}
}
