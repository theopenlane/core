package integrationrecommendations

import (
	"context"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/asset"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	perBatchScanSize = 50

	ssoProvidersWeight    = 35
	ssoProvidersWeightCap = 35
	assetWeight           = 5
	assetCap              = 15
	vendorWeight          = 15
	vendorCap             = 25
	authProviderWeight    = 5
	authProviderCap       = 25
)

type Recommendation struct {
	DefinitionID string
	Weight       int
	Label        string
}

type candidate struct {
	def    types.DefinitionSpec
	weight int
	label  string
}

type source struct {
	recommendation *candidate
	value          string
}

// Compute looks through the catalog and the organization data to see which integrations should be
// recommended to them to install
func Compute(ctx context.Context, client *generated.Client, catalog []types.DefinitionSpec, ownerID string) ([]Recommendation, error) {
	installed, err := getInstalledIntegrations(ctx, client, ownerID)
	if err != nil {
		return nil, err
	}

	recommendations, signals := getAvaialableCandidates(catalog, installed)
	weightTracker := map[string]map[types.RecommendationSource]int{}

	sources := []types.RecommendationSource{
		types.RecommendationSignalSourceSSOProvider,
		types.RecommendationSignalSourceAsset,
		types.RecommendationSignalSourceVendor,
		types.RecommendationSignalSourceSignInProvider,
	}

	for _, source := range sources {
		if err := calculateWeightForSource(ctx, client, recommendations, signals, weightTracker, ownerID, source); err != nil {
			return nil, err
		}
	}

	out := lo.FilterMap(lo.Values(recommendations), func(recommendation *candidate, _ int) (Recommendation, bool) {
		if recommendation.weight <= 0 {
			return Recommendation{}, false
		}

		return Recommendation{
			DefinitionID: recommendation.def.ID,
			Weight:       recommendation.weight,
			Label:        recommendation.label,
		}, true
	})

	slices.SortFunc(out, func(a, b Recommendation) int {
		if a.Weight == b.Weight {
			return strings.Compare(a.DefinitionID, b.DefinitionID)
		}

		return b.Weight - a.Weight
	})

	return out, nil
}

func getInstalledIntegrations(ctx context.Context, client *generated.Client, ownerID string) (map[string]struct{}, error) {
	integrations, err := client.Integration.Query().
		Where(integration.OwnerIDEQ(ownerID)).
		Select(integration.FieldDefinitionID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[string]struct{}, len(integrations))
	for _, v := range integrations {
		out[v.DefinitionID] = struct{}{}
	}

	return out, nil
}

func getAvaialableCandidates(catalog []types.DefinitionSpec, installed map[string]struct{}) (map[string]*candidate, map[string][]types.RecommendationSignal) {
	out := make(map[string]*candidate, len(catalog))
	signals := map[string][]types.RecommendationSignal{}

	for _, def := range catalog {
		if !def.Active || !def.Visible {
			continue
		}

		if _, ok := installed[def.ID]; ok {
			continue
		}

		out[def.ID] = &candidate{
			def: def,
		}
		signals[def.ID] = def.RecommendationSignals
	}

	return out, signals
}

func calculateWeightForSource(ctx context.Context, client *generated.Client, recommendations map[string]*candidate, signals map[string][]types.RecommendationSignal, weightTracker map[string]map[types.RecommendationSource]int, ownerID string, source types.RecommendationSource) error {
	switch source {
	case types.RecommendationSignalSourceSSOProvider:
		return calculateWeightForSSOProvider(ctx, client, recommendations, signals, weightTracker, ownerID)

	case types.RecommendationSignalSourceAsset:
		return calculateWeightForAssets(ctx, client, recommendations, signals, weightTracker, ownerID)

	case types.RecommendationSignalSourceVendor:
		return calculateWeightForVendors(ctx, client, recommendations, signals, weightTracker, ownerID)

	case types.RecommendationSignalSourceSignInProvider:
		return calculateWeightForAuthProviders(ctx, client, recommendations, signals, weightTracker, ownerID)

	default:
		return nil
	}
}

func calculateWeightForSSOProvider(ctx context.Context, client *generated.Client, recommendations map[string]*candidate, signals map[string][]types.RecommendationSignal, weightTracker map[string]map[types.RecommendationSource]int, ownerID string) error {
	activities := getRecommendationSources(recommendations, signals, types.RecommendationSignalSourceSSOProvider)

	setting, err := client.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationIDEQ(ownerID)).
		First(ctx)
	if generated.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	value := strings.ToLower(strings.TrimSpace(setting.IdentityProvider.String()))
	if value == "" {
		return nil
	}

	for _, activity := range activities {
		if !strings.EqualFold(activity.value, value) {
			continue
		}

		trackWeight(activity.recommendation, weightTracker, types.RecommendationSignalSourceSSOProvider, ssoProvidersWeight, ssoProvidersWeightCap)
	}

	return nil
}

func calculateWeightForAssets(ctx context.Context, client *generated.Client, recommendations map[string]*candidate, signals map[string][]types.RecommendationSignal, weightTracker map[string]map[types.RecommendationSource]int, ownerID string) error {
	activities := getRecommendationSources(recommendations, signals, types.RecommendationSignalSourceAsset)
	assets := []*generated.Asset{}

	for i := 0; ; i += perBatchScanSize {
		values, err := client.Asset.Query().
			Where(asset.OwnerIDEQ(ownerID)).
			Order(asset.ByID()).
			Limit(perBatchScanSize).
			Offset(i).
			All(ctx)
		if err != nil {
			return err
		}

		assets = append(assets, values...)

		if len(values) < perBatchScanSize {
			break
		}
	}

	for _, asset := range assets {
		text := text(asset.Name, asset.DisplayName)
		if text == "" {
			continue
		}

		for _, activity := range activities {
			if !strings.Contains(text, activity.value) {
				continue
			}

			trackWeight(activity.recommendation, weightTracker, types.RecommendationSignalSourceAsset, assetWeight, assetCap)
		}
	}

	return nil
}

