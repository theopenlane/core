package rule_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/v2/internal/ent/entconfig"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
)

func TestCustomDomainRule(t *testing.T) {

	tt := []struct {
		name      string
		purpose   enums.DomainPurpose
		feature   models.OrgModule
		isAllowed bool
	}{
		{
			name:      "can create assessment custom domain with assessment domain addon",
			purpose:   enums.DomainPurposeAssessment,
			feature:   models.CatalogAssessmentCustomDomainAddon,
			isAllowed: true,
		},
		{
			name:      "cannot create assessment custom domain with trust center module",
			purpose:   enums.DomainPurposeAssessment,
			feature:   models.CatalogTrustCenterModule,
			isAllowed: false,
		},
		{
			name:      "can create trust center custom domain with trust center module",
			purpose:   enums.DomainPurposeTrustcenter,
			feature:   models.CatalogTrustCenterModule,
			isAllowed: true,
		},
		{
			name:      "cannot create trust center custom domain with only assessment domain addon",
			purpose:   enums.DomainPurposeTrustcenter,
			feature:   models.CatalogAssessmentCustomDomainAddon,
			isAllowed: false,
		},
		{
			name:      "cannot create assessment domain with only base module",
			purpose:   enums.DomainPurposeAssessment,
			feature:   models.CatalogBaseModule,
			isAllowed: false,
		},
		{
			name:    "cannot create trust center custom domain with base module",
			purpose: enums.DomainPurposeTrustcenter,
			feature: models.CatalogBaseModule,
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {

			ctx := setupContext(t, []models.OrgModule{tt.feature})

			client := generated.NewClient(generated.EntConfig(&entconfig.Config{
				Modules: entconfig.Modules{
					Enabled: true,
				},
			}))

			m := client.CustomDomain.Create().
				SetPurpose(tt.purpose).
				Mutation()

			err := rule.DenyIfMissingCustomDomainFeature().EvalMutation(ctx, m)

			if tt.isAllowed {
				require.ErrorIs(t, err, privacy.Skip)
				return
			}

			require.ErrorIs(t, err, privacy.Deny)
			feature := models.CatalogTrustCenterModule
			if tt.purpose == enums.DomainPurposeAssessment {
				feature = models.CatalogAssessmentCustomDomainAddon
			}

			require.ErrorContains(t, err, feature.String())
		})
	}
}
