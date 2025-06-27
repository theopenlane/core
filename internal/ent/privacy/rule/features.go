package rule

import (
	"context"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/catalog/features"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// HasFeature reports whether the current organization has the given feature enabled
func HasFeature(ctx context.Context, feature string) (bool, error) {
	feats, err := orgFeatures(ctx)
	if err != nil {
		return false, err
	}

	if slices.Contains(feats, feature) {
		return true, nil
	}

	return false, nil
}

// orgFeatures returns the enabled features for the authenticated organization
func orgFeatures(ctx context.Context) ([]string, error) {
	// if the entitlements service is disabled skip feature checks
	if client := generated.FromContext(ctx); client != nil {
		if client.EntitlementManager == nil {
			return nil, nil
		}
	} else if tx := transaction.FromContext(ctx); tx != nil {
		if txClient := tx.Client(); txClient != nil && txClient.EntitlementManager == nil {
			return nil, nil
		}
	}

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil || au.OrganizationID == "" {
		return nil, nil
	}

	if cache, ok := features.CacheFromContext(ctx); ok {
		feats, err := cache.Get(ctx, au.OrganizationID)
		if err != nil {
			log.Err(err).Msg("failed to get feature cache")
		} else if len(feats) > 0 {
			return feats, nil
		}
	}

	var feats []string

	ac := utils.AuthzClientFromContext(ctx)

	if ac == nil {
		client := generated.FromContext(ctx)

		if client == nil {
			return nil, nil
		}

		modules, err := client.OrgModule.Query().Select(orgmodule.FieldModule).
			Where(orgmodule.OwnerID(au.OrganizationID), orgmodule.Active(true)).All(ctx)
		if err != nil {
			return nil, err
		}

		feats = make([]string, 0, len(modules))

		for _, m := range modules {
			feats = append(feats, m.Module)
		}
	} else {
		req := fgax.ListRequest{
			SubjectID:   au.OrganizationID,
			SubjectType: generated.TypeOrganization,
			ObjectType:  "feature",
			Relation:    "enabled",
		}

		resp, err := ac.ListObjectsRequest(ctx, req)
		if err != nil {
			return nil, err
		}

		feats = make([]string, 0, len(resp.Objects))

		for _, obj := range resp.Objects {
			ent, parseErr := fgax.ParseEntity(obj)
			if parseErr != nil {
				continue
			}

			feats = append(feats, ent.Identifier)
		}
	}

	if cache, ok := features.CacheFromContext(ctx); ok {
		if err := cache.Set(ctx, au.OrganizationID, feats); err != nil {
			log.Err(err).Msg("failed to set feature cache")
		}
	}

	return feats, nil
}

// AllowIfHasFeature is a privacy rule allowing the operation if the feature is enabled
// this is intentionally generic
func AllowIfHasFeature(feature string) privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ok, err := HasFeature(ctx, feature)
		if err != nil {
			return err
		}

		if ok {
			return privacy.Allow
		}

		return privacy.Denyf("feature %s not enabled", feature)
	})
}

// HasAnyFeature checks if any of the provided features are enabled for the organization
func HasAnyFeature(ctx context.Context, feats ...string) (bool, error) {
	enabled, err := orgFeatures(ctx)
	if err != nil {
		return false, err
	}

	enabledSet := make(map[string]struct{}, len(enabled))

	for _, f := range enabled {
		enabledSet[f] = struct{}{}
	}

	for _, f := range feats {
		if _, ok := enabledSet[f]; ok {
			return true, nil
		}
	}

	return false, nil
}

// AllowIfHasAnyFeature allows the operation if any of the provided features are enabled
func AllowIfHasAnyFeature(features ...string) privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ok, err := HasAnyFeature(ctx, features...)
		if err != nil {
			return err
		}

		if ok {
			return privacy.Allow
		}

		return privacy.Denyf("none of the features %v are enabled", features)
	})
}
