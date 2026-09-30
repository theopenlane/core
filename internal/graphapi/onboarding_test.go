package graphapi_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/brianvoe/gofakeit/v7"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/sladefinition"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/common"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
)

func TestMutationCreateOnboarding(t *testing.T) {
	t.Parallel()

	// create another user for this test
	// so it doesn't interfere with the other tests
	onboardingUser := suite.UserBuilder(context.Background(), t)
	onboardingUser2 := suite.UserBuilder(context.Background(), t)

	personalOrgCtx := auth.NewTestContextWithOrgID(onboardingUser.ID, onboardingUser.PersonalOrgID)
	personalOrgCtx2 := auth.NewTestContextWithOrgID(onboardingUser2.ID, onboardingUser2.PersonalOrgID)

	// the shared harness does not wire the onboarding program listener
	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, hooks.OnboardingProgramListeners())
	assert.NilError(t, err)

	t.Cleanup(setup.Teardown)

	sysCtx := th.SharedSystemAdminUser.UserCtx

	soc2Standard := (&th.StandardBuilder{Client: suite.Client, Name: "soc2", Framework: "soc2", IsPublic: true}).MustNew(sysCtx, t)
	isoStandard := (&th.StandardBuilder{Client: suite.Client, Name: "iso27001:2022", Framework: "iso27001:2022", IsPublic: true}).MustNew(sysCtx, t)

	// only the Security category is cloned for soc2, so the availability control is never copied
	controlIDs := []string{}
	for _, builder := range []*th.ControlBuilder{
		{Client: suite.Client, StandardID: soc2Standard.ID, RefCode: "CC1.1", Category: "Security"},
		{Client: suite.Client, StandardID: soc2Standard.ID, RefCode: "CC1.2", Category: "Security"},
		{Client: suite.Client, StandardID: soc2Standard.ID, RefCode: "A1.1", Category: "Availability"},
		{Client: suite.Client, StandardID: isoStandard.ID, RefCode: "A.5.1", Category: "Organizational"},
		{Client: suite.Client, StandardID: isoStandard.ID, RefCode: "A.6.1", Category: "People"},
	} {
		controlIDs = append(controlIDs, builder.MustNew(sysCtx, t).ID)
	}

	t.Cleanup(func() {
		(&th.Cleanup[*generated.ControlDeleteOne]{Client: suite.Client.DB.Control, IDs: controlIDs}).MustDelete(sysCtx, t)
		(&th.Cleanup[*generated.StandardDeleteOne]{Client: suite.Client.DB.Standard, IDs: []string{soc2Standard.ID, isoStandard.ID}}).MustDelete(sysCtx, t)
	})

	companyName := "Test Acme Corp, Inc."

	testCases := []struct {
		name             string
		request          testclient.CreateOnboardingInput
		client           *testclient.TestClient
		ctx              context.Context
		expectedPrograms int
		expectedControls int
		expectedErr      string
	}{
		{
			name: "happy path, minimal input",
			request: testclient.CreateOnboardingInput{
				CompanyName: companyName,
			},
			client: suite.Client.API,
			ctx:    personalOrgCtx,
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
					"frameworks":                   []string{"soc2", "iso27001:2022"},
					"has_auditor":                  false,
					"existing_controls":            false,
					"recommend_auditors":           true,
					"recommend_vciso_partner":      true,
					"existing_policies_procedures": false},
				DemoRequested: lo.ToPtr(true),
			},
			client:           suite.Client.API,
			ctx:              personalOrgCtx2,
			expectedPrograms: 1,
			expectedControls: 4,
		},
		{
			name:        "missing required field",
			request:     testclient.CreateOnboardingInput{},
			client:      suite.Client.API,
			ctx:         personalOrgCtx,
			expectedErr: "value is less than the required length",
		},
		{
			name: "not allowed with PAT",
			request: testclient.CreateOnboardingInput{
				CompanyName: companyName,
			},
			client:      suite.Client.APIWithPAT,
			ctx:         context.Background(),
			expectedErr: common.ErrResourceNotAccessibleWithToken.Error(),
		},
		{
			name: "not allowed with token",
			request: testclient.CreateOnboardingInput{
				CompanyName: companyName,
			},
			client:      suite.Client.APIWithToken,
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

			slaCount, err := suite.Client.DB.SLADefinition.Query().
				Where(sladefinition.OwnerID(*resp.CreateOnboarding.Onboarding.OrganizationID)).
				Count(privacy.DecisionContext(tc.ctx, privacy.Allow))
			assert.NilError(t, err)
			assert.Check(t, is.Equal(4, slaCount))

			orgID := *resp.CreateOnboarding.Onboarding.OrganizationID

			assert.NilError(t, suite.GalaRuntime.WaitIdle(t.Context()))

			programs, err := suite.Client.DB.Program.Query().
				Where(program.OwnerID(orgID)).
				WithControls().
				All(th.SetContext(tc.ctx, suite.Client.DB))
			assert.NilError(t, err)
			assert.Assert(t, is.Len(programs, tc.expectedPrograms))

			if tc.expectedPrograms > 0 {
				assert.Check(t, is.Len(programs[0].Edges.Controls, tc.expectedControls))
			}

			// th.Cleanup onboarding data
			(&th.Cleanup[*generated.OnboardingDeleteOne]{Client: suite.Client.DB.Onboarding, IDs: []string{resp.CreateOnboarding.Onboarding.ID}}).MustDelete(tc.ctx, t)
		})
	}
}
