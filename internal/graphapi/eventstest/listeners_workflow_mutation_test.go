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
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
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

	ctx := th.SetContext(initiator.UserCtx, suite.Client.DB)

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

func TestWorkflowMutationListenerCreateCallers(t *testing.T) {
	owner := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	policyManager := suite.OrgMemberWithFunctionalRoles(t, owner, "policy_manager")
	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)
	tokenClient := suite.SetupAPITokenClient(owner.UserCtx, t)

	_, workflowRuntime := acquireWorkflowRuntime(t)

	workflowDef := createWorkflowDefinition(ctx, t, owner.OrganizationID, "InternalPolicy", enums.WorkflowKindNotification, models.WorkflowDefinitionDocument{
		Triggers:   []models.WorkflowTrigger{{Operation: "CREATE"}},
		Conditions: []models.WorkflowCondition{{Expression: "true"}},
		Actions: []models.WorkflowAction{{
			Type: enums.WorkflowActionTypeNotification.String(),
			Key:  "policy_created_notification",
		}},
	})

	testCases := []struct {
		name   string
		client *testclient.TestClient
		ctx    context.Context
	}{
		{
			name:   "policy manager creates a policy",
			client: suite.Client.API,
			ctx:    policyManager.UserCtx,
		},
		{
			name:   "api token creates a policy",
			client: tokenClient,
			ctx:    context.Background(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := tc.client.CreateInternalPolicy(tc.ctx, testclient.CreateInternalPolicyInput{
				Name: "workflow trigger policy " + ulids.New().String(),
			})
			assert.NilError(t, err)

			waitForGala(t, workflowRuntime)

			waitForWorkflowInstance(ctx, t, workflowDef.ID, workflowinstance.InternalPolicyIDEQ(resp.CreateInternalPolicy.InternalPolicy.ID), "policy create should start the workflow")
		})
	}
}

func TestWorkflowMutationListenerMemberUpdateStartsReview(t *testing.T) {
	owner := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	policyManager := suite.OrgMemberWithFunctionalRoles(t, owner, "policy_manager")
	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)

	_, workflowRuntime := acquireWorkflowRuntime(t)

	params, err := json.Marshal(workflows.ReviewActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: owner.ID}},
		},
		Label: "Policy Name Review",
	})
	assert.NilError(t, err)

	workflowDef := createWorkflowDefinition(ctx, t, owner.OrganizationID, "InternalPolicy", enums.WorkflowKindApproval, models.WorkflowDefinitionDocument{
		Triggers:   []models.WorkflowTrigger{{Operation: "UPDATE", Fields: []string{"name"}}},
		Conditions: []models.WorkflowCondition{{Expression: "true"}},
		Actions: []models.WorkflowAction{{
			Type:   enums.WorkflowActionTypeReview.String(),
			Key:    "policy_name_review",
			Params: params,
		}},
	})

	created, err := suite.Client.API.CreateInternalPolicy(policyManager.UserCtx, testclient.CreateInternalPolicyInput{
		Name: "workflow review policy " + ulids.New().String(),
	})
	assert.NilError(t, err)

	policyID := created.CreateInternalPolicy.InternalPolicy.ID

	_, err = suite.Client.API.UpdateInternalPolicy(policyManager.UserCtx, policyID, testclient.UpdateInternalPolicyInput{
		Name: lo.ToPtr("workflow review policy renamed " + ulids.New().String()),
	})
	assert.NilError(t, err)

	waitForGala(t, workflowRuntime)

	instance := waitForWorkflowInstance(ctx, t, workflowDef.ID, workflowinstance.InternalPolicyIDEQ(policyID), "policy updated by a non-admin should start the review workflow")

	assignments, err := graphapi.WaitForAssignments(ctx, suite.Client.DB, instance.ID, 1)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.WorkflowAssignmentStatusPending, assignments[0].Status))
}

func TestWorkflowMutationListenerGroupSelectorMatchesPrivateGroup(t *testing.T) {
	owner := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	policyManager := suite.OrgMemberWithFunctionalRoles(t, owner, "policy_manager")
	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)

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

	workflowDef := createWorkflowDefinition(ctx, t, owner.OrganizationID, "InternalPolicy", enums.WorkflowKindNotification, models.WorkflowDefinitionDocument{
		Triggers: []models.WorkflowTrigger{{
			Operation: "UPDATE",
			Fields:    []string{"details"},
			Selector:  models.WorkflowSelector{GroupIDs: []string{privateGroup.ID}},
		}},
		Conditions: []models.WorkflowCondition{{Expression: "true"}},
		Actions: []models.WorkflowAction{{
			Type: enums.WorkflowActionTypeNotification.String(),
			Key:  "private_group_policy_notification",
		}},
	})

	created, err := suite.Client.API.CreateInternalPolicy(owner.UserCtx, testclient.CreateInternalPolicyInput{
		Name:      "workflow selector policy " + ulids.New().String(),
		EditorIDs: []string{privateGroup.ID},
	})
	assert.NilError(t, err)

	policyID := created.CreateInternalPolicy.InternalPolicy.ID

	_, err = suite.Client.API.UpdateInternalPolicy(policyManager.UserCtx, policyID, testclient.UpdateInternalPolicyInput{
		Details: lo.ToPtr("updated by the policy manager"),
	})
	assert.NilError(t, err)

	waitForGala(t, workflowRuntime)

	waitForWorkflowInstance(ctx, t, workflowDef.ID, workflowinstance.InternalPolicyIDEQ(policyID), "policy in a private group updated by a non-admin should start the workflow")
}

func TestWorkflowMutationListenerTriggeredByAnonymousRespondent(t *testing.T) {
	owner := suite.UserBuilder(context.Background(), t)
	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)

	_, workflowRuntime := acquireWorkflowRuntime(t)

	template := (&th.TemplateBuilder{Client: suite.Client}).MustNew(owner.UserCtx, t)
	assessment := (&th.AssessmentBuilder{Client: suite.Client, TemplateID: template.ID}).MustNew(owner.UserCtx, t)
	response := (&th.AssessmentResponseBuilder{Client: suite.Client, AssessmentID: assessment.ID, OwnerID: owner.OrganizationID}).MustNew(owner.UserCtx, t)

	workflowDef := createWorkflowDefinition(ctx, t, owner.OrganizationID, "AssessmentResponse", enums.WorkflowKindNotification, models.WorkflowDefinitionDocument{
		Triggers:   []models.WorkflowTrigger{{Operation: "UPDATE", Fields: []string{"status"}}},
		Conditions: []models.WorkflowCondition{{Expression: "true"}},
		Actions: []models.WorkflowAction{{
			Type: enums.WorkflowActionTypeNotification.String(),
			Key:  "response_completed_notification",
		}},
	})

	respondent := auth.NewQuestionnaireCaller(owner.OrganizationID, ulids.New().String(), "Anonymous Respondent", "")
	respondentCtx := privacy.DecisionContext(auth.WithCaller(context.Background(), respondent), privacy.Allow)

	assert.NilError(t, suite.Client.DB.AssessmentResponse.UpdateOneID(response.ID).
		SetStatus(enums.AssessmentResponseStatusCompleted).
		Exec(respondentCtx))

	waitForGala(t, workflowRuntime)

	waitForWorkflowInstance(ctx, t, workflowDef.ID, workflowinstance.AssessmentResponseIDEQ(response.ID), "response completed by an anonymous respondent should start the workflow")
}
