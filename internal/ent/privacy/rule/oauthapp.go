package rule

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/common/enums"
	fgamodel "github.com/theopenlane/core/v2/fga/model"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/core/v2/internal/oauthserver"
)

var (
	// ErrOAuthAppReadOnly is returned when an application with read-only access attempts a mutation
	ErrOAuthAppReadOnly = errors.New("this application has read-only access")
	// ErrOAuthAppNotAllowed is returned when an application attempts to change an object its schema protects from applications
	ErrOAuthAppNotAllowed = errors.New("applications are not allowed to modify this object")
	// ErrOAuthAppScopeMissing is returned when a personal connection's scopes don't cover the object type
	ErrOAuthAppScopeMissing = fmt.Errorf("%w: this application was not granted access to this object", ErrRequiredScopeNotSet)
)

// DenyIfReadOnlyOAuthApp denies every mutation made by an installed application granted read-only access
func DenyIfReadOnlyOAuthApp() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {
		if g, ok := oauthserver.CallerGrantFromContext(ctx); ok && g.Access != enums.OAuthAccessLevelWrite {
			return privacy.Denyf("%w", ErrOAuthAppReadOnly)
		}

		return privacy.Skip
	})
}

// DenyIfOAuthAppMissingMutationScope denies a personal connection's mutation unless its scopes cover the object type
func DenyIfOAuthAppMissingMutationScope() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		g, ok := oauthserver.CallerGrantFromContext(ctx)
		if !ok || !g.WriteLimited() {
			return privacy.Skip
		}

		op := m.Op()

		return checkOAuthAppScope(ctx, g, scopedRelation(strings.TrimSuffix(m.Type(), "History"), "", &op), true)
	})
}

// DenyIfOAuthAppMissingReadScope denies a personal connection's query unless its scopes cover the object type
func DenyIfOAuthAppMissingReadScope() privacy.QueryRule {
	return privacy.QueryRuleFunc(func(ctx context.Context, q generated.Query) error {
		g, ok := oauthserver.CallerGrantFromContext(ctx)
		if !ok || !g.ReadLimited() {
			return privacy.Skip
		}

		query, err := intercept.NewQuery(q)
		if err != nil {
			return err
		}

		objectType := strings.TrimSuffix(GetFGAObjectType(query), "History")

		// organizations are needed for every request, like CheckSubjectScope
		if objectType == generated.TypeOrganization {
			return privacy.Skip
		}

		return checkOAuthAppScope(ctx, g, scopedRelation(objectType, fgax.CanView, nil), false)
	})
}

// checkOAuthAppScope denies unless the grant's service subject holds relation on the organization
// a relation outside the scope model denies only when denyUnscoped is set
func checkOAuthAppScope(ctx context.Context, g oauthserver.CallerGrant, relation string, denyUnscoped bool) error {
	scopeSet, err := fgamodel.DefaultServiceScopeSet()
	if err != nil {
		return err
	}

	if _, ok := scopeSet[relation]; !ok {
		if denyUnscoped {
			return privacy.Denyf("%w: %s", ErrOAuthAppScopeMissing, relation)
		}

		return privacy.Skip
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	allowed, err := utils.CheckAccess(ctx, utils.AuthzClientFromContext(ctx), fgax.AccessCheck{
		SubjectID:   g.ID,
		SubjectType: auth.ServiceSubjectType,
		Relation:    relation,
		ObjectType:  generated.TypeOrganization,
		ObjectID:    orgID,
	})
	if err != nil {
		return err
	}

	if !allowed {
		return privacy.Denyf("%w: %s", ErrOAuthAppScopeMissing, relation)
	}

	return privacy.Skip
}

// DenyIfOAuthApp denies mutations made by any installed application, for schemas marked with policy.WithDenyAppMutations
func DenyIfOAuthApp() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {
		if auth.IsInternalRequest(ctx) {
			return privacy.Skip
		}

		if _, ok := oauthserver.CallerGrantFromContext(ctx); ok {
			return privacy.Denyf("%w", ErrOAuthAppNotAllowed)
		}

		return privacy.Skip
	})
}
