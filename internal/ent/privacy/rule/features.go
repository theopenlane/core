package rule

import (
	"context"
	"fmt"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/middleware/transaction"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/permissioncache"
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
	var client *generated.Client

	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil || au.OrganizationID == "" {
		return nil, nil
	}

	// try feature cache first
	if cache, ok := permissioncache.CacheFromContext(ctx); ok {
		feats, err := cache.GetFeatures(ctx, au.OrganizationID)
		if err != nil {
			log.Err(err).Msg("failed to get feature cache")
		} else if len(feats) > 0 {
			return feats, nil
		}
	}

	if c := generated.FromContext(ctx); c != nil {
		client = c
	} else if tx := transaction.FromContext(ctx); tx != nil {
		client = tx.Client()
	}

	var feats []string

	// attempt to use the EntitlementManager
	if client != nil && client.EntitlementManager != nil {
		ac := utils.AuthzClientFromContext(ctx)
		if ac != nil {
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
	}

	// if EntitlementManager was not usable or features still empty,
	// try to fallback to the OrgModule
	if len(feats) == 0 && client != nil {
		modules, err := client.OrgModule.Query().
			Select(orgmodule.FieldModule).
			Where(
				orgmodule.OwnerID(au.OrganizationID),
				orgmodule.Active(true),
			).All(ctx)

		if err != nil {
			return nil, err
		}

		feats = make([]string, 0, len(modules))
		for _, m := range modules {
			feats = append(feats, m.Module)
		}
	}

	// make sure to cache the result
	if cache, ok := permissioncache.CacheFromContext(ctx); ok {
		if err := cache.SetFeatures(ctx, au.OrganizationID, feats); err != nil {
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
func HasAnyFeature(ctx context.Context, feats ...models.OrgModule) (bool, error) {
	return checkFeatures(ctx, false, feats...)
}

// HasAllFeatures checks if all of the provided features are enabled for the organization
func HasAllFeatures(ctx context.Context, feats ...models.OrgModule) (bool, error) {
	return checkFeatures(ctx, true, feats...)
}

// checkFeatures is a utility function that checks features based on the requireAll flag
// If requireAll is true, all features must be enabled.
//
// If false, at least one must be enabled.
func checkFeatures(ctx context.Context, requireAll bool, feats ...models.OrgModule) (bool, error) {
	enabled, err := orgFeatures(ctx)
	if err != nil {
		return false, err
	}

	enabledSet := make(map[string]struct{}, len(enabled))

	fmt.Println(enabled)

	for _, f := range enabled {
		enabledSet[f] = struct{}{}
	}

	if requireAll {
		// all features must be enabled
		for _, f := range feats {
			if _, ok := enabledSet[string(f)]; !ok {
				return false, nil
			}
		}
		return true, nil
	}

	// at least one feature must be enabled
	for _, f := range feats {
		if _, ok := enabledSet[string(f)]; ok {
			return true, nil
		}
	}

	return false, nil
}

// AllowIfHasAnyFeature allows the operation if any of the provided features are enabled
func AllowIfHasAnyFeature(features ...models.OrgModule) privacy.QueryMutationRule {
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

// AllowIfHasAllFeatures allows the operation if all of the provided features are enabled
func AllowIfHasAllFeatures(features ...models.OrgModule) privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ok, err := HasAllFeatures(ctx, features...)
		if err != nil {
			return err
		}

		if ok {
			return privacy.Allow
		}

		return privacy.Denyf("not all features %v are enabled", features)
	})
}

// DenyIfMissingAllFeatures acts as a prerequisite check - denies if features missing, skips if present
func DenyIfMissingAllFeatures(features ...models.OrgModule) privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ok, err := HasAllFeatures(ctx, features...)
		if err != nil {
			return err
		}

		if !ok {
			return privacy.Denyf("features are not enabled", features)
		}

		return privacy.Skip
	})
}
