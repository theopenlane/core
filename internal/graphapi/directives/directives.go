package directives

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/lo"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	gqlhistorygenerated "github.com/theopenlane/core/internal/graphapi/historygenerated"

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/vektah/gqlparser/v2/ast"
)

var ErrReadOnlyField = errors.New("invalid input: attempted to set read only field")

const (
	// Hidden is used to mark a directive as hidden so it will be ignored in documentation
	Hidden = "hidden"
	// ReadOnly is used to mark a field as read only so it will be ignored in mutations
	ReadOnly = "readOnly"
	// ExternalReadOnly is used to mark a field as read only because it is populated by an external source
	ExternalReadOnly = "externalReadOnly"
	// ExternalSource is used to mark a field or object as being populated by an external source which
	// would affect the ability to update the field
	ExternalSource = "externalSource"
)

// ImplementAllDirectives is a helper function that can be used to add all active directives to the gqlgen config
// in the resolver setup
func ImplementAllDirectives(cfg *gqlgenerated.Config) {
	cfg.Directives.Hidden = HiddenDirective
	cfg.Directives.ReadOnly = ReadOnlyDirective
	cfg.Directives.ExternalReadOnly = ExternalReadOnlyDirective
	cfg.Directives.ExternalSource = ExternalSourceDirective
}

// ImplementAllHistoryDirectives is a helper function that can be used to add all active directives to the gqlgen config
// in the resolver setup for the history api
func ImplementAllHistoryDirectives(cfg *gqlhistorygenerated.Config) {
	cfg.Directives.Hidden = HiddenDirective
	cfg.Directives.ReadOnly = ReadOnlyDirective
	cfg.Directives.ExternalReadOnly = ExternalReadOnlyDirective
	cfg.Directives.ExternalSource = ExternalSourceDirective
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
	return entgql.NewDirective(ReadOnly)
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

	if checkFieldSet(ctx, noSkip) {
		return nil, ErrReadOnlyField
	}

	// otherwise, continue to the next resolver
	return next(ctx)
}

// NewExternalReadyOnlyDirective returns a new hidden directive with the value set
// to add @externalReadOnly to a field or object
func NewExternalReadyOnlyDirective(v enums.ControlSource) entgql.Directive {
	return entgql.NewDirective(ExternalReadOnly, argsWithControlSource(v))
}

// ExternalReadOnlyDirectiveAnnotation is an annotation that can be used to mark a field as read-only when
// the object is system-owned because the field is populated by an external source
var ExternalReadOnlyDirectiveAnnotation = entgql.Directives(
	NewExternalReadyOnlyDirective(enums.ControlSourceFramework),
)

// ExternalReadOnlyDirective is the implementation for the external read only directive that can be used to indicate a field cannot be set by users for objects that are system-owned because it is populated by an external source
// only system admins can change this field on system-owned objects, on objects that are not system-owned, the field can be set by anyone with permission to update the object
var ExternalReadOnlyDirective = func(ctx context.Context, _ any, next graphql.Resolver, source *enums.ControlSource) (any, error) {
	if admin, err := rule.CheckIsSystemAdminWithContext(ctx); err == nil && admin {
		// if the user is a system admin, always return the field
		return next(ctx)
	}

	fieldSet := checkFieldSet(ctx, skipCreateOperations)
	allowed := checkSourceAllowed(ctx, source)

	if fieldSet && !allowed {
		return nil, ErrReadOnlyField
	}

	// otherwise, continue to the next resolver
	return next(ctx)
}

// NewExternalSourceDirective returns a new directive with the value set
// to add @externalSource(source: v) to a field or object
func NewExternalSourceDirective(v enums.ControlSource) entgql.Directive {
	return entgql.NewDirective(ExternalSource, argsWithControlSource(v))
}

// ExternalSourceDirectiveAnnotation is an annotation that can be used to mark a field as read-only when
// the object is system-owned because the field is populated by an external source
var ExternalSourceDirectiveAnnotation = entgql.Directives(
	NewExternalSourceDirective(enums.ControlSourceFramework),
)

// ExternalSourceDirective is used to mark fields or objects that are populated by an external source
// that will prevent the ability to update the field if the object is framework sourced
var ExternalSourceDirective = func(ctx context.Context, _ any, next graphql.Resolver, _ *enums.ControlSource) (any, error) {
	// this is a no-op, this is only used for setting the annotations on the schema
	// to affect the mutation inputs
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

// argsWithReason returns an argument with the given reason set as the value.
func argsWithControlSource(value enums.ControlSource) *ast.Argument {
	return &ast.Argument{
		Name: "source",
		Value: &ast.Value{
			Raw:  value.String(),
			Kind: ast.EnumValue,
		},
	}
}

func checkFieldSet(ctx context.Context, skip func(*graphql.OperationContext) bool) bool {
	operationContext := graphql.GetOperationContext(ctx)
	if operationContext == nil || operationContext.Operation == nil {
		// if we can't get the operation context, continue to the next resolver
		return false
	}

	// if this is a mutation, check if the field is being set on update operations
	// we don't care about create operations s should not be set on create
	if operationContext.Operation.Operation == ast.Mutation && !skip(operationContext) {
		input := graphutils.GetMapInputVariableByName(ctx, graphutils.GetInputFieldVariableName(ctx))
		if input == nil {
			return false
		}

		// this will look like map[string]any{"fieldName": "value"}
		for _, val := range *input {
			if !lo.IsEmpty(&val) {
				// if any value is being set in the input, return an error
				return true
			}
		}
	}

	return false
}

// checkIsSystemOwned checks if the object is system owned
// this is used in the externalReadOnly directive to prevent non-system admin users
// from setting fields that are populated by an external source on system-owned objects
func checkSourceAllowed(ctx context.Context, restrictedSource *enums.ControlSource) bool {
	if restrictedSource == nil {
		return true
	}

	id := graphutils.GetStringInputVariableByName(ctx, "id")
	if id == nil {
		logx.FromContext(ctx).Error().Msg("no id found in context for externalReadOnly directive")
		return true
	}

	// now get the object from the database
	client := ent.FromContext(ctx)
	if client == nil {
		logx.FromContext(ctx).Error().Msg("no ent client found in context for externalReadOnly directive")
		return true
	}

	var objSource *enums.ControlSource
	obj, err := client.Control.Get(ctx, *id)
	if err == nil {
		objSource = &obj.Source
	} else {
		obj, err := client.Subcontrol.Get(ctx, *id)
		if err != nil {
			logx.FromContext(ctx).Error().Msg("failed to check for object source in externalReadOnly directive")

			return true
		}

		objSource = &obj.Source
	}

	// only allow if the source is different than the one on the object
	// the specified source is not allowed to make changes
	return *objSource != *restrictedSource
}

// skipCreateOperations is a helper function that can be used to skip create operations
var skipCreateOperations = func(operationContext *graphql.OperationContext) bool {
	// skip if the source is nil, this means the directive was applied to an object
	// and not a field, so we don't need to check if the field is being set
	return strings.Contains(operationContext.Operation.Name, "Create")
}

// noSkip is a helper function that can be used to never skip any operations
var noSkip = func(_ *graphql.OperationContext) bool {
	return false
}
