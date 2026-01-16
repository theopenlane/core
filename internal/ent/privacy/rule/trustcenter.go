package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

// trustCenterMutation defines an interface for mutations that involve a trust center ID.
type trustCenterMutation interface {
	TrustCenterID() (id string, exists bool)
}

// AllowIfTrustCenterEditor checks if the user has edit access to the trust center associated with the mutation
// so it can be used to allow mutations on trust center related entities.
func AllowIfTrustCenterEditor() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logx.FromContext(ctx).Debug().Msg("checking write access for trust center")

		trustCenterID := ""

		tcMutation, ok := m.(trustCenterMutation)
		if ok && tcMutation != nil {
			// check if the user has edit access to the trust center
			id, exists := tcMutation.TrustCenterID()
			if exists {
				trustCenterID = id
			}
		}

		if trustCenterID != "" {
			// else check the organization access, to look up trust center ID
			var err error
			trustCenterID, err = generated.FromContext(ctx).TrustCenter.Query().Where().OnlyID(ctx)
			if err != nil {
				return privacy.Skipf("unable to get trust center ID from context: %v", err)
			}
		}

		if trustCenterID == "" {
			return privacy.Skipf("trust center ID not found in mutation")
		}

		return checkTrustCenterAccess(ctx, fgax.CanEdit, trustCenterID)
	})
}

// checkTrustCenterAccess checks if the authenticated user has the specified relation access to the trust center.
func checkTrustCenterAccess(ctx context.Context, relation string, trustCenterID string) error {
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return err
	}

	ac := fgax.AccessCheck{
		SubjectID:   au.SubjectID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
		Relation:    relation,
		ObjectType:  generated.TypeTrustCenter,
		ObjectID:    trustCenterID,
		Context:     utils.NewOrganizationContextKey(au.SubjectEmail),
	}

	access, err := utils.AuthzClientFromContext(ctx).CheckOrgAccess(ctx, ac)
	if err != nil {
		return err
	}

	if access {
		return privacy.Allow
	}

	return privacy.Skipf("request denied by access for user in trust center")
}
