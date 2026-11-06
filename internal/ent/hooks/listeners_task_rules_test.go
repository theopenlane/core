//go:build test

package hooks_test

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	"github.com/theopenlane/core/v2/internal/ent/generated/program"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/hooks"
	"github.com/theopenlane/core/v2/internal/ent/taskrules"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestTaskRuleListenersCreateSuggestedTasks(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedUser()

	userCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)
	userCtx = generated.NewContext(userCtx, s.client)

	ctx := auth.WithInternalOperationContext(userCtx)
	ctx = generated.NewContext(ctx, s.client)

	onboarding, err := s.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Task Rule Co",
	}).Save(ctx)
	assert.NilError(t, err)

	s.waitForEvents()

	tasks, err := s.client.Task.Query().Where(task.OwnerIDEQ(onboarding.OrganizationID)).All(ctx)
	assert.NilError(t, err)

	sourceKeys := make([]string, 0, len(tasks))
	for _, tk := range tasks {
		sourceKeys = append(sourceKeys, tk.SourceKey)
	}

	// organization schema-level rules fire because the onboarding-created org is not personal
	assert.Check(t, is.Contains(sourceKeys, "organization-"+taskrules.RuleSecureOrganization))
	assert.Check(t, is.Contains(sourceKeys, "organization-"+taskrules.RuleInviteTeam))

	// onboarding compliance answers were left blank, so the unanswered-fallback rule fires
	assert.Check(t, is.Contains(sourceKeys, "onboarding-"+taskrules.RuleImportTemplateControls))
	assert.Check(t, is.Contains(sourceKeys, "onboarding-"+taskrules.RuleFrameworkGeneric))
}

// TestTaskRuleListenersNotificationTaskOwnerAttribution guards against a suggested task
// created off a system-context (CapInternalOperation) mutation ending up
// ownerless: ObjectOwnedMixin's create hook assumes such callers set owner_id themselves,
// so createSuggestedTask must set it explicitly rather than relying on auto-derivation
func TestTaskRuleListenersNotificationTaskOwnerAttribution(t *testing.T) {
	s := suite.setupTest(t)

	user := s.seedUser()
	orgA := user.Edges.OrgMemberships[0].OrganizationID

	userCtx := auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].ID)
	userCtx = generated.NewContext(userCtx, s.client)
	userCtx = auth.WithInternalOperationContext(userCtx)

	orgBEntity, err := s.client.Organization.Create().SetInput(generated.CreateOrganizationInput{
		Name: "Org B " + gofakeit.LetterN(8),
	}).Save(userCtx)
	assert.NilError(t, err)

	orgB := orgBEntity.ID

	// mimic domainScanSystemContext exactly: explicit caller.OrganizationID = orgA, bypassing FGA
	scanSystemCtx := auth.WithCaller(context.Background(), &auth.Caller{
		OrganizationID: orgA,
		Capabilities:   auth.CapInternalOperation,
	})

	_, err = s.client.Notification.Create().
		SetOwnerID(orgA).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType("scan.created").
		SetTitle("Domain scan completed").
		SetBody("test").
		SetData(map[string]any{}).
		SetTopic(enums.NotificationTopicDomainScan).
		Save(scanSystemCtx)
	assert.NilError(t, err)

	s.waitForEvents()

	bypassCtx := generated.NewContext(scanSystemCtx, s.client)

	tasksA, err := s.client.Task.Query().Where(task.OwnerIDEQ(orgA)).All(bypassCtx)
	assert.NilError(t, err)

	tasksB, err := s.client.Task.Query().Where(task.OwnerIDEQ(orgB)).All(bypassCtx)
	assert.NilError(t, err)

	assert.Check(t, len(tasksA) > 0)
	assert.Check(t, is.Len(tasksB, 0))
}

func TestOnboardingAddDemoDataLaterTask(t *testing.T) {
	testCases := []struct {
		name      string
		seedDemo  *bool
		wantCount int
	}{
		{name: "declined demo data", seedDemo: new(false), wantCount: 1},
		{name: "chose demo data", seedDemo: new(true), wantCount: 0},
		{name: "unanswered", wantCount: 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := suite.setupTest(t)
			user := s.seedUser()
			ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID), s.client)
			ctx = auth.WithInternalOperationContext(ctx)

			onboarding, err := s.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
				CompanyName:  "Demo Task Co " + gofakeit.LetterN(8),
				SeedDemoData: tc.seedDemo,
			}).Save(ctx)
			assert.NilError(t, err)

			s.waitForEvents()

			count, err := s.client.Task.Query().Where(
				task.OwnerIDEQ(onboarding.OrganizationID),
				task.SourceKeyEQ("onboarding-"+taskrules.RuleAddDemoDataLater),
			).Count(ctx)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tc.wantCount, count))
		})
	}
}

