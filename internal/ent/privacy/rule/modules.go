package rule

import (
	"context"
	"slices"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	features "github.com/theopenlane/core/internal/entitlements/features"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
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
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil || au.OrganizationID == "" {
		return nil, nil
	}

	// try feature cache first
	if cache, ok := permissioncache.CacheFromContext(ctx); ok {
		moduleFeats, err := cache.GetFeatures(ctx, au.OrganizationID)
		if err != nil {
			log.Err(err).Msg("failed to get feature cache")
		} else if len(moduleFeats) > 0 {
			feats := make([]string, 0, len(moduleFeats))
			for _, f := range moduleFeats {
				feats = append(feats, f.String())
			}
			return feats, nil
		}
	}

	ac := utils.AuthzClientFromContext(ctx)
	if ac == nil {
		return nil, nil
	}

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

	feats := make([]string, 0, len(resp.Objects))
	for _, obj := range resp.Objects {
		ent, parseErr := fgax.ParseEntity(obj)
		if parseErr != nil {
			continue
		}
		feats = append(feats, ent.Identifier)
	}

	// make sure to cache the result
	if cache, ok := permissioncache.CacheFromContext(ctx); ok {
		// convert strings to models.OrgModule for caching
		moduleFeats := make([]models.OrgModule, 0, len(feats))
		for _, f := range feats {
			moduleFeats = append(moduleFeats, models.OrgModule(f))
		}
		if err := cache.SetFeatures(ctx, au.OrganizationID, moduleFeats); err != nil {
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
func HasAnyFeature(ctx context.Context, feats ...models.OrgModule) (bool, models.OrgModule, error) {
	return checkFeatures(ctx, false, feats...)
}

// HasAllFeatures checks if all of the provided features are enabled for the organization
func HasAllFeatures(ctx context.Context, feats ...models.OrgModule) (bool, models.OrgModule, error) {
	return checkFeatures(ctx, true, feats...)
}

// checkFeatures is a utility function that checks features based on the requireAll flag
// If requireAll is true, all features must be enabled.
//
// If false, at least one must be enabled.
func checkFeatures(ctx context.Context, requireAll bool, modules ...models.OrgModule) (bool, models.OrgModule, error) {
	enabled, err := orgFeatures(ctx)
	if err != nil {
		return false, models.OrgModule(""), err
	}

	if len(enabled) == 0 {
		return true, models.OrgModule(""), nil
	}

	enabledSet := make(map[string]struct{}, len(enabled))

	for _, f := range enabled {
		enabledSet[f] = struct{}{}
	}

	if requireAll {
		// all features must be enabled
		for _, f := range modules {
			if _, ok := enabledSet[string(f)]; !ok {
				return false, f, nil
			}
		}
		return true, models.OrgModule(""), nil
	}

	// at least one feature must be enabled
	for _, f := range modules {
		if _, ok := enabledSet[string(f)]; ok {
			return true, models.OrgModule(""), nil
		}
	}

	// return the first feature by default
	return false, modules[0], nil
}

// AllowIfHasAnyFeature allows the operation if any of the provided features are enabled
func AllowIfHasAnyFeature(features ...models.OrgModule) privacy.QueryMutationRule {
	return privacy.ContextQueryMutationRule(func(ctx context.Context) error {
		ok, _, err := HasAnyFeature(ctx, features...)
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
		ok, _, err := HasAllFeatures(ctx, features...)
		if err != nil {
			return err
		}

		if ok {
			return privacy.Allow
		}

		return privacy.Denyf("not all features %v are enabled", features)
	})
}

// ShouldSkipFeatureCheck determines if module access checks should be bypassed based
// on the available context
func ShouldSkipFeatureCheck(ctx context.Context) bool {
	if auth.IsSystemAdminFromContext(ctx) {
		return true
	}

	if _, allowCtx := privacy.DecisionFromContext(ctx); allowCtx {
		return true
	}

	if _, ok := contextx.From[auth.OrgSubscriptionContextKey](ctx); ok {
		return true
	}

	if _, ok := contextx.From[auth.OrganizationCreationContextKey](ctx); ok {
		return true
	}

	if w := token.WebauthCreationContextKeyFromContext(ctx); w != nil {
		return true
	}

	skipTokenType := []token.PrivacyToken{
		&token.OauthTooToken{},
		&token.VerifyToken{},
		&token.SignUpToken{},
		&token.OrgInviteToken{},
		&token.ResetToken{},
		&token.JobRunnerRegistrationToken{},
	}

	return SkipTokenInContext(ctx, skipTokenType)
}

// DenyIfMissingAllModules acts as a prerequisite check - denies if features missing, Allows if present
func DenyIfMissingAllModules() privacy.MutationRule {
	return privacy.MutationRuleFunc(func(ctx context.Context, m ent.Mutation) error {

		if mut, ok := m.(interface{ Client() *generated.Client }); ok {
			if client := mut.Client(); client != nil && client.EntConfig != nil && !client.EntConfig.Modules.Enabled {
				return privacy.Skip
			}
		}

		mutationType := m.Type()

		if strings.HasSuffix(mutationType, "History") || ShouldSkipFeatureCheck(ctx) {
			return privacy.Skip
		}

		schemaFeatures, exists := features.FeatureOfType[m.Type()]
		if !exists {
			return privacy.Skip
		}

		ok, _, err := HasAllFeatures(ctx, schemaFeatures...)
		if err != nil {
			return err
		}

		if !ok {
			return privacy.Denyf("features are not enabled")
		}

		return privacy.Skip
	})
}
