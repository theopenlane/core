package schema

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	ownerFieldName = "owner_id"
)

type objectOwnedOption func(*ObjectOwnedMixin)

// newOrgOwnedMixin creates a new OrgOwnedMixin using the plural name of the schema
// and all defaults. The schema must implement the SchemaFuncs interface to be used.
// options can be passed to customize the mixin
func newOrgOwnedMixin(schema any, opts ...objectOwnedOption) ObjectOwnedMixin {
	sch := toSchemaFuncs(schema)

	// defaults settings
	o := ObjectOwnedMixin{
		// owner_id field
		FieldNames: []string{ownerFieldName},
		Kind:       Organization{},
		// plural name of the schema because the organization will usually have many of these objects
		Ref:              sch.PluralName(),
		HookFuncs:        []HookFunc{defaultOrgHookFunc},
		InterceptorFuncs: []InterceptorFunc{defaultOrgInterceptorFunc},
		OwnerFieldName:   ownerFieldName,
	}

	// apply options
	for _, opt := range opts {
		opt(&o)
	}

	return o
}

var defaultOrgHookFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			skip, err := o.orgHookSkipper(ctx)
			if err != nil {
				return nil, err
			}

			if skip {
				return next.Mutate(ctx, m)
			}

			// set owner on create mutation
			if m.Op() == ent.OpCreate {
				if err := o.setOwnerIDField(ctx, m); err != nil {
					return nil, err
				}

				return next.Mutate(ctx, m)
			}

			// for other operations, add where filter based on the orgs in the context
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

			o.PWithField(mx, ownerFieldName, orgIDs)

			return next.Mutate(ctx, m)
		})
	}
}

// orgHookCreateFunc is a HookFunc that sets the owner on create mutations
var orgHookCreateFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			skip, err := o.skipOrgHookForAdmins(ctx)
			if err != nil {
				return nil, err
			}

			if skip {
				return next.Mutate(ctx, m)
			}

			// set owner on create mutation
			if err := o.setOwnerIDField(ctx, m); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to set owner id field")

				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// add organization owner editor relation to the object
			id, err := hooks.GetObjectIDFromEntValue(retVal)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get object id from ent value")

				return nil, err
			}

			if err := addOrganizationOwnerEditorRelation(ctx, m, id); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to add organization owner editor relation")

				return nil, err
			}

			return retVal, err
		})
	}, ent.OpCreate)
}

// orgHookCreateServiceOnlyFunc is a HookFunc that sets the owner on create mutations
// and creates a parent relation tuple (not editor). This is for system-driven objects
// where only services should be able to create/edit/delete, but users can view through org membership.
var orgHookCreateServiceOnlyFunc HookFunc = func(o ObjectOwnedMixin) ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			skip, err := o.skipOrgHookForAdmins(ctx)
			if err != nil {
				return nil, err
			}

			if skip {
				return next.Mutate(ctx, m)
			}

			// set owner on create mutation
			if err := o.setOwnerIDField(ctx, m); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to set owner id field")

				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			// add organization owner parent relation to the object (not editor)
			// this allows users to view through org membership
			id, err := hooks.GetObjectIDFromEntValue(retVal)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get object id from ent value")

				return nil, err
			}

			if err := addOrganizationOwnerParentRelation(ctx, m, id); err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to add organization owner parent relation")

				return nil, err
			}

			return retVal, err
		})
	}, ent.OpCreate)
}

// skipUserParentTupleFunc returns a skip function that always skips creating user parent tuples.
// This is used for service-only objects where the parent relation only allows organization/service types.
func skipUserParentTupleFunc(_ context.Context, m ent.Mutation) bool {
	// Always skip creating user permissions for service-only objects
	// because users are not allowed as parent types in the FGA model
	if m.Op() == ent.OpCreate {
		return true
	}

	return false
}

// setOwnerIDField sets the owner id field on the mutation based on the current organization
func (o ObjectOwnedMixin) setOwnerIDField(ctx context.Context, m ent.Mutation) error {
	// if the context has the organization creation context key, skip the hook
	// because we don't want the owner to be based on the current organization
	if _, ok := contextx.From[auth.OrganizationCreationContextKey](ctx); ok {
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
func addOrganizationOwnerEditorRelation(ctx context.Context, m ent.Mutation, id string) (err error) {
	var orgID string

	orgID, err = auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization id from context: %w", err)
	}

	tr := fgax.TupleRequest{
		SubjectType:     generated.TypeOrganization,
		SubjectID:       orgID,
		SubjectRelation: fgax.OwnerRelation,
		ObjectID:        id,                                    // this is the object id being created
		ObjectType:      hooks.GetObjectTypeFromEntMutation(m), // this is the object type being created
		Relation:        fgax.EditorRelation,
	}

	t := fgax.GetTupleKey(tr)

	if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, []fgax.TupleKey{t}, nil); err != nil {
		return err
	}

	return nil
}

// addOrganizationOwnerParentRelation adds the organization as a parent to the object
// This is used for service-only objects where users can view through org membership
// but cannot edit (unlike addOrganizationOwnerEditorRelation which grants editor access)
func addOrganizationOwnerParentRelation(ctx context.Context, m ent.Mutation, id string) (err error) {
	var orgID string

	orgID, err = auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization id from context: %w", err)
	}

	tr := fgax.TupleRequest{
		SubjectType: generated.TypeOrganization,
		SubjectID:   orgID,
		ObjectID:    id,
		ObjectType:  hooks.GetObjectTypeFromEntMutation(m),
		Relation:    fgax.ParentRelation,
	}

	t := fgax.GetTupleKey(tr)

	if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, []fgax.TupleKey{t}, nil); err != nil {
		return err
	}

	return nil
}

