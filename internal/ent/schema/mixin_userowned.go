package schema

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/entx/accessmap"
)

// UserOwnedMixin defines a mixin for user-owned entities.
type UserOwnedMixin struct {
	mixin.Schema
	// Ref table for the id
	Ref string
	// Optional makes the owner id field not required
	Optional bool
	// SoftDeleteIndex creates a unique index on the owner id field where deleted_at is null
	SoftDeleteIndex bool
	// SkipInterceptor skips the interceptor for that schema for all queries, or specific types,
	// this is useful for tokens, etc
	SkipInterceptor interceptors.SkipMode
	// SkipTokenType skips the traverser or hook if the token type is found in the context
	SkipTokenType []token.PrivacyToken
}

// userOwnedOption is a functional option for the UserOwnedMixin
type userOwnedOption func(*UserOwnedMixin)

// newUserOwnedMixin creates a new UserOwnedMixin using the plural name of the schema
// and defined options provided
func newUserOwnedMixin(schema any, opts ...userOwnedOption) UserOwnedMixin {
	sch := toSchemaFuncs(schema)

	// defaults settings
	u := UserOwnedMixin{
		Ref: sch.PluralName(),
	}

	// apply options
	for _, opt := range opts {
		opt(&u)
	}

	return u
}

// withSkipInterceptor allows to set the skip mode for the interceptor
func withSkipInterceptor(skip interceptors.SkipMode) userOwnedOption {
	return func(u *UserOwnedMixin) {
		u.SkipInterceptor = skip
	}
}

// withSkipTokenTypesObjects allows to set custom token types to skip the traverser or hook
func withSkipTokenTypesUsers(tokens ...token.PrivacyToken) userOwnedOption {
	return func(u *UserOwnedMixin) {
		u.SkipTokenType = tokens
	}
}

// withOptionalUser makes the owner id field optional
func withOptionalUser() userOwnedOption {
	return func(u *UserOwnedMixin) {
		u.Optional = true
	}
}

// withSoftDeleteIndex creates a unique index on the owner id field where deleted_at is null
func withSoftDeleteIndex() userOwnedOption {
	return func(u *UserOwnedMixin) {
		u.SoftDeleteIndex = true
	}
}

// Fields of the UserOwnedMixin
func (userOwned UserOwnedMixin) Fields() []ent.Field {
	ownerIDField := field.String(ownerFieldName).
		Annotations(
			entgql.Skip(),
		).
		Comment("The user id that owns the object")

	if userOwned.Optional {
		ownerIDField.Optional()
	}

	return []ent.Field{
		ownerIDField,
	}
}

// Edges of the UserOwnedMixin
func (userOwned UserOwnedMixin) Edges() []ent.Edge {
	if userOwned.Ref == "" {
		panic(errors.New("ref must be non-empty string")) //nolint:err113
	}

	ownerEdge := edge.
		From("owner", User.Type).
		Field(ownerFieldName).
		Ref(userOwned.Ref).
		Annotations(
			entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			accessmap.EdgeNoAuthCheck(),
		).
		Unique()

	if !userOwned.Optional {
		ownerEdge.Required()
	}

	return []ent.Edge{
		ownerEdge,
	}
}

// Indexes of the UserOwnedMixin
func (userOwned UserOwnedMixin) Indexes() []ent.Index {
	if !userOwned.SoftDeleteIndex {
		return []ent.Index{}
	}

	return []ent.Index{
		index.Fields(ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Hooks of the UserOwnedMixin
func (userOwned UserOwnedMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				// skip hook if strictly set to allow
				if _, allow := privacy.DecisionFromContext(ctx); allow {
					return next.Mutate(ctx, m)
				}

				// skip the hook if the context has the token type
				// this is useful for tokens, where the user is not yet (e.g. email verification tokens)
				if skip := rule.SkipTokenInContext(ctx, userOwned.SkipTokenType); skip {
					return next.Mutate(ctx, m)
				}

				userID, err := auth.GetSubjectIDFromContext(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get user id from context: %w", err)
				}

				// set owner on create mutation
				if m.Op() == ent.OpCreate {
					// set owner on mutation
					if err := m.SetField(ownerFieldName, userID); err != nil {
						return nil, err
					}
				} else {
					// filter by owner on update and delete mutations
					mx, ok := m.(interface {
						SetOp(ent.Op)
						Client() *generated.Client
						WhereP(...func(*sql.Selector))
					})
					if !ok {
						return nil, ErrUnexpectedMutationType
					}

					userOwned.P(mx, userID)
				}

				return next.Mutate(ctx, m)
			})
		},
	}
}

// Interceptors of the UserOwnedMixin
func (userOwned UserOwnedMixin) Interceptors() []ent.Interceptor {
	if userOwned.Optional {
		// do not add interceptors if the field is optional
		return []ent.Interceptor{}
	}

	return []ent.Interceptor{
		intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
			// Skip the interceptor for all queries if BypassInterceptor flag is set
			// This is needed for schemas that are never authorized users such as email verification tokens
			if userOwned.SkipInterceptor == interceptors.SkipAll {
				return nil
			}

			userID, err := auth.GetSubjectIDFromContext(ctx)
			if err != nil {
				ctxQuery := ent.QueryFromContext(ctx)

				// Skip the interceptor if the query is for a single entity
				// and the BypassInterceptor flag is set for Only queries
				if userOwned.SkipInterceptor == interceptors.SkipOnlyQuery && ctxQuery.Op == interceptors.OnlyOperation {
					return nil
				}

				return err
			}

			// sets the owner id on the query for the current user
			userOwned.P(q, userID)

			return nil
		}),
	}
}

// P adds a storage-level predicate to the queries and mutations.
func (userOwned UserOwnedMixin) P(w interface{ WhereP(...func(*sql.Selector)) }, userID string) {
	w.WhereP(
		sql.FieldEQ(ownerFieldName, userID),
	)
}
