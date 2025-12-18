package graphapi_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/testclient"
)

func TestMutationCreateOnboarding(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	onboardingUser := suite.userBuilder(context.Background(), t)
	onboardingUser2 := suite.userBuilder(context.Background(), t)

	companyName := "Test Acme Corp, Inc."

	testCases := []struct {
		name        string
		request     testclient.CreateOnboardingInput
		client      *testclient.TestClient
		ctx         context.Context
		expectedErr string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateOnboardingInput{
				CompanyName: companyName,
			},
			client: suite.client.api,
			ctx:    onboardingUser.UserCtx,
		},
		{
			name: "happy path, all input, same name should not error due to retries",
			request: testclient.CreateOnboardingInput{
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
			ctx:    onboardingUser2.UserCtx,
		},
		{
			name:        "missing required field",
			request:     testclient.CreateOnboardingInput{},
			client:      suite.client.api,
			ctx:         onboardingUser.UserCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "not allowed with PAT",
			request: testclient.CreateOnboardingInput{
				CompanyName: companyName,
			},
			client:      suite.client.apiWithPAT,
			ctx:         context.Background(),
			expectedErr: common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name: "not allowed with token",
			request: testclient.CreateOnboardingInput{
				CompanyName: companyName,
			},
			client:      suite.client.apiWithToken,
			ctx:         context.Background(),
			expectedErr: common.ErrResourceNotAccessibleWithToken.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run("Create "+tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateOnboarding(tc.ctx, tc.request)
			if tc.expectedErr != "" {
				assert.Assert(t, is.ErrorContains(err, tc.expectedErr))
				assert.ErrorContains(t, err, tc.expectedErr)

				return
			}

			assert.NilError(t, err)
			assert.Assert(t, resp != nil)

			// check required fields
			assert.Check(t, resp.CreateOnboarding.Onboarding.ID != "")
			assert.Check(t, resp.CreateOnboarding.Onboarding.OrganizationID != nil)
			assert.Check(t, is.Equal(tc.request.CompanyName, resp.CreateOnboarding.Onboarding.CompanyName))

			// Cleanup onboarding data
			(&Cleanup[*generated.OnboardingDeleteOne]{client: suite.client.db.Onboarding, IDs: []string{resp.CreateOnboarding.Onboarding.ID}}).MustDelete(tc.ctx, t)
		})
	}
}
