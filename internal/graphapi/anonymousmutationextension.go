package graphapi

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

const (
	// bulkMutationMarker matches every bulk mutation: create, update, delete, clone, and the CSV uploads
	bulkMutationMarker = "Bulk"
	// deleteMutationPrefix matches every delete mutation as a fall through, the privacy rules should already deny this as well
	deleteMutationPrefix = "delete"
)

// anonymousMutationExtension denies anonymous callers that are not trust center callers every mutation, denies
// anonymous trust center callers bulk and delete mutations, and rejects any create or update input field
// not listed in the schema's entx.AnonymousFields annotation
type anonymousMutationExtension struct {
	// allowedFields maps a mutation input type name to the input fields anonymous callers may set
	allowedFields map[string]map[string]bool
}

// newAnonymousMutationExtension builds the allowed input fields from the entityops schema registry
func newAnonymousMutationExtension() *anonymousMutationExtension {
	allowedFields := map[string]map[string]bool{}

	for _, schema := range entityops.AllSchemas() {
		if len(schema.AnonymousInputFields) == 0 {
			continue
		}

		allowed := map[string]bool{}
		for _, f := range schema.AnonymousInputFields {
			allowed[f] = true
		}

		allowedFields["Create"+schema.Name+"Input"] = allowed
		allowedFields["Update"+schema.Name+"Input"] = allowed
	}

	return &anonymousMutationExtension{allowedFields: allowedFields}
}

// ExtensionName satisfies the extension interface
func (*anonymousMutationExtension) ExtensionName() string { return "AnonymousMutations" }

// Validate satisfies the extension interface
func (*anonymousMutationExtension) Validate(graphql.ExecutableSchema) error { return nil }

// InterceptField denies anonymous callers disallowed mutations and any input field they are not allowed to set
func (e *anonymousMutationExtension) InterceptField(ctx context.Context, next graphql.Resolver) (any, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil || fc.Object != "Mutation" || !auth.IsAnonymousFromContext(ctx) {
		return next(ctx)
	}

	// only anonymous trust center callers may run mutations, other anonymous tokens such as questionnaires never can
	if !auth.HasAnonymousTrustCenterCapability(ctx) {
		return nil, rout.ErrPermissionDenied
	}

	if strings.Contains(fc.Field.Name, bulkMutationMarker) || strings.HasPrefix(fc.Field.Name, deleteMutationPrefix) {
		return nil, rout.ErrPermissionDenied
	}

	variables := graphql.GetOperationContext(ctx).Variables

	for _, arg := range fc.Field.Arguments {
		def := fc.Field.Definition.Arguments.ForName(arg.Name)
		if def == nil {
			continue
		}

		allowed, ok := e.allowedFields[def.Type.Name()]
		if !ok {
			continue
		}

		value, err := arg.Value.Value(variables)
		if err != nil {
			return nil, err
		}

		input, ok := value.(map[string]any)
		if !ok {
			continue
		}

		for name := range input {
			if !allowed[name] {
				logx.FromContext(ctx).Warn().Str("mutation", fc.Field.Name).Str("input_type", def.Type.Name()).Str("field", name).Msg("anonymous caller attempted to set a disallowed input field")

				return nil, rout.ErrPermissionDenied
			}
		}
	}

	return next(ctx)
}

// WithAnonymousMutationBlock registers the anonymous mutation extension
func WithAnonymousMutationBlock(srv *handler.Server) {
	srv.Use(newAnonymousMutationExtension())
}
