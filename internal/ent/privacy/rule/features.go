package rule

import (
	"context"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/orgmodule"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/catalog/features"
)

// HasFeature reports whether the current organization has the given feature enabled.
func HasFeature(ctx context.Context, feature string) (bool, error) {
	au, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil || au.OrganizationID == "" {
		return false, nil
	}
	var feats []string
	if cache, ok := features.CacheFromContext(ctx); ok {
		feats, err = cache.Get(ctx, au.OrganizationID)
		if err != nil {
			return false, err
		}
	}
	if len(feats) == 0 {
		ac := utils.AuthzClientFromContext(ctx)
		if ac == nil {
			// fallback to database if no authz client
			client := generated.FromContext(ctx)
			if client == nil {
				return false, nil
			}
			modules, err := client.OrgModule.Query().
				Select(orgmodule.FieldModule).
				Where(orgmodule.OwnerID(au.OrganizationID), orgmodule.Active(true)).
				All(ctx)
			if err != nil {
				return false, err
			}
			feats = make([]string, 0, len(modules))
			for _, m := range modules {
				feats = append(feats, m.Module)
			}
		} else {
			req := fgax.ListRequest{
				SubjectID:   au.OrganizationID,
				SubjectType: organization.Table,
				ObjectType:  "feature",
				Relation:    "enabled",
			}
			resp, err := ac.ListObjectsRequest(ctx, req)
			if err != nil {
				return false, err
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
			_ = cache.Set(ctx, au.OrganizationID, feats)
		}
	}

	for _, f := range feats {
		if f == feature {
			return true, nil
		}
	}
	return false, nil
}

// AllowIfHasFeature is a privacy rule allowing the operation if the feature is enabled.
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

// HasAnyFeature reports whether the organization has at least one of the given features enabled.
func HasAnyFeature(ctx context.Context, features ...string) (bool, error) {
	for _, f := range features {
		ok, err := HasFeature(ctx, f)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// AllowIfHasAnyFeature allows the operation if any of the provided features are enabled.
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
