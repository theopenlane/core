package directives

import (
	"context"
	"errors"
	"strconv"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/vektah/gqlparser/v2/ast"
)

var ErrReadOnlyField = errors.New("invalid input: attempted to set read only field")

const (
	// Hidden is used to mark a directive as hidden so it will be ignored in documentation
	Hidden = "hidden"
	// ReadOnly is used to mark a field as read only so it will be ignored in mutations
	ReadOnly = "readOnly"
)

// ImplementAllDirectives is a helper function that can be used to add all active directives to the gqlgen config
// in the resolver setup
func ImplementAllDirectives(cfg *gqlgenerated.Config) {
	cfg.Directives.Hidden = HiddenDirective
	cfg.Directives.ReadOnly = ReadOnlyDirective
}

// NewHiddenDirective returns a new hidden directive with the value set
// to add @hidden(if: true) to a field or object
// this is used to hide fields from graphql schema as well as
// to only return the field for system admins
func NewHiddenDirective(v bool) entgql.Directive {
	return entgql.NewDirective(Hidden, argsWithIf(v))
}

// HiddenDirectiveAnnotation is an annotation that can be used to hide a field from the graphql schema
// this is added to the ent schema field annotations
var HiddenDirectiveAnnotation = entgql.Directives(
	NewHiddenDirective(true),
)

// HiddenDirective is the implementation for the hidden directive that can be used to hide a field from non-system admin users
// if the user is a system admin, the field will be returned
// otherwise, the field will be returned as nil
var HiddenDirective = func(ctx context.Context, _ any, next graphql.Resolver, isHidden *bool) (any, error) {
	if admin, err := rule.CheckIsSystemAdminWithContext(ctx); err == nil && admin {
		// if the user is a system admin, always return the field
		return next(ctx)
	}

	if isHidden != nil && *isHidden {
		// if the field is marked as hidden, return nil
		return nil, nil
	}

	// otherwise, continue to the next resolver
	return next(ctx)
}

// NewReadOnlyDirective returns a new readOnly directive to mark a field as read only
func NewReadOnlyDirective() entgql.Directive {
	return entgql.NewDirective(ReadOnly, nil)
}

// ReadOnlyDirectiveAnnotation is an annotation that can be used to mark a field as read only
var ReadOnlyDirectiveAnnotation = entgql.Directives(
	NewReadOnlyDirective(),
)

// ReadOnlyDirective is the implementation for the readOnly directive that can be used to mark input fields as read only
// this will prevent the field from being used in create and update mutations
var ReadOnlyDirective = func(ctx context.Context, _ any, next graphql.Resolver) (any, error) {
	// first check to make sure this is a graphql request
	if !graphql.HasOperationContext(ctx) {
		// if we can't get the operation context, continue to the next resolver
		return next(ctx)
	}

	// check if the user is a system admin, if so allow the mutation
	if admin, err := rule.CheckIsSystemAdminWithContext(ctx); err == nil && admin {
		// if the user is a system admin, always return the field
		return next(ctx)
	}

	operationContext := graphql.GetOperationContext(ctx)
	if operationContext == nil || operationContext.Operation == nil {
		// if we can't get the operation context, continue to the next resolver
		return next(ctx)
	}

	// if this is a mutation, check if the field is being set
	if operationContext.Operation.Operation == ast.Mutation {
		input := graphutils.GetMapInputVariableByName(ctx, graphutils.GetInputFieldVariableName(ctx))
		if input == nil {
			return next(ctx)
		}

		// this will look like map[string]any{"fieldName": "value"}
		for _, val := range *input {
			if !lo.IsEmpty(&val) {
				// if any value is being set in the input, return an error
				return nil, ErrReadOnlyField
			}
		}
	}

	// otherwise, continue to the next resolver
	return next(ctx)
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
