package hooks_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/taskrules"
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

func (suite *HookTestSuite) TestOnboardingProgramFrameworkSelections() {
	t := suite.T()

	admin := suite.seedSystemAdmin()
	sysCtx := generated.NewContext(auth.NewTestContextForSystemAdmin(admin.ID, admin.Edges.OrgMemberships[0].OrganizationID), suite.client)
	framework := gofakeit.UUID()

	std, err := suite.client.Standard.Create().SetName(framework).SetShortName(framework).
		SetFramework(framework).SetIsPublic(true).SetSystemOwned(true).
		SetStatus(enums.StandardActive).Save(sysCtx)
	require.NoError(t, err)

	_, err = suite.client.Control.Create().SetStandardID(std.ID).
		SetSystemOwned(true).SetRefCode("TEST-1").Save(sysCtx)
	require.NoError(t, err)

	for _, tc := range []struct {
		name         string
		compliance   map[string]interface{}
		programCount int
		controlCount int
		expectedErr  string
	}{
		{name: "no compliance"},
		{name: "no frameworks", compliance: map[string]interface{}{"existing_controls": true}},
		{name: "empty frameworks", compliance: map[string]interface{}{"frameworks": []string{}}},
		{
			name:         "single framework",
			compliance:   map[string]interface{}{"frameworks": []string{framework}},
			programCount: 1,
			controlCount: 1,
		},
		{
			name:         "other framework",
			compliance:   map[string]interface{}{"frameworks": []string{"other"}},
			programCount: 1,
		},
		{
			name:        "unknown framework",
			compliance:  map[string]interface{}{"frameworks": []string{"missing-framework"}},
			expectedErr: "resolve onboarding framework",
		},
		{
			name:        "invalid framework input",
			compliance:  map[string]interface{}{"frameworks": []interface{}{123}},
			expectedErr: "invalid onboarding frameworks",
		},
	} {

		t.Run(tc.name, func(t *testing.T) {
			user := suite.seedUser()
			ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID), suite.client)
			ctx = privacy.DecisionContext(ctx, privacy.Allow)

			onboarding, err := suite.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
				CompanyName: gofakeit.Company(),
				Compliance:  tc.compliance,
			}).Save(ctx)

			if tc.expectedErr != "" {
				require.ErrorContains(t, err, tc.expectedErr)
				return
			}
			require.NoError(t, err)

			suite.waitForEvents()

			programs, err := suite.client.Program.Query().Where(program.OwnerIDEQ(onboarding.OrganizationID)).WithControls().All(ctx)
			require.NoError(t, err)
			require.Len(t, programs, tc.programCount)

			if tc.programCount > 0 {
				assert.Len(t, programs[0].Edges.Controls, tc.controlCount)
			}

			taskCount, err := suite.client.Task.Query().Where(
				task.OwnerIDEQ(onboarding.OrganizationID),
				task.SourceKeyEQ("onboarding-"+taskrules.RuleFrameworkGeneric),
			).Count(ctx)
			require.NoError(t, err)

			if tc.programCount == 0 {
				assert.Equal(t, 1, taskCount)
			} else {
				assert.Zero(t, taskCount)
			}
		})
	}
}
