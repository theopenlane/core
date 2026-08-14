package graphapi

import (
	"context"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/asset"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const integrationBatchSize = 50

type recommendationSource struct {
	recommendation *model.IntegrationRecommendation
	value          string
}

func (r *Resolver) recommendedIntegrations(ctx context.Context) ([]*model.IntegrationRecommendation, error) {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return nil, rout.ErrPermissionDenied
	}

	rt := r.integrationsRuntime

	if rt == nil {
		return []*model.IntegrationRecommendation{}, nil
	}

	ctx, err := common.SetOrganizationInAuthContext(ctx, &caller.OrganizationID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")
		return nil, rout.ErrPermissionDenied
	}

	client := withTransactionalMutation(ctx)

	installed, err := getInstalledIntegrations(ctx, client)
	if err != nil {
		return nil, err
	}

	recs := buildRecommendationsFromIntegrations(rt.Catalog(), installed)

	var sources = []types.RecommendationSource{
		types.RecommendationSignalSourceSSOProvider,
		types.RecommendationSignalSourceAsset,
		types.RecommendationSignalSourceVendor,
		types.RecommendationSignalSourceSignInProvider,
	}

	for _, source := range sources {
		if err := buildIntegrationRecommendations(ctx, client, recs, caller.OrganizationID, source); err != nil {
			return nil, err
		}
	}

	return buildIntegrationResults(recs), nil
}

func buildRecommendationsFromIntegrations(catalog []types.DefinitionSpec, installed map[string]struct{}) map[string]*model.IntegrationRecommendation {
	out := make(map[string]*model.IntegrationRecommendation, len(catalog))
	for _, def := range catalog {
		if !def.Active || !def.Visible {
			continue
		}

		if _, ok := installed[def.ID]; ok {
			continue
		}

		out[def.ID] = &model.IntegrationRecommendation{
			ID:                    def.ID,
			Family:                def.Family,
			DisplayName:           def.DisplayName,
			Description:           def.Description,
			Category:              def.Category,
			DocsURL:               def.DocsURL,
			LogoURL:               def.LogoURL,
			Tags:                  def.Tags,
			Active:                def.Active,
			Visible:               def.Visible,
			RecommendationSignals: def.RecommendationSignals,
		}
	}

	return out
}

func getInstalledIntegrations(ctx context.Context, client *generated.Client) (map[string]struct{}, error) {
	installations, err := client.Integration.Query().
		All(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "integration"})
	}

	out := make(map[string]struct{}, len(installations))
	for _, installation := range installations {
		if installation.DefinitionID != "" {
			out[installation.DefinitionID] = struct{}{}
		}
	}

	return out, nil
}

func buildIntegrationRecommendations(ctx context.Context, client *generated.Client, recommendations map[string]*model.IntegrationRecommendation, ownerID string, source types.RecommendationSource) error {
	switch source {

	case types.RecommendationSignalSourceSSOProvider:
		return buildRecommendationsFromSSO(ctx, client, recommendations, ownerID)

	case types.RecommendationSignalSourceAsset:
		return buildRecommendationsFromAssets(ctx, client, recommendations)

	case types.RecommendationSignalSourceVendor:
		return buildIntegrationsFromVendors(ctx, client, recommendations)

	case types.RecommendationSignalSourceSignInProvider:
		return buildIntegrationsFromAuth(ctx, client, recommendations, ownerID)

	default:
		return nil
	}
}

func buildRecommendationsFromSSO(ctx context.Context, client *generated.Client, recommendations map[string]*model.IntegrationRecommendation, ownerID string) error {
	activities := getRecommendationSources(recommendations, types.RecommendationSignalSourceSSOProvider)

	setting, err := client.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationIDEQ(ownerID)).
		First(ctx)
	if err != nil {
		return parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "organization setting"})
	}

	value := strings.ToLower(strings.TrimSpace(setting.IdentityProvider.String()))
	if value == "" {
		return nil
	}

	for _, activity := range activities {
		if !strings.EqualFold(activity.value, value) {
			continue
		}

		trackIntegrationScore(activity.recommendation, types.RecommendationSignalSourceSSOProvider, 70)
	}

	return nil
}

func buildRecommendationsFromAssets(ctx context.Context, client *generated.Client, recommendations map[string]*model.IntegrationRecommendation) error {
	activities := getRecommendationSources(recommendations, types.RecommendationSignalSourceAsset)

	assets := []*generated.Asset{}

	for i := 0; ; i += integrationBatchSize {
		values, err := client.Asset.Query().
			Order(asset.ByID()).
			Limit(integrationBatchSize).
			Offset(i).
			All(ctx)
		if err != nil {
			return parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "asset"})
		}

		assets = append(assets, values...)

		if len(values) < integrationBatchSize {
			break
		}
	}

	for _, asset := range assets {
		text := formatRecommendationLabel(
			asset.Name,
			asset.DisplayName,
		)
		if text == "" {
			continue
		}

		for _, activity := range activities {
			if !strings.Contains(text, activity.value) {
				continue
			}

			trackIntegrationScore(activity.recommendation, types.RecommendationSignalSourceAsset, 35)
		}
	}

	return nil
}

