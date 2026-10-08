//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
)

func TestWorkflowAssignmentMutationListener(t *testing.T) {
	initiator := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	approver := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	suite.AddUserToOrganization(initiator.UserCtx, t, &approver, enums.RoleAdmin, initiator.OrganizationID)

	ctx := th.SetInternalContext(initiator.UserCtx, suite.Client.DB)

	workflowEngine, workflowRuntime := acquireWorkflowRuntime(t)

	params, err := json.Marshal(struct {
		Targets  []workflows.TargetConfig `json:"targets"`
		Required bool                     `json:"required"`
		Label    string                   `json:"label"`
	}{
		Targets:  []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: approver.ID}},
		Required: true,
		Label:    "Assignment Listener Approval",
	})
	assert.NilError(t, err)

	workflowDef, err := suite.Client.DB.WorkflowDefinition.Create().
		SetName("Assignment Listener Workflow").
		SetSchemaType("Control").
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetActive(true).
		SetOwnerID(initiator.OrganizationID).
		SetDefinitionJSON(models.WorkflowDefinitionDocument{
			Triggers:   []models.WorkflowTrigger{{Operation: "UPDATE", Fields: []string{"status"}}},
			Conditions: []models.WorkflowCondition{{Expression: "true"}},
			Actions: []models.WorkflowAction{{
				Type:   enums.WorkflowActionTypeApproval.String(),
				Key:    "assignment_listener_approval",
				Params: params,
			}},
		}).
		Save(ctx)
	assert.NilError(t, err)

	control, err := suite.Client.DB.Control.Create().
		SetRefCode("CTL-" + ulids.New().String()).
		SetTitle("Assignment Listener Control").
		SetStatus(enums.ControlStatusNotImplemented).
		SetOwnerID(initiator.OrganizationID).
		Save(ctx)
	assert.NilError(t, err)

	instance, err := workflowEngine.TriggerWorkflow(ctx, workflowDef, &workflows.Object{
		ID:   control.ID,
		Type: enums.WorkflowObjectTypeControl,
	}, engine.TriggerInput{EventType: "UPDATE", ChangedFields: []string{"status"}})
	assert.NilError(t, err)

	waitForGala(t, workflowRuntime)

	assignments, err := graphapi.WaitForAssignments(ctx, suite.Client.DB, instance.ID, 1)
	assert.NilError(t, err)
	assignment := assignments[0]
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, assignment.Status))

	_, err = graphapi.WaitForInstanceState(ctx, suite.Client.DB, instance.ID, enums.WorkflowInstanceStatePaused)
	assert.NilError(t, err)

	assertStillPending := func(t *testing.T) {
		t.Helper()

		waitForGala(t, workflowRuntime)

		current, err := suite.Client.DB.WorkflowInstance.Get(ctx, instance.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowInstanceStatePaused, current.State))

		reloaded, err := suite.Client.DB.WorkflowAssignment.Get(ctx, assignment.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, reloaded.Status))
	}

	t.Run("update without status change is skipped", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetNotes("still deciding").
			Exec(ctx))

		assertStillPending(t)
	})

	t.Run("status set to pending is skipped", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetStatus(enums.WorkflowAssignmentStatusPending).
			Exec(ctx))

		assertStillPending(t)
	})

	t.Run("non-pending status completes the assignment and the instance", func(t *testing.T) {
		assert.NilError(t, suite.Client.DB.WorkflowAssignment.UpdateOneID(assignment.ID).
			SetStatus(enums.WorkflowAssignmentStatusApproved).
			Exec(ctx))

		waitForGala(t, workflowRuntime)

		completed, err := graphapi.WaitForInstanceState(ctx, suite.Client.DB, instance.ID, enums.WorkflowInstanceStateCompleted)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowInstanceStateCompleted, completed.State))

		reloaded, err := suite.Client.DB.WorkflowAssignment.Get(ctx, assignment.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusApproved, reloaded.Status))
	})
}