func TestOnboardingCreatesProgramWithSelectedFrameworks(t *testing.T) {
	s := suite.setupTest(t)
	user := s.seedUser()
	ctx := generated.NewContext(auth.NewTestContextWithOrgID(user.ID, user.Edges.OrgMemberships[0].OrganizationID), s.client)
	ctx = auth.WithInternalOperationContext(ctx)

	admin := s.seedSystemAdmin()
	sysCtx := generated.NewContext(auth.NewTestContextForSystemAdmin(admin.ID, admin.Edges.OrgMemberships[0].OrganizationID), s.client)

	for _, framework := range []struct {
		code  string
		label string
	}{
		{code: "soc2", label: "SOC 2"},
		{code: "iso27001", label: "ISO 27001"},
	} {
		std, err := s.client.Standard.Create().
			SetSystemOwned(true).
			SetIsPublic(true).
			SetFramework(framework.code).
			SetShortName(framework.label).
			SetName(framework.label).
			SetStatus(enums.StandardActive).
			Save(sysCtx)
		assert.NilError(t, err)

		for _, category := range []string{"Security", "Availability", "Confidentiality", "Processing Integrity", "Privacy"} {
			control, err := s.client.Control.Create().
				SetSystemOwned(true).
				SetStandardID(std.ID).
				SetRefCode(framework.code + "-" + category).
				SetCategory(category).
				Save(sysCtx)
			assert.NilError(t, err)

			_, err = s.client.Subcontrol.Create().
				SetSystemOwned(true).
				SetControlID(control.ID).
				SetRefCode(control.RefCode + "-1").
				SetCategory(category).
				Save(sysCtx)
			assert.NilError(t, err)
		}
	}

	onboarding, err := s.client.Onboarding.Create().SetInput(generated.CreateOnboardingInput{
		CompanyName: "Framework Program Co",
		Compliance: map[string]any{
			"frameworks":    []any{"soc2", "iso27001", "soc2"},
			"auditor_name":  "Jane Doe",
			"auditor_email": "jane@example.com",
		},
	}).Save(ctx)
	assert.NilError(t, err)

	s.waitForEvents()

	created, err := s.client.Program.Query().Where(program.OwnerIDEQ(onboarding.OrganizationID)).
		WithControls(func(q *generated.ControlQuery) { q.WithSubcontrols() }).
		WithMembers().Only(ctx)

	assert.NilError(t, err)
	assert.Check(t, is.Equal("SOC 2, ISO 27001", created.FrameworkName))
	assert.Check(t, is.Equal(fmt.Sprintf("Compliance Program %d", time.Now().Year()), created.Name))
	assert.Check(t, is.Equal("Jane Doe", created.Auditor))
	assert.Check(t, is.Equal("jane@example.com", created.AuditorEmail))
	assert.Assert(t, is.Len(created.Edges.Members, 1))

	refs := make([]string, 0, len(created.Edges.Controls))

	for _, control := range created.Edges.Controls {
		refs = append(refs, control.RefCode)
		assert.Check(t, is.Equal(onboarding.OrganizationID, control.OwnerID))
		assert.Check(t, !control.SystemOwned)
		assert.Assert(t, is.Len(control.Edges.Subcontrols, 1))
		assert.Check(t, is.Equal(control.RefCode+"-1", control.Edges.Subcontrols[0].RefCode))
		assert.Check(t, is.Equal(onboarding.OrganizationID, control.Edges.Subcontrols[0].OwnerID))
		assert.Check(t, is.Equal(control.ID, control.Edges.Subcontrols[0].ControlID))
	}
	assert.Check(t, is.DeepEqual(slices.Sorted(slices.Values([]string{
		"soc2-Security", "iso27001-Security", "iso27001-Availability",
		"iso27001-Confidentiality", "iso27001-Processing Integrity", "iso27001-Privacy",
	})), slices.Sorted(slices.Values(refs))))

	tasks, err := s.client.Task.Query().Where(task.OwnerIDEQ(onboarding.OrganizationID)).All(ctx)
	assert.NilError(t, err)
	assert.Assert(t, len(tasks) > 0)

	for _, tk := range tasks {
		assert.Check(t, "onboarding-framework-soc2" != tk.SourceKey)
		assert.Check(t, "onboarding-framework-iso27001" != tk.SourceKey)
		assert.Check(t, "onboarding-"+taskrules.RuleFrameworkGeneric != tk.SourceKey)
	}
}

func TestTaskRuleNotificationTopicGate(t *testing.T) {
	t.Parallel()

	listener, ok := taskRuleListenerFor(generated.TypeNotification)
	assert.Assert(t, ok)

	gate := listener.Definition().Gate

	assert.Check(t, gate(context.Background(), notificationCreatePayload(enums.NotificationTopicDomainScan)))
	assert.Check(t, !gate(context.Background(), notificationCreatePayload(enums.NotificationTopicApproval)))
	assert.Check(t, !gate(context.Background(), entityops.MutationPayload{MutationType: generated.TypeNotification, Operation: entityops.OpCreate, EntityID: "notification-1"}))
}

func taskRuleListenerFor(schemaName string) (entityops.MutationListener, bool) {
	for _, registration := range hooks.TaskRuleListeners() {
		listener, ok := registration.(entityops.MutationListener)
		if ok && listener.Schema.Name == schemaName {
			return listener, true
		}
	}

	return entityops.MutationListener{}, false
}

func notificationCreatePayload(topic enums.NotificationTopic) entityops.MutationPayload {
	return entityops.MutationPayload{
		MutationType:    generated.TypeNotification,
		Operation:       entityops.OpCreate,
		EntityID:        "notification-1",
		ProposedChanges: map[string]any{notification.FieldTopic: topic},
	}
}
