package rule

import (
	"context"
	"fmt"

	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

// CheckGroupBasedObjectCreationAccess is a rule that returns allow decision if user has
// access to create the given object in the organization
func CheckGroupBasedObjectCreationAccess() privacy.MutationRuleFunc {
	return privacy.MutationRuleFunc(func(ctx context.Context, m generated.Mutation) error {
		if m.Op() != generated.OpCreate {
			return privacy.Skipf("mutation is not a create operation, skipping")
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil || caller.IsAnonymous() {
			logx.FromContext(ctx).Info().Msg("unable to get caller from context")

			return auth.ErrNoAuthUser
		}

		if caller.OrganizationID == "" {
			return privacy.Skipf("no organization set on request, skipping")
		}

		// get the relation, which will be can_create_<object_type>
		relation := fmt.Sprintf("can_create_%s", strcase.SnakeCase(m.Type()))

		ac := fgax.AccessCheck{
			SubjectID:   caller.SubjectID,
			SubjectType: caller.SubjectType(),
			ObjectID:    caller.OrganizationID,
			ObjectType:  generated.TypeOrganization,
			Relation:    relation,
			Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
		}

		access, err := utils.AuthzClientFromContext(ctx).CheckAccess(ctx, ac)
		if err != nil {
			logx.FromContext(ctx).Err(err).Interface("req", ac).Msg("failed to check access")

			return generated.ErrPermissionDenied
		}

		if !access {
			// deny if the user does not have access to create the object
			logx.FromContext(ctx).Error().Str("relation", relation).Str("organization_id", caller.OrganizationID).Str("user_id", caller.SubjectID).Str("email", caller.SubjectEmail).Msg("access denied to create object in organization")

			return generated.ErrPermissionDenied
		}

		// if we reach here, user has access to create the object in the organization
		return privacy.Allow
	})
}