// defaultOrgInterceptorFunc is the default interceptor function for the organization owned mixin
// this applies a filter on organization ID for any request to a schema that applies the org
// owned mixin
var defaultOrgInterceptorFunc InterceptorFunc = func(o ObjectOwnedMixin) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if skip := o.orgInterceptorSkipper(ctx, q); skip {
			return nil
		}

		// check query context skips
		if skipQueryModeCheck(ctx, o.SkipOrgInterceptorType) {
			return nil
		}

		anon, hasAnonUser := auth.AnonymousTrustCenterUserFromContext(ctx)

		if o.AllowAnonymousTrustCenterAccess && hasAnonUser {
			if anon.TrustCenterID != "" && anon.OrganizationID != "" {
				o.PWithField(q, ownerFieldName, []string{anon.OrganizationID})
				return nil
			}
		} else if !o.AllowAnonymousTrustCenterAccess && hasAnonUser {
			return privacy.Denyf("anonymous trust center access not allowed")
		}

		// add owner id(s) to the query
		orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
		if err != nil {
			return err
		}

		if len(orgIDs) == 0 {
			logx.FromContext(ctx).Warn().Msg("no organization ids found in context, but interceptor was not skipped, no results will be returned")
		}

		// sets the owner id on the query for the current organization
		o.PWithField(q, ownerFieldName, orgIDs)

		return nil
	})
}

// orgInterceptorSkipper skips the organization interceptor based on the context
// and query type
// if soft deletes are bypassed; so is the interceptor - the user will no longer have access to the organization and
// filters will skip the organization
// if the context has a privacy token type, the interceptor is skipped
// if the context has the managed group key, the interceptor is skipped
// if the query is for a token and explicitly allowed, the interceptor is skipped
// if the context has an internal request key, the interceptor is skipped
// if the user is a system admin and AllowEmptyForSystemAdmin is true, the interceptor is skipped
func (o ObjectOwnedMixin) orgInterceptorSkipper(ctx context.Context, q intercept.Query) bool {
	// allow for tests and internal requests to skip the interceptor
	if rule.IsInternalRequest(ctx) {
		return true
	}

	allow, err := rule.CheckIsSystemAdminWithContext(ctx)
	if err == nil && allow {
		return true
	}

	if entx.CheckSkipSoftDelete(ctx) {
		return true
	}

	// skip the interceptor if the context has the token type
	// this is useful for tokens, where the user is not yet authenticated to
	// a particular organization yet
	if skip := rule.SkipTokenInContext(ctx, o.SkipTokenType); skip {
		return true
	}

	// Allow the interceptor to skip the query if the context has an allow
	// bypass and its for a token
	// these are queried during the auth flow and should not be filtered
	if q.Type() == generated.TypeAPIToken || q.Type() == generated.TypePersonalAccessToken {
		if _, allow := privacy.DecisionFromContext(ctx); allow {
			return true
		}
	}

	// skip the interceptor if the context has the organization creation context key
	// the events need to query the subscription for updates
	if _, orgSubscription := contextx.From[auth.OrgSubscriptionContextKey](ctx); orgSubscription {
		return true
	}

	// skip the interceptor if the context has the acme solver context key
	if _, acmeSolver := contextx.From[auth.AcmeSolverContextKey](ctx); acmeSolver {
		return true
	}

	if _, trustCenterAnonAuth := contextx.From[auth.TrustCenterContextKey](ctx); trustCenterAnonAuth {
		return true
	}

	if _, trustCenterNda := contextx.From[auth.TrustCenterNDAContextKey](ctx); trustCenterNda {
		return true
	}

	if _, questionnaireAnonAuth := contextx.From[auth.QuestionnaireContextKey](ctx); questionnaireAnonAuth {
		return true
	}

	// skip interceptor if the context has the managed group key
	if _, managedGroup := contextx.From[hooks.ManagedContextKey](ctx); managedGroup {
		return true
	}

	if _, keystore := contextx.From[auth.KeyStoreContextKey](ctx); keystore {
		return true
	}

	return false
}

// orgHookSkipper skips the organization hook based on the context
// looking for specific token types or mutations done by system admins
func (o ObjectOwnedMixin) orgHookSkipper(ctx context.Context) (bool, error) {
	// skip the hook if the context has the token type
	// this is useful for tokens, where the user is not yet authenticated to
	// a particular organization yet and auth policy allows this
	if skip := rule.SkipTokenInContext(ctx, o.SkipTokenType); skip {
		return true, nil
	}

	// skip the interceptor if the context has the organization creation context key
	// the events need to query for objects such as api tokens, which are org owned
	if _, orgSubscription := contextx.From[auth.OrgSubscriptionContextKey](ctx); orgSubscription {
		return true, nil
	}

	// skip the interceptor if the context has the acme solver context key
	if _, acmeSolver := contextx.From[auth.AcmeSolverContextKey](ctx); acmeSolver {
		return true, nil
	}

	skip, err := o.skipOrgHookForAdmins(ctx)
	if err != nil {
		return false, err
	}

	return skip, nil
}
