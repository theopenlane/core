package schema

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/iam/fgax"
)

// ObjectOwnedMixin is a mixin for object owned entities
type ObjectOwnedMixin struct {
	mixin.Schema
	// Ref table for the id
	Ref string
	// Kind of the object
	Kind any
	// FieldNames are the name of the field in the schema that can own / controls permissions of the object, e.g. "owner_id" or "program_id"
	FieldNames []string
	// OwnerRelation is the relation of the owner (user or service) that created the object, defaults to "parent"
	OwnerRelation string
	// SkipUserTuple skips the user tuple creation for the object owned mixin
	SkipUserTuple bool
	// Required makes the owner id field required as input
	Required bool
	// AllowEmptyForSystemAdmin allows the owner id field to be empty for system admins
	AllowEmptyForSystemAdmin bool
	// SkipOASGeneration skips open api spec generation for the field
	SkipOASGeneration bool
	// SkipInterceptor skips the interceptor for that schema for all queries, or specific types,
	// this is useful for tokens, etc
	SkipInterceptor interceptors.SkipMode
	// SkipTokenType skips the traverser or hook if the token type is found in the context
	SkipTokenType []token.PrivacyToken
	// WithOrganizationOwner adds the organization owner_id field and hooks to the schema
	WithOrganizationOwner bool
	// HookFuncs is the hook functions for the object owned mixin
	// that will be called on all mutations
	HookFuncs []HookFunc
	// InterceptorFunc is the interceptor function for the object owned mixin
	// that will be called on all queries
	InterceptorFuncs []InterceptorFunc
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
	if o.HookFuncs == nil {
		o.HookFuncs = []HookFunc{defaultObjectHookFunc, defaultTupleUpdateFunc}
	}

	if o.InterceptorFuncs == nil {
		o.InterceptorFuncs = []InterceptorFunc{defaultObjectInterceptorFunc}
	}

	if o.WithOrganizationOwner {
		o.HookFuncs = append(o.HookFuncs, orgHookCreateFunc)
		o.InterceptorFuncs = append(o.InterceptorFuncs, defaultOrgInterceptorFunc)
	}

	return o
}

// Fields of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Fields() []ent.Field {
	var fields []ent.Field

	// add the organization owner field if the flag is set
	if o.WithOrganizationOwner {
		fields = append(fields,
			field.String(ownerFieldName).
				Comment("the ID of the organization owner of the object").
				Immutable(). // Immutable because it is set on creation and never changes
				Optional().  // Optional because it doesn't need to be provided as input
				NotEmpty())  // NotEmpty because it is required to be set in the database
	}

	// if the field name is not defined, skip adding fields
	if len(o.FieldNames) == 0 || o.Kind == nil {
		return fields
	}

	for _, fieldName := range o.FieldNames {
		objectType := o.Kind
		objectIDField := field.
			String(fieldName).
			Comment(fmt.Sprintf("the %v id that owns the object", getObjectType(objectType)))

		if !o.Required {
			objectIDField.Optional()

			// if explicitly set to allow empty values, otherwise ensure it is not empty
			if !o.AllowEmptyForSystemAdmin {
				objectIDField.NotEmpty()
			}
		}

		fields = append(fields, objectIDField)
	}

	return fields
}

// Edges of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Edges() []ent.Edge {
	var edges []ent.Edge

	// if there is no ref, don't add any edges
	if o.Ref == "" {
		return edges
	}

	// add the organization owner edge if the flag is set
	if o.WithOrganizationOwner {
		edges = append(edges,
			edge.From("owner", Organization.Type).
				Field(ownerFieldName).
				Immutable().
				Unique().
				Ref(o.Ref))
	}

	if o.Kind == nil {
		return edges
	}

	for _, fieldName := range o.FieldNames {
		ownerEdge := edge.
			From("owner", o.Kind).
			Field(fieldName).
			Ref(o.Ref).
			Unique()

		if o.Required {
			ownerEdge.Required()
		}

		edges = append(edges, ownerEdge)
	}

	return edges
}

