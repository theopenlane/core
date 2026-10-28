package rule

import (
	"context"
	"errors"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
	"github.com/theopenlane/logx"
)

// AllowIfTrustCenterEditor checks if the user has edit access to the trust center associated with the mutation
// so it can be used to allow mutations on trust center related entities.
func AllowIfTrustCenterEditor() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {
		logx.FromContext(ctx).Debug().Msg("checking write access for trust center")

		trustCenterIDs := getTrustCenterIDsFromMutation(ctx, m)

		if len(trustCenterIDs) == 0 {
			return privacy.Skipf("trust center ID not found in mutation")
		}

		// the caller must be able to edit every trust center the object belongs to or is moved to
		for _, trustCenterID := range trustCenterIDs {
			if err := checkTrustCenterAccess(ctx, fgax.CanEdit, trustCenterID); !errors.Is(err, privacy.Allow) {
				return err
			}
		}

		return privacy.Allow
	})
}

// AllowIfTrustCenterAnonRequest returns a privacy allow for anon trust center context
func AllowIfTrustCenterAnonRequest() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, _ ent.Mutation) error {
		logx.FromContext(ctx).Debug().Msg("checking for trust center anon context")

		if auth.IsTrustCenterFromContext(ctx) {
			return privacy.Allow
		}

		return privacy.Skip
	})
}

// getTrustCenterIDsFromMutation returns the stored trust center ID for existing objects plus the
// trust center ID set by the mutation, so a mutation cannot name a trust center the object does not belong to
func getTrustCenterIDsFromMutation(ctx context.Context, m ent.Mutation) []string {
	tcMutation, ok := m.(utils.TrustCenterMutation)
	if !ok {
		logx.FromContext(ctx).Warn().Str("mutation", m.Type()).Str("rule", "AllowIfTrustCenterEditor").Msg("mutation does not implement trustCenterMutation interface")

		return nil
	}

	newID, _ := tcMutation.TrustCenterID()

	storedID := ""
	if !m.Op().Is(ent.OpCreate) {
		storedID = getStoredTrustCenterID(ctx, tcMutation, m.Type())
	}

	return lo.Uniq(lo.Compact([]string{storedID, newID}))
}

// getStoredTrustCenterID returns the trust center ID currently stored on the object, querying the database when the old value is unavailable
func getStoredTrustCenterID(ctx context.Context, tcMutation utils.TrustCenterMutation, objectType string) string {
	// try to get the old trust center id if available (for updates), this does not work for delete
	id, err := tcMutation.OldTrustCenterID(ctx)
	if err == nil {
		return id
	}

	objectID, ok := tcMutation.ID()
	if !ok {
		return ""
	}

	pluralObjectType := pluralize.NewClient().Plural(objectType)
	tableName := strcase.SnakeCase(pluralObjectType)

	query := "SELECT trust_center_id FROM " + tableName + " WHERE id = $1"

	var rows sql.Rows
	if err := tcMutation.Client().Driver().Query(ctx, query, []any{objectID}, &rows); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to check for object in organization")

		return ""
	}

	defer rows.Close()

	if rows.Next() {
		var trustCenterID string
		if err := rows.Scan(&trustCenterID); err == nil {
			return trustCenterID
		}
	}

	return ""
}

// checkTrustCenterAccess checks if the authenticated user has the specified relation access to the trust center.
func checkTrustCenterAccess(ctx context.Context, relation string, trustCenterID string) error {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller.IsAnonymous() {
		return auth.ErrNoAuthUser
	}

	// org-scoped support sessions bypass the FGA relation check, only
	// for the trust center owned by the org the session is scoped to
	if caller.Has(auth.CapOrgSupport) {
		return checkTrustCenterSupportAccess(ctx, caller, trustCenterID)
	}

	ac := fgax.AccessCheck{
		SubjectID:   caller.SubjectID,
		SubjectType: caller.SubjectType(),
		Relation:    relation,
		ObjectType:  fgax.Kind(strcase.SnakeCase(generated.TypeTrustCenter)),
		ObjectID:    trustCenterID,
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

// checkTrustCenterSupportAccess allows an org-scoped support session to act on trustCenterID only
// when the trust center actually belongs to the org the session is scoped to; without this check a
// support session for one org could edit another org's trust center just by holding CapOrgSupport
func checkTrustCenterSupportAccess(ctx context.Context, caller *auth.Caller, trustCenterID string) error {
	tc, err := generated.FromContext(ctx).TrustCenter.Get(ctx, trustCenterID)
	if err != nil {
		if generated.IsNotFound(err) {
			return privacy.Skipf("trust center not found")
		}

		return err
	}

	if tc.OwnerID != caller.OrganizationID {
		logx.FromContext(ctx).Info().Str("trust_center_id", trustCenterID).Str("user_id", caller.SubjectID).Msg("support attempting to access trust center outside scoped org")

		return privacy.Denyf("support session not scoped to this trust center's organization")
	}

	return privacy.Allow
}
