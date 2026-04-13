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

const (
	CanCreatePrefix = "can_create_"
	CanEditPrefix   = "can_edit_"
	CanDeletePrefix = "can_delete_"
	CanViewPrefix   = "can_view_"
)

// scopedRelation returns the scoped relation based on the object type, relation, and operation. A operation is checked for for create, update, delete. If instead a specific relation should be checked, that should be passed instead of the operation
func scopedRelation(objectType string, relation string, op *ent.Op) string {
	object := strcase.SnakeCase(objectType)
	if object == "" {
		return ""
	}

	if op != nil {
		switch {
		case op.Is(ent.OpCreate):
			return fmt.Sprintf("%s%s", CanCreatePrefix, object)
		case op.Is(ent.OpUpdate | ent.OpUpdateOne):
			return fmt.Sprintf("%s%s", CanEditPrefix, object)
		case op.Is(ent.OpDelete | ent.OpDeleteOne):
			return fmt.Sprintf("%s%s", CanDeletePrefix, object)
		}
	}

	switch relation {
	case fgax.CanEdit:
		return fmt.Sprintf("%s%s", CanEditPrefix, object)
	case fgax.CanView:
		return fmt.Sprintf("%s%s", CanViewPrefix, object)
	case fgax.CanDelete:
		return fmt.Sprintf("%s%s", CanDeletePrefix, object)
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
		return CheckSubjectScope(ctx, objectType, "", &op)
	})
}

// CheckSubjectScope enforces that the authorized subject has the required scope for the given object type, relation, and operation.
// Returns nil if the rule should be skipped (no scoped relation), privacy.Allow if access is granted, or an error if denied
func CheckSubjectScope(ctx context.Context, objectType string, relation string, op *ent.Op) error {
	// allow api token access to api tokens
	if auth.IsAPITokenAuthentication(ctx) && objectType == generated.TypeAPIToken {
		return privacy.Allow
	}

	// allow organizations, as they are needed for for requests
	// filters will be enforced elsewhere
	if objectType == generated.TypeOrganization && relation == fgax.CanView {
		return privacy.Allow
	}

	scopedRelation := scopedRelation(objectType, relation, op)
	if scopedRelation == "" {
		return privacy.Skip
	}

	scopeSet, err := fgamodel.DefaultServiceScopeSet()
	if err != nil {
		return err
	}

	if _, ok := scopeSet[scopedRelation]; !ok {
		logx.FromContext(ctx).Debug().Str("relation", scopedRelation).Str("object_type", objectType).Msg("invalid scoped relation, skipping scope check")

		return privacy.Skip
	}

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		logx.FromContext(ctx).Debug().Msg("unable to get caller from context for scope check, skipping scope check")

		return privacy.Skip
	}

	// this could happen before user is logged in, or if the token is missing organization scope, so we will log and skip the scope check in this case
	orgID := caller.OrganizationID
	if orgID == "" {
		logx.FromContext(ctx).Debug().Str("relation", scopedRelation).Msg("subject missing organization scope, skipping scope check")

		return privacy.Skip
	}

	authzClient := utils.AuthzClientFromContext(ctx)
	if authzClient == nil {
		logx.FromContext(ctx).Error().Msg("missing authz client for scope check")

		return generated.ErrPermissionDenied
	}

	ac := fgax.AccessCheck{
		SubjectID:   caller.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    scopedRelation,
		ObjectType:  generated.TypeOrganization,
		ObjectID:    orgID,
	}

	hasAccess, err := authzClient.CheckAccess(ctx, ac)
	if err != nil {
		logx.FromContext(ctx).Debug().Err(err).Interface("check", ac).Msg("failed scope check, unable to determine access")

		return privacy.Skip
	}

	if hasAccess {
		return privacy.Allow
	}

	if auth.IsAPITokenAuthentication(ctx) {
		logx.FromContext(ctx).Info().Interface("check", ac).Msg("api token missing required scope for access")

		return generated.ErrPermissionDenied
	}

	logx.FromContext(ctx).Debug().Str("required_relation", scopedRelation).Msg("subject not scoped for required relation")

	return privacy.Skip
}
