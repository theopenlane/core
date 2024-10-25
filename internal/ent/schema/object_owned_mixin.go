package schema

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

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

// NewObjectOwnMixinWithRef creates a new ObjectOwnedMixin with the given ref
// and sets the defaults
func NewObjectOwnMixinWithRef(ref string) ObjectOwnedMixin {
	return NewObjectOwnedMixin(
		ObjectOwnedMixin{
			Ref: ref,
		})
}

// NewObjectOwnedMixin creates a new ObjectOwnedMixin with the given ObjectOwnedMixin
// and sets the HookFunc to defaultOrgHookFunc
func NewObjectOwnedMixin(o ObjectOwnedMixin) ObjectOwnedMixin {
	o.Kind = Organization.Type

	if o.HookFunc == nil {
		o.HookFunc = defaultObjectHookFunc
	}

	if o.InterceptorFunc == nil {
		o.InterceptorFunc = defaultObjectInterceptorFunc
	}

	return o
}

// Fields of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Fields() []ent.Field {
	// if the field name is not defined, skip adding a field
	if o.FieldName == "" {
		return []ent.Field{}
	}

	objectType := o.Kind
	objectIDField := field.
		String(o.FieldName).
		Comment(fmt.Sprintf("the %v id that owns the object", getObjectType(objectType)))

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

	// if there is no ref, don't add the edge
	if o.Ref == "" {
		return []ent.Edge{}
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

// defaultObjectHookFunc is the default hook function for the object owned mixin
var defaultObjectHookFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// skip the hook if the context has the token type
			// this is useful for tokens, where the user is not yet authenticated to
			// a particular organization yet and auth policy allows this
			for _, tokenType := range o.SkipTokenType {
				if rule.ContextHasPrivacyTokenOfType(ctx, tokenType) {
					return next.Mutate(ctx, m)
				}
			}

			if m.Op() != ent.OpCreate {
				objectIDs, err := interceptors.GetAuthorizedObjectIDs(ctx, m.Type())
				if err != nil {
					return nil, err
				}

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
	}
}

// emptyObjectHookFunc is a hook function that does not add any hooks but satisfies the HookFunc type
var emptyObjectHookFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			return next.Mutate(ctx, m)
		})
	}
}

// defaultObjectInterceptorFunc is the default interceptor function for the object owned mixin
// it will filter the query to only include the objects that the user has access to based on the FGA list objects
// setting
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

		return interceptors.AddIDPredicate(ctx, q)
	})
}

// getObjectType takes the `kind` and returns the object type
// this should be type of the schema, e.g. `func(schema.Organization)` which will return `organization`
func getObjectType(kind any) string {
	objectType := reflect.TypeOf(kind).String()

	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(objectType, "func(schema.", ""), ")", ""))
}
