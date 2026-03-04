package rule

import (
	"context"

	"entgo.io/ent"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
)

// trustCenterMutation defines an interface for mutations that involve a trust center ID.
type trustCenterMutation interface {
	TrustCenterID() (string, bool)
	OldTrustCenterID(context.Context) (string, error)
}

// AllowIfTrustCenterEditor checks if the user has edit access to the trust center associated with the mutation
// so it can be used to allow mutations on trust center related entities.
func AllowIfTrustCenterEditor() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logx.FromContext(ctx).Debug().Msg("checking write access for trust center")

		trustCenterID := getTrustCenterIDFromMutation(ctx, m)

		if trustCenterID == "" {
			return privacy.Skipf("trust center ID not found in mutation")
		}

		return checkTrustCenterAccess(ctx, fgax.CanEdit, trustCenterID)
	})
}

// getTrustCenterIDFromMutation extracts the trust center ID from the mutation
// if available, otherwise tries to query it from the database.
func getTrustCenterIDFromMutation(ctx context.Context, m ent.Mutation) string {
	tcMutation, ok := m.(trustCenterMutation)
	if !ok {
		logx.FromContext(ctx).Warn().Str("mutation", m.Type()).Str("rule", "AllowIfTrustCenterEditor").Msg("mutation does not implement trustCenterMutation interface")

		return ""
	}

	// check if the user has edit access to the trust center
	id, exists := tcMutation.TrustCenterID()
	if exists {
		return id
	}

	// try to get the old trust center id if available (for updates)
	id, err := tcMutation.OldTrustCenterID(ctx)
	if err == nil {
		return id
	}

	logx.FromContext(ctx).Debug().Msgf("error getting old trust center id from mutation :%v", err)

	return ""
}

// checkTrustCenterAccess checks if the authenticated user has the specified relation access to the trust center.
func checkTrustCenterAccess(ctx context.Context, relation string, trustCenterID string) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.IsAnonymous() {
		return auth.ErrNoAuthUser
	}

	ac := fgax.AccessCheck{
		SubjectID:   caller.SubjectID,
		SubjectType: caller.SubjectType(),
		Relation:    relation,
		ObjectType:  fgax.Kind(strcase.SnakeCase(generated.TypeTrustCenter)),
		ObjectID:    trustCenterID,
		Context:     utils.NewOrganizationContextKey(caller.SubjectEmail),
	}

	access, err := utils.AuthzClientFromContext(ctx).CheckAccess(ctx, ac)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Interface("ac", ac).Msg("error checking trust center access")
		return err
	}

	if access {
		return privacy.Allow
	}

	return privacy.Skipf("request denied by access for user in trust center")
}
