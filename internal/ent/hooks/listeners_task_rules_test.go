//go:build test

package hooks_test

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/taskrules"
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

	suite.waitForEvents()

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
	assert.Contains(t, sourceKeys, "onboarding-"+taskrules.RuleFrameworkGeneric)
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

	suite.waitForEvents()

	bypassCtx := generated.NewContext(scanSystemCtx, suite.client)

	tasksA, err := suite.client.Task.Query().Where(task.OwnerIDEQ(orgA)).All(bypassCtx)
	require.NoError(t, err)

	tasksB, err := suite.client.Task.Query().Where(task.OwnerIDEQ(orgB)).All(bypassCtx)
	require.NoError(t, err)

	assert.NotEmpty(t, tasksA)
	assert.Empty(t, tasksB)
}

func (suite *HookTestSuite) TestOnboardingCreatesProgramWithSelectedFrameworks() {
	t := suite.T()
	user := suite.seedUser()
	ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID), suite.client)
	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	admin := suite.seedSystemAdmin()
	sysCtx := generated.NewContext(auth.NewTestContextForSystemAdmin(admin.ID, admin.Edges.OrgMemberships[0].OrganizationID), suite.client)

	for _, framework := range []struct {
		code  string
		label string
	}{
		{code: "soc2", label: "SOC 2"},
		{code: "iso27001", label: "ISO 27001"},
	} {
		std, err := suite.client.Standard.Create().
			SetSystemOwned(true).
			SetIsPublic(true).
			SetFramework(framework.code).
			SetShortName(framework.label).
			SetName(framework.label).
			SetStatus(enums.StandardActive).
			Save(sysCtx)
		require.NoError(t, err)

		for _, category := range []string{"Security", "Availability", "Confidentiality", "Processing Integrity", "Privacy"} {
			control, err := suite.client.Control.Create().
				SetSystemOwned(true).
				SetStandardID(std.ID).
				SetRefCode(framework.code + "-" + category).
				SetCategory(category).
				Save(sysCtx)
			require.NoError(t, err)

			_, err = suite.client.Subcontrol.Create().
				SetSystemOwned(true).
				SetControlID(control.ID).
				SetRefCode(control.RefCode + "-1").
				SetCategory(category).
				Save(sysCtx)
			require.NoError(t, err)
		}
	}

	onboarding, err := suite.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Framework Program Co",
		Compliance: map[string]interface{}{
			"frameworks":    []interface{}{"soc2", "iso27001", "soc2"},
			"auditor_name":  "Jane Doe",
			"auditor_email": "jane@example.com",
		},
	}).Save(ctx)
	require.NoError(t, err)

	suite.waitForEvents()

	created, err := suite.client.Program.Query().Where(program.OwnerIDEQ(onboarding.OrganizationID)).
		WithControls(func(q *generated.ControlQuery) { q.WithSubcontrols() }).
		WithMembers().Only(ctx)

	require.NoError(t, err)
	assert.Equal(t, "SOC 2, ISO 27001", created.FrameworkName)
	assert.Equal(t, fmt.Sprintf("Compliance Program %d", time.Now().Year()), created.Name)
	assert.Equal(t, "Jane Doe", created.Auditor)
	assert.Equal(t, "jane@example.com", created.AuditorEmail)
	require.Len(t, created.Edges.Members, 1)

	refs := make([]string, 0, len(created.Edges.Controls))

	for _, control := range created.Edges.Controls {
		refs = append(refs, control.RefCode)
		assert.Equal(t, onboarding.OrganizationID, control.OwnerID)
		assert.False(t, control.SystemOwned)
		require.Len(t, control.Edges.Subcontrols, 1)
		assert.Equal(t, control.RefCode+"-1", control.Edges.Subcontrols[0].RefCode)
		assert.Equal(t, onboarding.OrganizationID, control.Edges.Subcontrols[0].OwnerID)
		assert.Equal(t, control.ID, control.Edges.Subcontrols[0].ControlID)
	}
	assert.ElementsMatch(t, []string{
		"soc2-Security", "iso27001-Security", "iso27001-Availability",
		"iso27001-Confidentiality", "iso27001-Processing Integrity", "iso27001-Privacy",
	}, refs)

	tasks, err := suite.client.Task.Query().Where(task.OwnerIDEQ(onboarding.OrganizationID)).All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)

	for _, tk := range tasks {
		assert.NotEqual(t, "onboarding-framework-soc2", tk.SourceKey)
		assert.NotEqual(t, "onboarding-framework-iso27001", tk.SourceKey)
		assert.NotEqual(t, "onboarding-"+taskrules.RuleFrameworkGeneric, tk.SourceKey)
	}
}
