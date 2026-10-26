package rule

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/utils"
)

// DenyIfMissingCustomDomainFeature checks to validate that the org has the right modules based off the mutation.
// the "purpose" field is used to determine what the user should have access to.
func DenyIfMissingCustomDomainFeature() privacy.CustomDomainMutationRuleFunc {
	return privacy.CustomDomainMutationRuleFunc(func(ctx context.Context, m *generated.CustomDomainMutation) error {

		if !m.Op().Is(ent.OpCreate) || !utils.ModulesEnabled(m.Client()) {
			return privacy.Skip
		}

		purpose, _ := m.Purpose()

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

		return privacy.Skip
	})
}
