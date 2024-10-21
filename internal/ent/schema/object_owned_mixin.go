package schema

import (
	"context"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// ObjectOwnedMixin is a mixin for object owned entities
type ObjectOwnedMixin struct {
	mixin.Schema
	// Ref table for the id
	Ref string
	// Kind of the object
	Kind any
	// FieldName is the name of the field in the schema, e.g. "owner_id" or "program_id"
	FieldName string
	// Required makes the owner id field required as input
	Required bool
	// AllowEmpty allows the owner id field to be empty
	AllowEmpty bool
	// SkipOASGeneration skips open api spec generation for the field
	SkipOASGeneration bool
	// SkipInterceptor skips the interceptor for that schema for all queries, or specific types,
	// this is useful for tokens, etc
	SkipInterceptor interceptors.SkipMode
	// SkipTokenType skips the traverser or hook if the token type is found in the context
	SkipTokenType []token.PrivacyToken
	// HookFunc is the hook function for the object owned mixin
	// that will be called on all mutations
	HookFunc HookFunc
	// InterceptorFunc is the interceptor function for the object owned mixin
	// that will be called on all queries
	InterceptorFunc InterceptorFunc
}

type HookFunc func(o ObjectOwnedMixin) ent.Hook

type InterceptorFunc func(o ObjectOwnedMixin) ent.Interceptor

// Fields of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Fields() []ent.Field {
	objectIDField := field.
		String(o.FieldName).
		Comment("the object id that owns the object")

	if !o.Required {
		objectIDField.Optional()

		// if explicitly set to allow empty values, otherwise ensure it is not empty
		if !o.AllowEmpty {
			objectIDField.NotEmpty()
		}
	}

	return []ent.Field{
		objectIDField,
	}
}

// Edges of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Edges() []ent.Edge {
	if o.Kind == nil {
		panic(errors.New("kind must be non-empty type")) // nolint: goerr113
	}

	if o.Ref == "" {
		panic(errors.New("ref must be non-empty string")) // nolint: goerr113
	}

	ownerEdge := edge.
		From("owner", o.Kind).
		Field(o.FieldName).
		Ref(o.Ref).
		Unique()

	if o.Required {
		ownerEdge.Required()
	}

	return []ent.Edge{
		ownerEdge,
	}
}

// Hooks of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Hooks() []ent.Hook {
	if o.AllowEmpty {
		// do not add hooks if the field is optional
		return []ent.Hook{}
	}

	return []ent.Hook{
		o.HookFunc(o),
	}
}

// Interceptors of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Interceptors() []ent.Interceptor {
	if o.AllowEmpty {
		// do not add interceptors if the field is optional
		return []ent.Interceptor{}
	}

	return []ent.Interceptor{
		o.InterceptorFunc(o),
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (o ObjectOwnedMixin) P(w interface{ WhereP(...func(*sql.Selector)) }, objectIDs []string) {
	w.WhereP(
		sql.FieldIn(o.FieldName, objectIDs...),
	)
}

func defaultObjectHookFunc(o ObjectOwnedMixin) []ent.Hook { // nolint:unused
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// skip the hook if the context has the token type
				// this is useful for tokens, where the user is not yet authenticated to
				// a particular organization yet and auth policy allows this
				for _, tokenType := range o.SkipTokenType {
					if rule.ContextHasPrivacyTokenOfType(ctx, tokenType) {
						return next.Mutate(ctx, m)
					}
				}

				// set owner on create mutation
				if m.Op() == ent.OpCreate {
					// TODO: set the owner id on the object
				} else {
					// TODO get IDs from the object
					objectIDs := []string{}

					// filter by owner on update and delete mutations
					mx, ok := m.(interface {
						SetOp(ent.Op)
						Client() *generated.Client
						WhereP(...func(*sql.Selector))
					})
					if !ok {
						return nil, ErrUnexpectedMutationType
					}

					o.P(mx, objectIDs)
				}

				return next.Mutate(ctx, m)
			})
		},
	}
}

var defaultObjectInterceptorFunc InterceptorFunc = func(o ObjectOwnedMixin) ent.Interceptor { // nolint:unused
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// skip the interceptor if the context has the token type
		// this is useful for tokens, where the user is not yet authenticated to
		// a particular organization yet
		for _, tokenType := range o.SkipTokenType {
			if rule.ContextHasPrivacyTokenOfType(ctx, tokenType) {
				return nil
			}
		}

		// check query context skips
		ctxQuery := ent.QueryFromContext(ctx)

		switch o.SkipInterceptor {
		case interceptors.SkipAll:
			return nil
		case interceptors.SkipOnlyQuery:
			{
				if ctxQuery.Op == "Only" {
					return nil
				}
			}
		}

		// TODO get IDs from the object
		objectIDs := []string{}

		// sets the owner object on the query
		o.P(q, objectIDs)

		return nil
	})
}