func buildIntegrationsFromVendors(ctx context.Context, client *generated.Client, recommendations map[string]*model.IntegrationRecommendation) error {
	activities := getRecommendationSources(recommendations, types.RecommendationSignalSourceVendor)

	entities := []*generated.Entity{}

	for i := 0; ; i += integrationBatchSize {
		values, err := client.Entity.Query().
			Order(entity.ByID()).
			Limit(integrationBatchSize).
			Offset(i).
			All(ctx)
		if err != nil {
			return parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "vendor"})
		}

		entities = append(entities, values...)

		if len(values) < integrationBatchSize {
			break
		}
	}

	for _, entity := range entities {
		text := formatRecommendationLabel(
			entity.Name,
			entity.DisplayName,
		)
		if text == "" {
			continue
		}

		for _, activity := range activities {
			if !strings.Contains(text, activity.value) {
				continue
			}

			trackIntegrationScore(activity.recommendation, types.RecommendationSignalSourceVendor, 50)
		}
	}

	return nil
}

// buildIntegrationsFromAuth uses the auth from org members to check if we have an existing intgration for that to
// recommend
func buildIntegrationsFromAuth(ctx context.Context, client *generated.Client, recommendations map[string]*model.IntegrationRecommendation, ownerID string) error {
	signals := getRecommendationSources(recommendations, types.RecommendationSignalSourceSignInProvider)

	authProviders := lo.Uniq(lo.FilterMap(signals, func(signal recommendationSource, _ int) (enums.AuthProvider, bool) {
		provider := enums.ToAuthProvider(signal.value)
		if provider == nil || *provider == enums.AuthProviderInvalid {
			return enums.AuthProviderInvalid, false
		}

		return *provider, true
	}))

	slices.SortFunc(authProviders, func(a, b enums.AuthProvider) int {
		return strings.Compare(string(a), string(b))
	})

	providers := []struct {
		provider enums.AuthProvider
		count    int
	}{}

	for _, provider := range authProviders {
		count, err := client.OrgMembership.Query().
			Where(
				orgmembership.OrganizationIDEQ(ownerID),
				orgmembership.HasUserWith(user.Or(
					user.LastLoginProviderEQ(provider),
					user.AuthProviderEQ(provider),
				)),
			).
			Count(ctx)
		if err != nil {
			return parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "organization membership"})
		}
		if count == 0 {
			continue
		}

		providers = append(providers, struct {
			provider enums.AuthProvider
			count    int
		}{
			provider: provider,
			count:    count,
		})
	}

	for _, providerCount := range providers {
		value := strings.ToLower(strings.TrimSpace(providerCount.provider.String()))
		if value == "" {
			continue
		}

		for _, signal := range signals {
			if !strings.EqualFold(signal.value, value) {
				continue
			}

			trackIntegrationScore(signal.recommendation, types.RecommendationSignalSourceSignInProvider, 25*providerCount.count)
		}
	}

	return nil
}

func getRecommendationSources(recommendations map[string]*model.IntegrationRecommendation, src types.RecommendationSource) []recommendationSource {
	return lo.FlatMap(lo.Values(recommendations), func(recommendation *model.IntegrationRecommendation, _ int) []recommendationSource {

		return lo.FlatMap(recommendation.RecommendationSignals, func(signal types.RecommendationSignal, _ int) []recommendationSource {
			if signal.Source != src {
				return nil
			}

			return lo.FilterMap(signal.Values, func(value string, _ int) (recommendationSource, bool) {
				value = strings.ToLower(strings.TrimSpace(value))
				if value == "" {
					return recommendationSource{}, false
				}

				return recommendationSource{
					recommendation: recommendation,
					value:          value,
				}, true
			})
		})
	})
}

func trackIntegrationScore(rec *model.IntegrationRecommendation, src types.RecommendationSource, score int) {
	rec.Score += score
	if rec.Label != "" {
		return
	}

	switch src {
	case types.RecommendationSignalSourceAsset:
		rec.Label = "Recommended based on matching assets"
	case types.RecommendationSignalSourceVendor:
		rec.Label = "Recommended based on matching vendors"
	case types.RecommendationSignalSourceSSOProvider:
		rec.Label = "Recommended based on your SSO provider"
	case types.RecommendationSignalSourceSignInProvider:
		rec.Label = "Recommended based on sign-in activity"
	default:
		rec.Label = "Recommended based on organization activity"
	}
}

func formatRecommendationLabel(strs ...string) string {
	var s strings.Builder

	for _, str := range strs {
		str = strings.TrimSpace(str)
		if str == "" {
			continue
		}

		s.WriteString(" ")
		s.WriteString(strings.ToLower(str))
	}

	return s.String()
}

func buildIntegrationResults(values map[string]*model.IntegrationRecommendation) []*model.IntegrationRecommendation {
	results := lo.FilterMap(lo.Values(values), func(recommendation *model.IntegrationRecommendation, _ int) (*model.IntegrationRecommendation, bool) {
		if recommendation.Score == 0 {
			return nil, false
		}

		recommendation.Score = min(recommendation.Score, 100)

		return recommendation, true
	})

	slices.SortFunc(results, func(a, b *model.IntegrationRecommendation) int {
		if a.Score == b.Score {
			return strings.Compare(a.DisplayName, b.DisplayName)
		}

		return b.Score - a.Score
	})

	return results
}
