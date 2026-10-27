package graphapi

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/rout"
	"github.com/vektah/gqlparser/v2/ast"
)

// anonymousQueryExtension denies anonymous callers every root query that is not an ent schema query,
// so hand-written query resolvers that never reach ent privacy and interceptors are closed to anonymous callers by default
type anonymousQueryExtension struct {
	// schemaQueries maps a lowercased ent schema query name to the type it returns
	schemaQueries map[string]string
}

// newAnonymousQueryExtension builds the ent schema query names from the entityops schema registry
func newAnonymousQueryExtension() anonymousQueryExtension {
	plural := pluralize.NewClient()
	schemaQueries := map[string]string{}

	for _, schema := range entityops.AllSchemas() {
		name := strings.ToLower(schema.Name)

		// like ent, a name with no distinct plural gets a slice suffix for its list query (e.g. documentDataSlice)
		list := plural.Plural(name)
		if list == name {
			list += "slice"
		}

		schemaQueries[name] = schema.Name
		schemaQueries[list] = schema.Name + "Connection"
	}

	return anonymousQueryExtension{schemaQueries: schemaQueries}
}

// ExtensionName satisfies the extension interface
func (anonymousQueryExtension) ExtensionName() string { return "AnonymousQueries" }

// Validate satisfies the extension interface
func (anonymousQueryExtension) Validate(graphql.ExecutableSchema) error { return nil }

// InterceptRootField denies anonymous callers root queries that are not ent schema queries, it runs once per top level field
func (e anonymousQueryExtension) InterceptRootField(ctx context.Context, next graphql.RootResolver) graphql.Marshaler {
	if !auth.IsAnonymousFromContext(ctx) || graphql.GetOperationContext(ctx).Operation.Operation != ast.Query {
		return next(ctx)
	}

	field := graphql.GetRootFieldContext(ctx).Field

	if e.schemaQueries[strings.ToLower(field.Name)] != field.Definition.Type.Name() {
		logx.FromContext(ctx).Warn().Str("query", field.Name).Msg("anonymous caller attempted a query that is not an ent schema query")

		graphql.AddError(ctx, rout.ErrPermissionDenied)

		return graphql.Null
	}

	return next(ctx)
}

// WithAnonymousQueryBlock registers the anonymous query extension
func WithAnonymousQueryBlock(srv *handler.Server) {
	srv.Use(newAnonymousQueryExtension())
}
