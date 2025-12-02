package hooks_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/organization"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
)

func (suite *HookTestSuite) TestHookOnboarding() {
	t := suite.T()

	user := suite.seedUser()

	userCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)

	// add the client to the context for hooks
	userCtx = generated.NewContext(userCtx, suite.client)

	name := "MITB"

	testCases := []struct {
		name        string
		input       generated.CreateOnboardingInput
		expectedErr string
	}{
		{
			name: "valid onboarding, full data",
			input: generated.CreateOnboardingInput{
				CompanyName: name,
				Domains:     []string{gofakeit.DomainName(), gofakeit.DomainName()},
				CompanyDetails: map[string]interface{}{
					"sector":       "Technology",
					"company_size": "100-500",
				},
				UserDetails: map[string]interface{}{
					"name":       gofakeit.Name(),
					"job_title":  gofakeit.JobTitle(),
					"department": gofakeit.JobDescriptor(),
				},
				Compliance: map[string]interface{}{
					"existing_policies": true,
					"existing_controls": false,
					"risk_assessment":   true,
				},
			},
		},
		{
			name: "valid onboarding, same name, no details",
			input: generated.CreateOnboardingInput{
				CompanyName: name,
			},
		},
		{
			name: "valid onboarding, same name again, with domains",
			input: generated.CreateOnboardingInput{
				CompanyName: name,
				Domains:     []string{gofakeit.DomainName(), gofakeit.DomainName()},
			},
		},
		{
			name: "invalid onboarding, empty name",
			input: generated.CreateOnboardingInput{
				Domains: []string{gofakeit.DomainName(), gofakeit.DomainName()},
			},
			expectedErr: "company name is required",
		},
	}

	for i, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			// setup the allow context
			ctx := privacy.DecisionContext(userCtx, privacy.Allow)

			// add the client to the context
			ctx = generated.NewContext(ctx, suite.client)

			onboarding, err := suite.client.Onboarding.Create().SetInput(tc.input).Save(ctx)
			if tc.expectedErr != "" {

				require.Error(t, err)
				assert.Nil(t, onboarding)

				return
			}

			require.NoError(t, err)
			assert.NotNil(t, onboarding)

			assert.NotEmpty(t, onboarding.ID)
			assert.NotEmpty(t, onboarding.OrganizationID)
			assert.Equal(t, name, onboarding.CompanyName)

			org, err := suite.client.Organization.Query().
				Where(organization.IDEQ(onboarding.OrganizationID)).
				WithSetting().
				Only(ctx)
			require.NoError(t, err)

			assert.Equal(t, tc.input.CompanyName, org.DisplayName)

			if i == 0 {
				assert.Equal(t, tc.input.CompanyName, org.Name)
			} else {
				assert.NotEqual(t, tc.input.CompanyName, org.Name)
			}

			assert.ElementsMatch(t, tc.input.Domains, org.Edges.Setting.Domains)
		})
	}
}
