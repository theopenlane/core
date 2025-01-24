package schema

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

const (
	ownerFieldName = "owner_id"
)

// NewOrgOwnMixinWithRef creates a new OrgOwnedMixin with the given ref
// and sets the defaults
func NewOrgOwnMixinWithRef(ref string) ObjectOwnedMixin {
	return NewOrgOwnedMixin(
		ObjectOwnedMixin{
			Ref: ref,
		})
}

// NewOrgOwnedMixin creates a new OrgOwnedMixin with the given ObjectOwnedMixin
// and sets the Kind to ownerFieldName and the HookFunc to defaultOrgHookFunc
func NewOrgOwnedMixin(o ObjectOwnedMixin) ObjectOwnedMixin {
	o.FieldNames = []string{ownerFieldName}
	o.Kind = Organization.Type

	if o.HookFuncs == nil {
		o.HookFuncs = []HookFunc{defaultOrgHookFunc}
	}

	if o.InterceptorFunc == nil {
		o.InterceptorFunc = defaultOrgInterceptorFunc
	}

	return o
}

var defaultOrgHookFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
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

			// set owner on create mutation
			if m.Op() == ent.OpCreate {
				if err := setOwnerIDField(ctx, m); err != nil {
					return nil, err
				}
			} else {
				orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to get organization id from context: %w", err)
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

				o.P(mx, orgIDs)
			}

			return next.Mutate(ctx, m)
		})
	}
}

// orgHookCreateFunc is a HookFunc that sets the owner on create mutations
var orgHookCreateFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if m.Op() != ent.OpCreate {
				return next.Mutate(ctx, m)
			}

			// set owner on create mutation
			if m.Op() == ent.OpCreate {
				if err := setOwnerIDField(ctx, m); err != nil {
					log.Error().Err(err).Msg("failed to set owner id field")

					return nil, err
				}
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			id := ""
			if m.Op() == ent.OpCreate {
				// add organization owner editor relation to the object
				id, err = hooks.GetObjectIDFromEntValue(retVal)
				if err != nil {
					log.Error().Err(err).Msg("failed to get object id from ent value")

					return nil, err
				}
			}

			if entx.CheckIsSoftDelete(ctx) || m.Op() == ent.OpDelete || m.Op() == ent.OpDeleteOne {
				id, err = hooks.GetObjectIDFromEntDeleteValue(retVal)
				if err != nil {
					log.Error().Err(err).Msg("failed to get object id from ent delete value")

					return nil, err
				}
			}

			if id == "" {
				return retVal, nil
			}

			if err := addOrganizationOwnerEditorRelation(ctx, m, id); err != nil {
				log.Error().Err(err).Msg("failed to add organization owner editor relation")

				return nil, err
			}

			return retVal, err
		})
	}
}

// setOwnerIDField sets the owner id field on the mutation based on the current organization
func setOwnerIDField(ctx context.Context, m ent.Mutation) error {
	// if the context has the organization creation context key, skip the hook
	// because we don't want the owner to be based on the current organization
	if _, ok := contextx.From[hooks.OrganizationCreationContextKey](ctx); ok {
		return nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization id from context: %w", err)
	}

	// set owner on mutation
	if err := m.SetField(ownerFieldName, orgID); err != nil {
		return err
	}

	return nil
}

// addOrganizationOwnerEditorRelation adds the organization owner as an editor to the object
func addOrganizationOwnerEditorRelation(ctx context.Context, m ent.Mutation, id string) error {
	// always add the organization owner relationship as an editor
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization id from context: %w", err)
	}

	tr := fgax.TupleRequest{
		SubjectType:     "organization",
		SubjectID:       orgID,
		SubjectRelation: fgax.OwnerRelation,
		ObjectID:        id,                                    // this is the object id being created
		ObjectType:      hooks.GetObjectTypeFromEntMutation(m), // this is the object type being created
		Relation:        fgax.EditorRelation,
	}

	t := fgax.GetTupleKey(tr)

	if _, err := utils.AuthzClientFromContext(ctx).WriteTupleKeys(ctx, []fgax.TupleKey{t}, nil); err != nil {
		return err
	}

	return nil
}

var defaultOrgInterceptorFunc InterceptorFunc = func(o ObjectOwnedMixin) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		// skip the interceptor if the context has the token type
		// this is useful for tokens, where the user is not yet authenticated to
		// a particular organization yet
		for _, tokenType := range o.SkipTokenType {
			if rule.ContextHasPrivacyTokenOfType(ctx, tokenType) {
				return nil
			}
		}

		// skip interceptor if the context has the managed group key
		if _, managedGroup := contextx.From[hooks.ManagedContextKey](ctx); managedGroup {
			return nil
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

		// add owner id(s) to the query
		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		// sets the owner id on the query for the current organization
		o.P(q, orgIDs)

		return nil
	})
}
