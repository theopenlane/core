//go:build test

package hooks_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/internal/ent/taskrules"
	"github.com/theopenlane/iam/auth"
)

func (suite *HookTestSuite) TestTaskRuleListenersCreateSuggestedTasks() {
	t := suite.T()

	user := suite.seedUser()

	userCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)
	userCtx = generated.NewContext(userCtx, suite.client)

	ctx := privacy.DecisionContext(userCtx, privacy.Allow)
	ctx = generated.NewContext(ctx, suite.client)

	onboarding, err := suite.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Task Rule Co",
	}).Save(ctx)
	require.NoError(t, err)

	suite.galaRuntime.WaitIdle()

	tasks, err := suite.client.Task.Query().Where(task.OwnerIDEQ(onboarding.OrganizationID)).All(ctx)
	require.NoError(t, err)

	sourceKeys := make([]string, 0, len(tasks))
	for _, tk := range tasks {
		sourceKeys = append(sourceKeys, tk.SourceKey)
	}

	// organization schema-level rules fire because the onboarding-created org is not personal
	assert.Contains(t, sourceKeys, "organization-"+taskrules.RuleSecureOrganization)
	assert.Contains(t, sourceKeys, "organization-"+taskrules.RuleInviteTeam)

	// onboarding compliance answers were left blank, so the unanswered-fallback rule fires
	assert.Contains(t, sourceKeys, "onboarding-"+taskrules.RuleImportTemplateControls)
}
