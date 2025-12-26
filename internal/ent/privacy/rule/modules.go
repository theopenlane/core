package rule

import (
	"context"
	"slices"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/contextx"

	features "github.com/theopenlane/core/internal/entitlements/features"

	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/entitlements"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/permissioncache"
)

// HasFeature reports whether the current organization has the given feature enabled
func HasFeature(ctx context.Context, feature string) (bool, error) {
	// all organizations have access to the base module
	if feature == models.CatalogBaseModule.String() {
		return true, nil
	}

	feats, err := GetOrgFeatures(ctx)
	if err != nil {
		return false, err
	}

	if slices.Contains(feats, feature) {
		return true, nil
	}

	return false, nil
}

// GetFeaturesForSpecificOrganization returns the enabled features for a specific organization
func GetFeaturesForSpecificOrganization(ctx context.Context, orgID string) ([]string, error) {
	// try feature cache first
	if cache, ok := permissioncache.CacheFromContext(ctx); ok {
		moduleFeats, err := cache.GetFeatures(ctx, orgID)
		if err != nil {
			logx.FromContext(ctx).Err(err).Msg("failed to get feature cache")
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
		SubjectID:   orgID,
		SubjectType: strings.ToLower(generated.TypeOrganization),
		ObjectType:  entitlements.TupleObjectType,
		Relation:    entitlements.TupleRelation,
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

		if err := cache.SetFeatures(ctx, orgID, moduleFeats); err != nil {
			logx.FromContext(ctx).Err(err).Msg("failed to set feature cache")
		}
	}

	return feats, nil
}

// GetOrgFeatures returns the enabled features for the authenticated organization
func GetOrgFeatures(ctx context.Context) ([]string, error) {
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		// this intentionally returns nil for the error
		// this is so requests that aren't yet authenticated, but only require the base module
		// e.g. sso login, will continue
		return nil, nil
	}

	// if there is only one authorized org on the pat, set it as the authorized organization
	// more organization require using the X-Organization-ID header
	if au.OrganizationID == "" && len(au.OrganizationIDs) == 1 {
		au.OrganizationID = au.OrganizationIDs[0]
	}

	return GetFeaturesForSpecificOrganization(ctx, au.OrganizationID)
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
func HasAnyFeature(ctx context.Context, feats ...models.OrgModule) (bool, *models.OrgModule, error) {
	return checkFeatures(ctx, false, feats...)
}

// HasAllFeatures checks if all of the provided features are enabled for the organization
func HasAllFeatures(ctx context.Context, feats ...models.OrgModule) (bool, *models.OrgModule, error) {
	return checkFeatures(ctx, true, feats...)
}

// checkFeatures is a utility function that checks features based on the requireAll flag
// If requireAll is true, all features must be enabled.
//
// If false, at least one must be enabled.
func checkFeatures(ctx context.Context, requireAll bool, modules ...models.OrgModule) (bool, *models.OrgModule, error) {
	enabled, err := GetOrgFeatures(ctx)
	if err != nil {
		return false, nil, err
	}

	enabledSet := utils.SliceToMap(enabled)

	if requireAll {
		// all features must be enabled
		for _, f := range modules {
			if f == models.CatalogBaseModule {
				continue
			}

			if _, ok := enabledSet[f.String()]; !ok {
				return false, &f, nil
			}
		}

		return true, nil, nil
	}

	// at least one feature must be enabled
	for _, f := range modules {
		if f == models.CatalogBaseModule {
			return true, nil, nil
		}

		if _, ok := enabledSet[f.String()]; ok {
			return true, nil, nil
		}
	}

	// return the first feature by default
	return false, &modules[0], nil
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

	if w := token.WebauthnCreationContextKeyFromContext(ctx); w != nil {
		return true
	}

	// bypass module checks on trust center users
	if _, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
		return true
	}

	if _, ok := auth.AnonymousQuestionnaireUserFromContext(ctx); ok {
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
			if !utils.ModulesEnabled(mut.Client()) {
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

		ok, _, err := HasAnyFeature(ctx, schemaFeatures...)
		if err != nil {
			return err
		}

		if !ok {
			return privacy.Denyf("features are not enabled")
		}

		return privacy.Skip
	})
}
