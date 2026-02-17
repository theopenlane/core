package rule

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"
	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// scopedRelationForAPIToken returns the scoped relation for an api token based on the object type, relation, and operation
func scopedRelationForAPIToken(objectType string, relation string, op *ent.Op) string {
	object := strcase.SnakeCase(objectType)
	if object == "" {
		return ""
	}

	if op != nil {
		switch {
		case op.Is(ent.OpCreate), op.Is(ent.OpUpdate | ent.OpUpdateOne):
			return fmt.Sprintf("can_edit_%s", object)
		case op.Is(ent.OpDelete | ent.OpDeleteOne):
			return fmt.Sprintf("can_delete_%s", object)
		}
	}

	switch relation {
	case fgax.CanEdit:
		return fmt.Sprintf("can_edit_%s", object)
	case fgax.CanView:
		return fmt.Sprintf("can_view_%s", object)
	case fgax.CanDelete:
		return fmt.Sprintf("can_delete_%s", object)
	default:
		return ""
	}
}

// getFGAObjectType returns the object type for the query
// for membership tables, it will return the type with the membership suffix removed
// e.g. GroupMembership -> Group
func GetFGAObjectType(q intercept.Query) string {
	// Membership tables should use the object_id field,
	// e.g. GroupMembership should use group_id
	isMembership := strings.Contains(q.Type(), "Membership")

	objectType := q.Type()
	if isMembership {
		objectType = strings.ReplaceAll(q.Type(), "Membership", "")
	}

	return objectType
}

// AllowIfTokenHasMutationScope is a rule that allows mutation if the api token has the appropriate scope
// for the object type and operation
// this is used on the base mutation policy to enforce api token scope checks
func AllowIfTokenHasMutationScope() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		objectType := m.Type()
		if objectType == "" {
			return privacy.Skip
		}

		// strip history suffix for history tables
		objectType = strings.TrimSuffix(objectType, "History")

		op := m.Op()
		return CheckAPITokenScope(ctx, objectType, "", &op)
	})
}

// CheckAPITokenScope enforces that the api token has the required scope for the given object type, relation, and operation.
// Returns nil if the rule should be skipped (not an API token or no scoped relation), privacy.Allow if access is granted, or an error if denied
func CheckAPITokenScope(ctx context.Context, objectType string, relation string, op *ent.Op) error {
	if !auth.IsAPITokenAuthentication(ctx) {
		return privacy.Skip
	}

	// allow api token access to api tokens and organizations, as they are needed for for requests
	// filters will be enforced elsewhere
	if objectType == generated.TypeAPIToken || objectType == generated.TypeOrganization {
		return privacy.Allow
	}

	scopedRelation := scopedRelationForAPIToken(objectType, relation, op)
	if scopedRelation == "" {
		return privacy.Skip
	}

	scopeSet, err := fgamodel.DefaultServiceScopeSet()
	if err != nil {
		return err
	}

	if _, ok := scopeSet[scopedRelation]; !ok {
		logx.FromContext(ctx).Error().Str("relation", scopedRelation).Str("object_type", objectType).Msg("invalid scoped relation for api token")

		return fmt.Errorf("%w: invalid scoped relation %s for object type %s", generated.ErrPermissionDenied, scopedRelation, objectType)
	}

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return err
	}

	orgID := au.OrganizationID
	if orgID == "" {
		logx.FromContext(ctx).Error().Str("relation", scopedRelation).Msg("api token missing organization scope")

		return generated.ErrPermissionDenied
	}

	authzClient := utils.AuthzClientFromContext(ctx)
	if authzClient == nil {
		logx.FromContext(ctx).Error().Msg("missing authz client for api token scope check")

		return generated.ErrPermissionDenied
	}

	ac := fgax.AccessCheck{
		SubjectID:   au.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    scopedRelation,
		ObjectType:  generated.TypeOrganization,
		ObjectID:    orgID,
	}

	hasAccess, err := authzClient.CheckAccess(ctx, ac)
	if err != nil {
		logx.FromContext(ctx).Err(err).Interface("check", ac).Msg("failed api token scope check")

		return fmt.Errorf("%w: token not scoped for %s", generated.ErrPermissionDenied, scopedRelation)
	}

	if hasAccess {
		return privacy.Allow
	}

	return generated.ErrPermissionDenied
}
