package schema

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
)

// ObjectOwnedMixin is a mixin for object owned entities
type ObjectOwnedMixin struct {
	mixin.Schema
	// Ref table for the id
	Ref string
	// Kind of the object, only used for Organization owned flavors, its set by default either
	// with the OrgOwnedMixin or using the withOrganizationOwner option
	Kind any
	// FieldNames are the name of the field in the schema that can own / controls permissions of the object, e.g. "owner_id" or "program_id"
	FieldNames []string
	// OwnerRelation is the relation of the owner (user or service) that created the object, defaults to "parent"
	OwnerRelation string
	// AllowEmptyForSystemAdmin allows the owner id field to be empty for system admins
	AllowEmptyForSystemAdmin bool
	// SkipInterceptor skips the interceptor for that schema for all queries, or specific types,
	// this is useful for tokens, etc
	SkipInterceptor interceptors.SkipMode
	// SkipTokenType skips the traverser or hook if the token type is found in the context
	SkipTokenType []token.PrivacyToken
	// IncludeOrganizationOwner adds the organization owner_id field and hooks to the schema
	IncludeOrganizationOwner bool
	// HookFuncs is the hook functions for the object owned mixin
	// that will be called on all mutations
	HookFuncs []HookFunc
	// InterceptorFunc is the interceptor function for the object owned mixin
	// that will be called on all queries
	InterceptorFuncs []InterceptorFunc
}

type HookFunc func(o ObjectOwnedMixin) ent.Hook

type InterceptorFunc func(o ObjectOwnedMixin) ent.Interceptor

// newOrgOwnedMixin creates a new OrgOwnedMixin using the plural name of the schema
// and all defaults. The schema must implement the SchemaFuncs interface to be used.
// options can be passed to customize the mixin
func newObjectOwnedMixin[V any](schema any, opts ...objectOwnedOption) ObjectOwnedMixin {
	sch := toSchemaFuncs(schema)

	// defaults settings
	o := ObjectOwnedMixin{
		Ref:       sch.PluralName(),
		HookFuncs: []HookFunc{defaultObjectHookFunc, defaultTupleUpdateFunc},
		InterceptorFuncs: []InterceptorFunc{func(_ ObjectOwnedMixin) ent.Interceptor {
			return interceptors.FilterQueryResults[V]()
		}},
		OwnerRelation: fgax.ParentRelation,
	}

	// apply options
	for _, opt := range opts {
		opt(&o)
	}

	if (!o.IncludeOrganizationOwner) && o.AllowEmptyForSystemAdmin {
		log.Fatal().Msg("ObjectOwnedMixin: AllowEmptyForSystemAdmin cannot be set to true if WithOrganizationOwner is false")
	}

	return o
}

// withSkipTokenTypesObjects allows to set custom token types to skip the traverser or hook
func withSkipTokenTypesObjects(tokens ...token.PrivacyToken) objectOwnedOption {
	return func(o *ObjectOwnedMixin) {
		o.SkipTokenType = tokens
	}
}

// withHookFuncs allows to set custom hook functions
func withHookFuncs(hookFuncs ...HookFunc) objectOwnedOption {
	return func(o *ObjectOwnedMixin) {
		if hookFuncs == nil {
			o.HookFuncs = []HookFunc{}

			return
		}

		o.HookFuncs = hookFuncs
	}
}

// withSkipForSystemAdmin allows the owner id field to be empty for system admins
func withSkipForSystemAdmin(allow bool) objectOwnedOption {
	return func(o *ObjectOwnedMixin) {
		o.AllowEmptyForSystemAdmin = allow
	}
}

// withAllowGlobalRead allows unauthenticated queries
func withAllowGlobalRead(allow bool) objectOwnedOption {
	return func(o *ObjectOwnedMixin) {
		if allow {
			o.SkipInterceptor = interceptors.SkipAll
		}
	}
}

// withOwnerRelation allows to set custom owner relation for the object, the default is "parent"
func withOwnerRelation(relation string) objectOwnedOption {
	return func(o *ObjectOwnedMixin) {
		o.OwnerRelation = relation
	}
}

// withParents allows to set custom parents for the object and it will automatically
// set the field name to be <parent>_id
func withParents(schemas ...any) objectOwnedOption {
	return func(o *ObjectOwnedMixin) {
		for _, schema := range schemas {
			sch := toSchemaFuncs(schema)

			o.FieldNames = append(o.FieldNames, fmt.Sprintf("%s_id", sch.Name()))
		}
	}
}

// withOrganizationOwner adds the organization owner_id field and hooks to the schema
// and optionally allows system admins to have empty owner_id
func withOrganizationOwner(skipSystemAdmin bool) objectOwnedOption {
	return func(o *ObjectOwnedMixin) {
		o.IncludeOrganizationOwner = true

		if skipSystemAdmin {
			o.AllowEmptyForSystemAdmin = skipSystemAdmin
		}

		o.HookFuncs = append(o.HookFuncs, orgHookCreateFunc)
		o.InterceptorFuncs = append(o.InterceptorFuncs, defaultOrgInterceptorFunc)
	}
}

// Fields of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Fields() []ent.Field {
	var fields []ent.Field

	// add the organization owner field if the flag is set
	if o.IncludeOrganizationOwner {
		fields = append(fields,
			field.String(ownerFieldName).
				Comment("the ID of the organization owner of the object").
				Immutable(). // Immutable because it is set on creation and never changes
				Optional().  // Optional because it doesn't need to be provided as input
				NotEmpty())  // NotEmpty because it is required to be set in the database
	}

	// if the field name is not defined, skip adding fields
	// this only happens for the org owned objects
	if len(o.FieldNames) == 0 || o.Kind == nil {
		return fields
	}

	for _, fieldName := range o.FieldNames {
		objectIDField := field.
			String(fieldName).
			Optional().
			Comment(fmt.Sprintf("the %v id that owns the object", getName(o.Kind)))

		// if explicitly set to allow empty values, otherwise ensure it is not empty
		if !o.AllowEmptyForSystemAdmin {
			objectIDField.NotEmpty()
		}

		fields = append(fields, objectIDField)
	}

	return fields
}

// Edges of the ObjectOwnedMixin
func (o ObjectOwnedMixin) Edges() []ent.Edge {
	var edges []ent.Edge

	// add the organization owner edge if the flag is set
	if o.IncludeOrganizationOwner {
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
			From("owner", getType(o.Kind)).
			Field(fieldName).
			Ref(o.Ref).
			Unique()

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
		hooks.HookObjectOwnedTuples(o.FieldNames, ownerRelation),
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

			objectIDs, err := interceptors.GetAuthorizedObjectIDs(ctx, strcase.SnakeCase(m.Type()), fgax.CanEdit)
			if err != nil {
				return nil, err
			}

			// filter by object ids on update and delete mutations
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
