package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
)

// DenyIfMissingCustomDomainFeature checks to validate that the org has the right modules based off the mutation.
// the "purpose" field is used to determine what the user should have access to.
func DenyIfMissingCustomDomainFeature() privacy.CustomDomainMutationRuleFunc {
	return privacy.CustomDomainMutationRuleFunc(func(ctx context.Context, m *generated.CustomDomainMutation) error {

		// we want to skip internal/explicit ctx bypass so we can "turn off" the domain
		// from the webhook handlers
		if !utils.ModulesEnabled(m.Client()) || ShouldSkipFeatureCheck(ctx) {
			return privacy.Skip
		}

		purposes, err := getPurposesFromMutation(ctx, m)
		if err != nil {
			return err
		}

		for _, purpose := range purposes {

			// TODO(adelowo): if we ever add more purposes, we should auto-generate the mapping
			// for purposes <> module instead
			feature := models.CatalogTrustCenterModule

			if purpose == enums.DomainPurposeAssessment {
				feature = models.CatalogAssessmentCustomDomainAddon
			}

			ok, err := HasFeature(ctx, feature.String())
			if err != nil {
				return err
			}

			if !ok {
				return privacy.Denyf("feature %s not enabled", feature)
			}
		}

		return privacy.Skip
	})
}

func getPurposesFromMutation(ctx context.Context, m *generated.CustomDomainMutation) ([]enums.DomainPurpose, error) {

	if m.Op().Is(ent.OpCreate) {
		purpose, _ := m.Purpose()
		return []enums.DomainPurpose{purpose}, nil
	}

	ids, err := m.IDs(ctx)
	if err != nil {
		return nil, err
	}

	domains, err := m.Client().CustomDomain.Query().
		Where(customdomain.IDIn(ids...)).
		Select(customdomain.FieldPurpose).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return lo.Map(domains, func(domain *generated.CustomDomain, _ int) enums.DomainPurpose {
		return domain.Purpose
	}), nil
}
