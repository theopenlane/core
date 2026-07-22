//go:build test

package hooks_test

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
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

// TestTaskRuleListenersNotificationTaskOwnerAttribution guards against a suggested task
// created off a system-context (CapInternalOperation|CapBypassFGA) mutation ending up
// ownerless: ObjectOwnedMixin's create hook assumes such callers set owner_id themselves,
// so createSuggestedTask must set it explicitly rather than relying on auto-derivation
func (suite *HookTestSuite) TestTaskRuleListenersNotificationTaskOwnerAttribution() {
	t := suite.T()

	user := suite.seedUser()
	orgA := user.Edges.OrgMemberships[0].OrganizationID

	userCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)
	userCtx = generated.NewContext(userCtx, suite.client)
	userCtx = privacy.DecisionContext(userCtx, privacy.Allow)

	orgBEntity, err := suite.client.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name: "Org B " + gofakeit.LetterN(8),
	}).Save(userCtx)
	require.NoError(t, err)

	orgB := orgBEntity.ID

	// mimic domainScanSystemContext exactly: explicit caller.OrganizationID = orgA, bypassing FGA
	scanSystemCtx := auth.WithCaller(privacy.DecisionContext(context.Background(), privacy.Allow), &auth.Caller{
		OrganizationID: orgA,
		Capabilities:   auth.CapBypassFGA | auth.CapInternalOperation,
	})

	_, err = suite.client.Notification.Create().
		SetOwnerID(orgA).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType("scan.created").
		SetTitle("Domain scan completed").
		SetBody("test").
		SetData(map[string]interface{}{}).
		SetTopic(enums.NotificationTopicDomainScan).
		Save(scanSystemCtx)
	require.NoError(t, err)

	suite.galaRuntime.WaitIdle()

	bypassCtx := generated.NewContext(scanSystemCtx, suite.client)

	tasksA, err := suite.client.Task.Query().Where(task.OwnerIDEQ(orgA)).All(bypassCtx)
	require.NoError(t, err)

	tasksB, err := suite.client.Task.Query().Where(task.OwnerIDEQ(orgB)).All(bypassCtx)
	require.NoError(t, err)

	assert.NotEmpty(t, tasksA)
	assert.Empty(t, tasksB)
}

func (suite *HookTestSuite) TestTaskRuleListenersFrameworkLinkIncludesAuditorParams() {
	t := suite.T()

	user := suite.seedUser()

	userCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)
	userCtx = generated.NewContext(userCtx, suite.client)
	ctx := privacy.DecisionContext(userCtx, privacy.Allow)
	ctx = generated.NewContext(ctx, suite.client)

	onboarding, err := suite.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Framework Link Co",
		Compliance: map[string]interface{}{
			"frameworks":    []interface{}{"soc2", "iso27001"},
			"auditor_name":  "Jane Doe",
			"auditor_email": "jane@example.com",
		},
	}).Save(ctx)
	require.NoError(t, err)

	_, err = suite.client.Standard.Create().
		SetOwnerID(onboarding.OrganizationID).
		SetFramework("iso27001").
		SetShortName("ISO 27001").
		SetName("ISO/IEC 27001").
		SetStatus(enums.StandardActive).
		Save(ctx)
	require.NoError(t, err)

	suite.galaRuntime.WaitIdle()

	tasks, err := suite.client.Task.Query().Where(task.OwnerIDEQ(onboarding.OrganizationID)).All(ctx)
	require.NoError(t, err)

	links := make(map[string]string, len(tasks))
	for _, tk := range tasks {
		if link, ok := tk.Metadata["link"].(string); ok {
			links[tk.SourceKey] = link
		}
	}

	assert.Equal(t, "/programs/create/soc2?onboarding=true&auditorName=Jane Doe&auditorEmail=jane@example.com", links["onboarding-framework-soc2"])
	assert.Equal(t, "/programs/create/framework-based?onboarding=true&framework=ISO 27001&auditorName=Jane Doe&auditorEmail=jane@example.com", links["onboarding-framework-iso27001"])
}
