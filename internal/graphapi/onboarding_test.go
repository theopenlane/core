package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/openlaneclient"
)

func (suite *GraphTestSuite) TestMutationCreateOnboarding() {
	t := suite.T()

	companyName := "Test Acme Corp, Inc."

	testCases := []struct {
		name        string
		request     openlaneclient.CreateOnboardingInput
		client      *openlaneclient.OpenlaneClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: openlaneclient.CreateOnboardingInput{
				CompanyName: companyName,
			},
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name: "happy path, all input",
			request: openlaneclient.CreateOnboardingInput{
				CompanyName: companyName,
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
			client: suite.client.api,
			ctx:    testUser1.UserCtx,
		},
		{
			name:        "missing required field",
			request:     openlaneclient.CreateOnboardingInput{},
			client:      suite.client.api,
			ctx:         testUser1.UserCtx,
			expectedErr: "value is less than the required length",
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateOnboarding(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr)
				assert.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			// check required fields
			require.NotNil(t, resp.CreateOnboarding)
			require.NotNil(t, resp.CreateOnboarding.Onboarding)
			assert.NotNil(t, resp.CreateOnboarding.Onboarding.ID)
			assert.NotNil(t, resp.CreateOnboarding.Onboarding.OrganizationID)
			assert.Equal(t, tc.request.CompanyName, resp.CreateOnboarding.Onboarding.CompanyName)
		})
	}
}
