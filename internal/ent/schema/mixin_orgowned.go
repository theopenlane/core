package schema

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

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

// createOrgOwnerParentTuple creates the tuple for the parent org owner relationship
func createOrgOwnerParentTuple(ctx context.Context, m ent.Mutation, objectID string) ([]fgax.TupleKey, error) {
	var addTuples []fgax.TupleKey

	// create the tuple for the parent org owner relationship without the subject id
	// this will be filled in by getTuplesToAdd based on the owner id field
	tr := fgax.TupleRequest{
		SubjectType: generated.TypeOrganization,
		ObjectID:    objectID,                              // this is the object id being created
		ObjectType:  hooks.GetObjectTypeFromEntMutation(m), // this is the object type being created
		Relation:    fgax.ParentContextRelation,
	}

	t, err := hooks.GetTuplesToAdd(ctx, m, tr, ownerFieldName)
	if err != nil {
		return nil, err
	}

	addTuples = append(addTuples, t...)

	return addTuples, nil
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

			// schemas that have an owner_id field but do not have permissions based on the parent_context
			// directly will not create a parent_context tuple
			if o.SkipParentContextTuple {
				return retVal, err
			}

			objectID, err := hooks.GetObjectIDFromEntValue(retVal)
			if err != nil {
				return nil, err
			}

			tuples, err := createOrgOwnerParentTuple(ctx, m, objectID)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to create organization owner parent tuple")

				return nil, err
			}

			if len(tuples) > 0 {
				if _, err := utils.AuthzClient(ctx, m).WriteTupleKeys(ctx, tuples, nil); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to write organization owner parent tuples")

					return nil, ErrInternalServerError
				}
			}

			return retVal, err
		})
	}, ent.OpCreate)
}

// orgHookCreateServiceOnlyFunc is a HookFunc that sets the owner on create mutations
// but does not add the organization as an editor, because for service-only objects, org membership should grant view access but not edit access
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

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate)
}

// setOwnerIDField sets the owner id field on the mutation based on the current organization
func (o ObjectOwnedMixin) setOwnerIDField(ctx context.Context, m ent.Mutation) error {
	caller, ok := auth.CallerFromContext(ctx)
	// skip setting owner if this is a service-level internal operation (e.g. org creation, subscription management)
	// CapBypassFGA distinguishes real service callers from test internal contexts created by
	// rule.WithInternalContext, which adds CapInternalOperation but not CapBypassFGA
	if ok && caller != nil && caller.Has(auth.CapInternalOperation|auth.CapBypassFGA) {
		return nil
	}

	if !ok || caller == nil {
		return fmt.Errorf("failed to get organization id from context: %w", auth.ErrNoAuthUser)
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

// defaultOrgInterceptorFunc is the default interceptor function for the organization owned mixin
// this applies a filter on organization ID for any request to a schema that applies the org
// owned mixin
var defaultOrgInterceptorFunc InterceptorFunc = func(o ObjectOwnedMixin) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if skip := o.orgInterceptorSkipper(ctx); skip {
			return nil
		}

		// check query context skips
		if skipQueryModeCheck(ctx, o.SkipOrgInterceptorType) {
			return nil
		}

		isAnon, err := isAnonCaller(ctx)
		if err != nil {
			return err
		}

		trustCenterOrganization, isTrustCenterAnon, err := isAnonTrustCenterCaller(ctx)
		if err != nil {
			return err
		}

		isQuestionnaireCaller := isQuestionnaireAnonCaller(ctx)

		// TC anon users without GraphQL key are blocked by BlockNonTrustCenterAnonymous middleware;
		// questionnaire callers are legitimate REST callers and fall through to the org filter
		if isAnon && !isTrustCenterAnon && !isQuestionnaireCaller {
			return privacy.Denyf("anonymous access not allowed unless filtered by a trust center")
		}

		// check for anon access on schemas, if its not allowed
		// deny the query, otherwise filter on trust center
		if isTrustCenterAnon {
			if !o.AllowAnonymousTrustCenterAccess {
				return privacy.Denyf("anonymous trust center access not allowed")
			}

			// TC users with an active trust center key: scope results to their specific org
			// to prevent cross-org data leakage through the org filter fallback below
			o.PWithField(q, ownerFieldName, []string{trustCenterOrganization})

			return nil
		}

		// check API Token scope and return error if scope not set on token for object
		if auth.IsAPITokenAuthentication(ctx) {
			if err := rule.CheckSubjectScope(ctx, q.Type(), fgax.CanView, nil); errors.Is(err, rule.ErrRequiredScopeNotSet) {
				return err
			}
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

// isAnonTrustCenterCaller returns true for anon trust center callers
// with the organization id the trust center is associated with from
// the caller context
func isAnonTrustCenterCaller(ctx context.Context) (string, bool, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return "", false, auth.ErrNoAuthUser
	}

	tcID, hasAnonTCUser := auth.ActiveTrustCenterIDKey.Get(ctx)
	if !hasAnonTCUser || tcID == "" {
		if caller.Has(auth.CapTrustCenterAnonymous) {
			return "", false, privacy.Denyf("trust center request without active trust center key")
		}

		return "", false, nil
	}

	return caller.OrganizationID, caller.OrganizationRole == auth.AnonymousRole && (caller.Has(auth.CapTrustCenterAnonymous)), nil
}

// isQuestionnaireAnonCaller returns true for a questionnaire anonymous caller
func isQuestionnaireAnonCaller(ctx context.Context) bool {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		return false
	}

	return caller.OrganizationRole == auth.AnonymousRole && caller.Has(auth.CapQuestionnaireAnonymous)
}

// isAnonCaller returns true for any anonymous caller
func isAnonCaller(ctx context.Context) (bool, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return false, auth.ErrNoAuthUser
	}

	return caller.OrganizationRole == auth.AnonymousRole, nil
}

// orgInterceptorSkipper skips the organization interceptor based on the context
// and query type. Callers with CapBypassOrgFilter skip the interceptor.
func (o ObjectOwnedMixin) orgInterceptorSkipper(ctx context.Context) bool {
	if caller, ok := auth.CallerFromContext(ctx); ok && caller.Has(auth.CapBypassOrgFilter) {
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

	// skip the hook for internal operations (subscription management, acme solver, keystore, etc.)
	// CapBypassFGA distinguishes real service callers from test internal contexts created by
	// rule.WithInternalContext, which adds CapBypassOrgFilter|CapInternalOperation but not CapBypassFGA
	if caller, ok := auth.CallerFromContext(ctx); ok && caller.Has(auth.CapBypassOrgFilter|auth.CapInternalOperation|auth.CapBypassFGA) {
		return true, nil
	}

	skip, err := o.skipOrgHookForAdmins(ctx)
	if err != nil {
		return false, err
	}

	return skip, nil
}