// Hooks of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Hooks() []ent.Hook {
	res := []ent.Hook{}
	for _, hookFunc := range o.HookFuncs {
		res = append(res, hookFunc(o))
	}

	return res
}

// Interceptors of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Interceptors() []ent.Interceptor {
	res := []ent.Interceptor{}
	for _, interceptorFunc := range o.InterceptorFuncs {
		res = append(res, interceptorFunc(o))
	}

	return res
}

// P adds a storage-level predicate to the queries and mutations for the provided field name
func (o ObjectOwnedMixin) PWithField(w interface{ WhereP(...func(*sql.Selector)) }, fieldName string, objectIDs []string) {
	selector := sql.FieldIn(fieldName, objectIDs...)
	if o.AllowEmptyForSystemAdmin && fieldName == ownerFieldName {
		// allow for empty values if the flag is set
		w.WhereP(
			sql.OrPredicates(
				sql.FieldIsNull(fieldName),
				selector,
			),
		)

		return
	}

	// otherwise we are using getting all objects filtered by the field name
	// usually "id" or "owner_id"
	w.WhereP(selector)
}

// P adds the predicate to the queries, using the "id" field
func (o ObjectOwnedMixin) P(w interface{ WhereP(...func(*sql.Selector)) }, objectIDs []string) {
	o.PWithField(w, "id", objectIDs)

	return
}

// defaultTupleUpdateFunc is the default hook function for the object owned mixin
// to add tuples to the database when creating or updating an object based on the edges
// that can own the object
var defaultTupleUpdateFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	ownerRelation := fgax.ParentRelation
	if o.OwnerRelation != "" {
		ownerRelation = o.OwnerRelation
	}

	return hook.On(
		hooks.HookObjectOwnedTuples(o.FieldNames, ownerRelation, o.SkipUserTuple),
		ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
	)
}

// defaultObjectHookFunc is the default hook function for the object owned mixin
var defaultObjectHookFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// skip the hook if the context has the token type
			// this is useful for tokens, where the user is not yet authenticated to
			// a particular organization yet and auth policy allows this
			if skip := rule.SkipTokenInContext(ctx, o.SkipTokenType); skip {
				return next.Mutate(ctx, m)
			}

			skip, err := o.skipOrgHookForAdmins(ctx, m)
			if err != nil {
				return nil, err
			}

			if skip {
				return next.Mutate(ctx, m)
			}

			objectIDs, err := interceptors.GetAuthorizedObjectIDs(ctx, strcase.SnakeCase(m.Type()))
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

			return next.Mutate(ctx, m)
		})
	}, ent.OpUpdateOne|ent.OpUpdate|ent.OpDelete|ent.OpDeleteOne)
}

// defaultObjectInterceptorFunc is the default interceptor function for the object owned mixin
// it will filter the query to only include the objects that the user has access to based on the FGA list objects
// setting
var defaultObjectInterceptorFunc InterceptorFunc = func(o ObjectOwnedMixin) ent.Interceptor { // nolint:unused
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		return interceptors.AddIDPredicate(ctx, q)
	})
}

// getObjectType takes the `kind` and returns the object type
// this should be type of the schema, e.g. `func(schema.Organization)` which will return `organization`
func getObjectType(kind any) string {
	objectType := reflect.TypeOf(kind).String()

	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(objectType, "func(schema.", ""), ")", ""))
}

// skipOrgHookForAdmins checks if the hook should be skipped for the given mutation for system admins
func (o ObjectOwnedMixin) skipOrgHookForAdmins(ctx context.Context, m ent.Mutation) (bool, error) {
	if o.AllowEmptyForSystemAdmin {
		isAdmin, err := rule.CheckIsSystemAdmin(ctx, m)
		if err != nil {
			return false, err
		}

		// skip hook for system admins to create system level objects
		if isAdmin {
			return true, nil
		}
	}

	return false, nil
}