func TestWorkflowMutationListenerInternalPolicy(t *testing.T) {
	org := suite.SeedOrgOwner(t)
	owner := *org.Owner
	policyManager := suite.OrgMemberWithFunctionalRoles(t, owner, "policy_manager")

	_, workflowRuntime := acquireWorkflowRuntime(t)

	privateGroup := (&th.GroupBuilder{Client: suite.Client}).MustNew(owner.UserCtx, t)

	groupResp, err := suite.Client.API.GetGroupByID(owner.UserCtx, privateGroup.ID)
	assert.NilError(t, err)

	_, err = suite.Client.API.UpdateGroupSetting(owner.UserCtx, groupResp.Group.Setting.ID, testclient.UpdateGroupSettingInput{
		Visibility: &enums.VisibilityPrivate,
	})
	assert.NilError(t, err)

	_, err = suite.Client.API.GetGroupByID(policyManager.UserCtx, privateGroup.ID)
	assert.Assert(t, err != nil)

	reviewParams, err := json.Marshal(workflows.ReviewActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: owner.ID}},
		},
		Label: "Policy Name Review",
	})
	assert.NilError(t, err)

	testCases := []struct {
		name                string
		client              *testclient.TestClient
		ctx                 context.Context
		kind                enums.WorkflowKind
		trigger             models.WorkflowTrigger
		action              models.WorkflowAction
		editorIDs           []string
		update              *testclient.UpdateInternalPolicyInput
		expectedAssignments int
	}{
		{
			name:    "policy manager create starts the workflow",
			client:  suite.Client.API,
			ctx:     policyManager.UserCtx,
			kind:    enums.WorkflowKindNotification,
			trigger: models.WorkflowTrigger{Operation: "CREATE"},
			action:  models.WorkflowAction{Type: enums.WorkflowActionTypeNotification.String(), Key: "policy_created_notification"},
		},
		{
			name:    "api token create starts the workflow",
			client:  org.APIClient,
			ctx:     context.Background(),
			kind:    enums.WorkflowKindNotification,
			trigger: models.WorkflowTrigger{Operation: "CREATE"},
			action:  models.WorkflowAction{Type: enums.WorkflowActionTypeNotification.String(), Key: "policy_created_notification"},
		},
		{
			name:                "policy manager update starts the review workflow",
			client:              suite.Client.API,
			ctx:                 policyManager.UserCtx,
			kind:                enums.WorkflowKindApproval,
			trigger:             models.WorkflowTrigger{Operation: "UPDATE", Fields: []string{"name"}},
			action:              models.WorkflowAction{Type: enums.WorkflowActionTypeReview.String(), Key: "policy_name_review", Params: reviewParams},
			update:              &testclient.UpdateInternalPolicyInput{Name: lo.ToPtr("workflow review policy renamed " + ulids.New().String())},
			expectedAssignments: 1,
		},
		{
			name:      "policy manager update matches a private group selector",
			client:    suite.Client.API,
			ctx:       owner.UserCtx,
			kind:      enums.WorkflowKindNotification,
			trigger:   models.WorkflowTrigger{Operation: "UPDATE", Fields: []string{"details"}, Selector: models.WorkflowSelector{GroupIDs: []string{privateGroup.ID}}},
			action:    models.WorkflowAction{Type: enums.WorkflowActionTypeNotification.String(), Key: "private_group_policy_notification"},
			editorIDs: []string{privateGroup.ID},
			update:    &testclient.UpdateInternalPolicyInput{Details: lo.ToPtr("updated by the policy manager")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workflowDef := createWorkflowDefinition(owner.UserCtx, t, owner.OrganizationID, "InternalPolicy", tc.kind, models.WorkflowDefinitionDocument{
				Triggers:   []models.WorkflowTrigger{tc.trigger},
				Conditions: []models.WorkflowCondition{{Expression: "true"}},
				Actions:    []models.WorkflowAction{tc.action},
			})

			created, err := tc.client.CreateInternalPolicy(tc.ctx, testclient.CreateInternalPolicyInput{
				Name:      "workflow trigger policy " + ulids.New().String(),
				EditorIDs: tc.editorIDs,
			})
			assert.NilError(t, err)

			policyID := created.CreateInternalPolicy.InternalPolicy.ID

			if tc.update != nil {
				_, err = suite.Client.API.UpdateInternalPolicy(policyManager.UserCtx, policyID, *tc.update)
				assert.NilError(t, err)
			}

			waitForGala(t, workflowRuntime)

			instance := waitForWorkflowInstance(owner.UserCtx, t, workflowDef.ID, workflowinstance.InternalPolicyIDEQ(policyID), tc.name)

			if tc.expectedAssignments == 0 {
				return
			}

			assignments, err := graphapi.WaitForAssignments(owner.UserCtx, suite.Client.DB, instance.ID, tc.expectedAssignments)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, assignments[0].Status))
		})
	}
}

func TestWorkflowMutationListenerTriggeredByAnonymousRespondent(t *testing.T) {
	owner := suite.UserBuilder(context.Background(), t)

	_, workflowRuntime := acquireWorkflowRuntime(t)

	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(owner.UserCtx, t)
	assessment := (&th.AssessmentBuilder{Client: suite.Client, TemplateID: template.ID}).MustNew(owner.UserCtx, t)
	response := (&th.AssessmentResponseBuilder{Client: suite.Client, AssessmentID: assessment.ID, OwnerID: owner.OrganizationID}).MustNew(owner.UserCtx, t)

	workflowDef := createWorkflowDefinition(owner.UserCtx, t, owner.OrganizationID, "AssessmentResponse", enums.WorkflowKindNotification, models.WorkflowDefinitionDocument{
		Triggers:   []models.WorkflowTrigger{{Operation: "UPDATE", Fields: []string{"status"}}},
		Conditions: []models.WorkflowCondition{{Expression: "true"}},
		Actions: []models.WorkflowAction{{
			Type: enums.WorkflowActionTypeNotification.String(),
			Key:  "response_completed_notification",
		}},
	})

	// covers the listener under the anonymous respondent caller which is why the internal operation is here, the handler path is covered by TestSubmitQuestionnaire
	respondent := auth.NewQuestionnaireCaller(owner.OrganizationID, ulids.New().String(), "Anonymous Respondent", "")
	respondentCtx := auth.WithInternalOperationContext(auth.WithCaller(context.Background(), respondent))

	assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		Exec(respondentCtx))

	waitForGala(t, workflowRuntime)

	waitForWorkflowInstance(owner.UserCtx, t, workflowDef.ID, workflowinstance.AssessmentResponseIDEQ(response.ID), "response completed by an anonymous respondent should start the workflow")
}
