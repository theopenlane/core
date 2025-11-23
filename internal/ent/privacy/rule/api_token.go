package rule

import (
	"context"
	"fmt"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"
	fgamodel "github.com/theopenlane/core/fga/model"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// scopedRelationForAPIToken returns the scoped relation for an api token based on the object type, relation, and operation
func scopedRelationForAPIToken(objectType string, relation string, op ent.Op) string {
	object := strcase.SnakeCase(objectType)
	if object == "" {
		return ""
	}

	switch {
	case op.Is(ent.OpCreate):
		return fmt.Sprintf("can_edit_%s", object)
	case op.Is(ent.OpDelete | ent.OpDeleteOne):
		return fmt.Sprintf("can_delete_%s", object)
	case op.Is(ent.OpUpdate | ent.OpUpdateOne):
		return fmt.Sprintf("can_edit_%s", object)
	}

	switch relation {
	case fgax.CanEdit:
		return fmt.Sprintf("can_edit_%s", object)
	case fgax.CanView:
		return fmt.Sprintf("can_view_%s", object)
	case fgax.CanDelete:
		return fmt.Sprintf("can_delete_%s", object)
	}

	return ""
}

// CheckAPITokenScope enforces that the api token has the required scope for the given object type, relation, and operation.
// Returns nil if the rule should be skipped (not an API token or no scoped relation), privacy.Allow if access is granted, or an error if denied
func CheckAPITokenScope(ctx context.Context, objectType string, relation string, op ent.Op) error {
	if !auth.IsAPITokenAuthentication(ctx) {
		return nil
	}

	scopedRelation := scopedRelationForAPIToken(objectType, relation, op)
	if scopedRelation == "" {
		return nil
	}

	scopeSet, err := fgamodel.DefaultServiceScopeSet()
	if err != nil {
		return err
	}

	if _, ok := scopeSet[scopedRelation]; !ok {
		return generated.ErrPermissionDenied
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
		return generated.ErrPermissionDenied
	}

	if !hasAccess {
		return generated.ErrPermissionDenied
	}

	return privacy.Allow
}