func calculateWeightForVendors(ctx context.Context, client *generated.Client, recommendations map[string]*candidate, signals map[string][]types.RecommendationSignal, weightTracker map[string]map[types.RecommendationSource]int, ownerID string) error {
	activities := getRecommendationSources(recommendations, signals, types.RecommendationSignalSourceVendor)
	entities := []*generated.Entity{}

	for i := 0; ; i += perBatchScanSize {
		values, err := client.Entity.Query().
			Where(entity.OwnerIDEQ(ownerID)).
			Order(entity.ByID()).
			Limit(perBatchScanSize).
			Offset(i).
			All(ctx)
		if err != nil {
			return err
		}

		entities = append(entities, values...)

		if len(values) < perBatchScanSize {
			break
		}
	}

	for _, entity := range entities {
		text := text(entity.Name, entity.DisplayName)
		if text == "" {
			continue
		}

		for _, activity := range activities {
			if !strings.Contains(text, activity.value) {
				continue
			}

			trackWeight(activity.recommendation, weightTracker, types.RecommendationSignalSourceVendor, vendorWeight, vendorCap)
		}
	}

	return nil
}

func calculateWeightForAuthProviders(ctx context.Context, client *generated.Client, recommendations map[string]*candidate, recommendationSignals map[string][]types.RecommendationSignal, weightTracker map[string]map[types.RecommendationSource]int, ownerID string) error {
	signals := getRecommendationSources(recommendations, recommendationSignals, types.RecommendationSignalSourceSignInProvider)

	authProviders := lo.Uniq(lo.FilterMap(signals, func(signal source, _ int) (enums.AuthProvider, bool) {
		provider := enums.ToAuthProvider(signal.value)
		if provider == nil || *provider == enums.AuthProviderInvalid {
			return enums.AuthProviderInvalid, false
		}

		return *provider, true
	}))

	slices.SortFunc(authProviders, func(a, b enums.AuthProvider) int {
		return strings.Compare(string(a), string(b))
	})

	if len(authProviders) == 0 {
		return nil
	}

	memberships, err := client.OrgMembership.Query().
		Where(
			orgmembership.OrganizationIDEQ(ownerID),
			orgmembership.HasUserWith(user.Or(
				user.LastLoginProviderIn(authProviders...),
				user.AuthProviderIn(authProviders...),
			)),
		).
		WithUser().
		All(ctx)
	if err != nil {
		return err
	}

	providerCounts := map[enums.AuthProvider]int{}
	for _, membership := range memberships {
		user, err := membership.Edges.UserOrErr()
		if err != nil {
			continue
		}

		if slices.Contains(authProviders, user.LastLoginProvider) {
			providerCounts[user.LastLoginProvider]++
		}
		if user.AuthProvider != user.LastLoginProvider && slices.Contains(authProviders, user.AuthProvider) {
			providerCounts[user.AuthProvider]++
		}
	}

	for provider, count := range providerCounts {
		value := strings.ToLower(strings.TrimSpace(provider.String()))
		if value == "" {
			continue
		}

		for _, signal := range signals {
			if !strings.EqualFold(signal.value, value) {
				continue
			}

			trackWeight(signal.recommendation, weightTracker, types.RecommendationSignalSourceSignInProvider, authProviderWeight*count, authProviderCap)
		}
	}

	return nil
}

func getRecommendationSources(recommendations map[string]*candidate, signals map[string][]types.RecommendationSignal, src types.RecommendationSource) []source {
	return lo.FlatMap(lo.Values(recommendations), func(recommendation *candidate, _ int) []source {
		return lo.FlatMap(signals[recommendation.def.ID], func(signal types.RecommendationSignal, _ int) []source {
			if signal.Source != src {
				return nil
			}

			return lo.FilterMap(signal.Values, func(value string, _ int) (source, bool) {
				value = strings.ToLower(strings.TrimSpace(value))
				if value == "" {
					return source{}, false
				}

				return source{
					recommendation: recommendation,
					value:          value,
				}, true
			})
		})
	})
}

func trackWeight(rec *candidate, tracker map[string]map[types.RecommendationSource]int, src types.RecommendationSource, weight, srcCap int) {
	if tracker[rec.def.ID] == nil {
		tracker[rec.def.ID] = map[types.RecommendationSource]int{}
	}

	// cap weight scoring
	current := tracker[rec.def.ID][src]
	remaining := srcCap - current
	if remaining <= 0 {
		return
	}

	if weight > remaining {
		weight = remaining
	}

	tracker[rec.def.ID][src] = current + weight
	rec.weight += weight

	if rec.label != "" {
		return
	}

	switch src {
	case types.RecommendationSignalSourceAsset:
		rec.label = "Recommended based on matching assets"
	case types.RecommendationSignalSourceVendor:
		rec.label = "Recommended based on matching vendors"
	case types.RecommendationSignalSourceSSOProvider:
		rec.label = "Recommended based on your SSO provider"
	case types.RecommendationSignalSourceSignInProvider:
		rec.label = "Recommended based on sign-in activity"
	default:
		rec.label = "Recommended based on organization activity"
	}
}

func text(strs ...string) string {
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
