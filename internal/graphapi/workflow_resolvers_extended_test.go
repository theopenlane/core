package graphapi_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/graphapi/testclient"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/core/v2/internal/workflows/engine"
	"github.com/theopenlane/utils/rout"
	"github.com/theopenlane/utils/ulids"
)

var (
	workflowEngineOnce   sync.Once
	workflowEngineErr    error
	sharedWorkflowEngine *engine.WorkflowEngine
)

// ensureWorkflowEngine registers one shared process-wide workflow engine; it is never
// unset so parallel tests always see a stable default
func ensureWorkflowEngine(t *testing.T) *engine.WorkflowEngine {
	t.Helper()

	workflowEngineOnce.Do(func() {
		sharedWorkflowEngine, workflowEngineErr = engine.NewWorkflowEngine(suite.Client.DB, nil)
		if workflowEngineErr == nil {
			engine.SetDefault(sharedWorkflowEngine)
		}
	})
	th.RequireNoError(t, workflowEngineErr)

	return sharedWorkflowEngine
}

func createWorkflowDefinition(t *testing.T, ctx context.Context, ownerID string) *ent.WorkflowDefinition {
	t.Helper()

	definition, err := suite.Client.DB.WorkflowDefinition.Create().
		SetName("Test Workflow " + ulids.New().String()).
		SetSchemaType("Control").
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetOwnerID(ownerID).
		SetActive(true).
		Save(ctx)
	assert.NilError(t, err)

	return definition
}

func createControlForWorkflow(t *testing.T, ctx context.Context, ownerID string) *ent.Control {
	t.Helper()

	control, err := suite.Client.DB.Control.Create().
		SetRefCode("CTL-" + ulids.New().String()).
		SetTitle("Test Control").
		SetStatus(enums.ControlStatusNotImplemented).
		SetOwnerID(ownerID).
		Save(ctx)
	assert.NilError(t, err)

	return control
}

func createWorkflowInstance(t *testing.T, ctx context.Context, ownerID string, definitionID string, control *ent.Control) *ent.WorkflowInstance {
	t.Helper()

	instance, err := suite.Client.DB.WorkflowInstance.Create().
		SetWorkflowDefinitionID(definitionID).
		SetOwnerID(ownerID).
		SetState(enums.WorkflowInstanceStatePaused).
		SetContext(models.WorkflowInstanceContext{
			ObjectType: enums.WorkflowObjectTypeControl,
			ObjectID:   control.ID,
		}).
		Save(ctx)
	assert.NilError(t, err)

	return instance
}

func createWorkflowAssignmentWithTarget(t *testing.T, ctx context.Context, ownerID string, instanceID string, targetUserID string) *ent.WorkflowAssignment {
	t.Helper()

	actionKey := "action_" + ulids.New().String()

	assignment, err := suite.Client.DB.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instanceID).
		SetAssignmentKey("approval_" + actionKey + "_" + ulids.New().String()).
		SetApprovalMetadata(models.WorkflowAssignmentApproval{
			ActionKey: actionKey,
			Required:  true,
		}).
		SetOwnerID(ownerID).
		Save(ctx)
	assert.NilError(t, err)

	err = suite.Client.DB.WorkflowAssignmentTarget.Create().
		SetWorkflowAssignmentID(assignment.ID).
		SetTargetType(enums.WorkflowTargetTypeUser).
		SetTargetUserID(targetUserID).
		SetOwnerID(ownerID).
		Exec(ctx)
	assert.NilError(t, err)

	return assignment
}

func createWorkflowProposal(t *testing.T, ctx context.Context, ownerID string, instance *ent.WorkflowInstance, control *ent.Control, domainKey string, changes map[string]any) *ent.WorkflowProposal {
	t.Helper()

	objRef, err := suite.Client.DB.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(ownerID).
		Save(ctx)
	assert.NilError(t, err)

	proposal, err := suite.Client.DB.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey(domainKey).
		SetChanges(changes).
		SetOwnerID(ownerID).
		Save(ctx)
	assert.NilError(t, err)

	return proposal
}

func TestRequestChangesWorkflowAssignment(t *testing.T) {
	ensureWorkflowEngine(t)
	t.Parallel()

	user := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := th.SetContext(user.UserCtx, suite.Client.DB)

	resolver := graphapi.NewResolver(suite.Client.DB, nil)

	control := createControlForWorkflow(t, ctx, user.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, user.OrganizationID)
	instance := createWorkflowInstance(t, ctx, user.OrganizationID, definition.ID, control)
	assignment := createWorkflowAssignmentWithTarget(t, ctx, user.OrganizationID, instance.ID, user.ID)

	reason := "needs more info"
	inputs := map[string]any{"status": "in_review"}

	res, err := resolver.Mutation().RequestChangesWorkflowAssignment(th.SetUserContext(user.UserCtx, suite.Client.DB), assignment.ID, &reason, inputs)
	assert.NilError(t, err)
	assert.Check(t, res != nil)
	assert.Check(t, res.WorkflowAssignment != nil)

	updated := res.WorkflowAssignment
	assert.Check(t, is.Equal(updated.Status, enums.WorkflowAssignmentStatusChangesRequested))
	assert.Check(t, is.Equal(updated.ActorUserID, user.ID))
	assert.Check(t, updated.DecidedAt != nil)
	assert.Check(t, is.Equal(updated.Notes, reason))

	meta := updated.Metadata
	assert.Check(t, meta != nil)
	assert.Check(t, is.Equal(meta["change_reason"], reason))
	assert.Check(t, is.Equal(meta["change_requested_by"], user.ID))

	inputsVal, ok := meta["change_inputs"].(map[string]any)
	assert.Check(t, ok)
	if ok {
		assert.Check(t, is.Equal(inputsVal["status"], "in_review"))
	}

	rejection := updated.RejectionMetadata
	assert.Check(t, rejection.RejectedAt != "")
	assert.Check(t, is.Equal(rejection.RejectedByUserID, user.ID))
	assert.Check(t, is.Equal(rejection.RejectionReason, reason))
	assert.Check(t, rejection.ActionKey != "")
	inputsVal = rejection.ChangeRequestInputs
	assert.Check(t, inputsVal != nil)
	if inputsVal != nil {
		assert.Check(t, is.Equal(inputsVal["status"], "in_review"))
	}
}

func TestReassignWorkflowAssignment(t *testing.T) {
	ensureWorkflowEngine(t)
	t.Parallel()

	owner := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	orgMember := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	suite.AddUserToOrganization(owner.UserCtx, t, &orgMember, enums.RoleAdmin, owner.OrganizationID)
	outsider := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)

	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)
	resolver := graphapi.NewResolver(suite.Client.DB, nil)

	control := createControlForWorkflow(t, ctx, owner.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, owner.OrganizationID)
	instance := createWorkflowInstance(t, ctx, owner.OrganizationID, definition.ID, control)

	testCases := []struct {
		name         string
		targetUserID string
		expectedErr  string
	}{
		{
			name:         "happy path, target in the organization",
			targetUserID: orgMember.ID,
		},
		{
			name:         "target outside the organization is not found",
			targetUserID: outsider.ID,
			expectedErr:  th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assignment := createWorkflowAssignmentWithTarget(t, ctx, owner.OrganizationID, instance.ID, owner.ID)

			updated, err := resolver.Mutation().ReassignWorkflowAssignment(th.SetUserContext(owner.UserCtx, suite.Client.DB), assignment.ID, tc.targetUserID)

			exists, existsErr := suite.Client.DB.WorkflowAssignmentTarget.Query().
				Where(
					workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID),
					workflowassignmenttarget.TargetUserIDEQ(tc.targetUserID),
				).
				Exist(ctx)
			assert.NilError(t, existsErr)

			if tc.expectedErr != "" {
				assert.Check(t, is.ErrorContains(err, tc.expectedErr))
				assert.Check(t, !exists)

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(updated.Status, enums.WorkflowAssignmentStatusPending))
			assert.Check(t, exists)
		})
	}
}

func TestAdminReassignWorkflowAssignment(t *testing.T) {
	ensureWorkflowEngine(t)
	t.Parallel()

	owner := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	oldTarget := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	newTarget := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	suite.AddUserToOrganization(owner.UserCtx, t, &oldTarget, enums.RoleAdmin, owner.OrganizationID)
	suite.AddUserToOrganization(owner.UserCtx, t, &newTarget, enums.RoleAdmin, owner.OrganizationID)

	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)
	resolver := graphapi.NewResolver(suite.Client.DB, nil)

	control := createControlForWorkflow(t, ctx, owner.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, owner.OrganizationID)
	instance := createWorkflowInstance(t, ctx, owner.OrganizationID, definition.ID, control)
	assignment := createWorkflowAssignmentWithTarget(t, ctx, owner.OrganizationID, instance.ID, oldTarget.ID)

	_, err := suite.Client.DB.WorkflowAssignment.UpdateOneID(assignment.ID).
		SetStatus(enums.WorkflowAssignmentStatusRejected).
		SetDecidedAt(time.Now()).
		SetActorUserID(oldTarget.ID).
		SetRejectionMetadata(models.WorkflowAssignmentRejection{
			RejectionReason: "not good",
		}).
		Save(ctx)
	assert.NilError(t, err)

	targetID := newTarget.ID
	input := model.ReassignWorkflowAssignmentInput{
		ID: assignment.ID,
		Targets: []*model.WorkflowAssignmentTargetInput{
			{
				Type: enums.WorkflowTargetTypeUser,
				ID:   &targetID,
			},
		},
	}

	res, err := resolver.Mutation().AdminReassignWorkflowAssignment(th.SetUserContext(owner.UserCtx, suite.Client.DB), input)
	assert.NilError(t, err)
	assert.Check(t, res != nil)
	assert.Check(t, res.WorkflowAssignment != nil)

	updated := res.WorkflowAssignment
	assert.Check(t, is.Equal(updated.Status, enums.WorkflowAssignmentStatusPending))
	assert.Check(t, updated.DecidedAt == nil)
	assert.Check(t, is.Equal(updated.ActorUserID, ""))
	assert.Check(t, is.Equal(updated.RejectionMetadata.RejectionReason, ""))

	oldExists, err := suite.Client.DB.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID),
			workflowassignmenttarget.TargetUserIDEQ(oldTarget.ID),
		).
		Exist(ctx)
	assert.NilError(t, err)
	assert.Check(t, !oldExists)

	newExists, err := suite.Client.DB.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID),
			workflowassignmenttarget.TargetUserIDEQ(newTarget.ID),
		).
		Exist(ctx)
	assert.NilError(t, err)
	assert.Check(t, newExists)
}

func TestWorkflowProposalSubmitAndWithdraw(t *testing.T) {
	ensureWorkflowEngine(t)

	user := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := th.SetContext(user.UserCtx, suite.Client.DB)

	resolver := graphapi.NewResolver(suite.Client.DB, nil)

	control := createControlForWorkflow(t, ctx, user.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, user.OrganizationID)
	instance := createWorkflowInstance(t, ctx, user.OrganizationID, definition.ID, control)

	changes := map[string]any{"status": string(enums.ControlStatusApproved)}
	proposal := createWorkflowProposal(t, ctx, user.OrganizationID, instance, control, "Control:status", changes)

	submitRes, err := resolver.Mutation().SubmitWorkflowProposal(th.SetUserContext(user.UserCtx, suite.Client.DB), proposal.ID)
	assert.NilError(t, err)
	assert.Check(t, submitRes != nil)
	assert.Check(t, submitRes.WorkflowProposal != nil)
	assert.Check(t, is.Equal(submitRes.WorkflowProposal.State, enums.WorkflowProposalStateSubmitted))
	assert.Check(t, submitRes.WorkflowProposal.SubmittedAt != nil)
	assert.Check(t, is.Equal(submitRes.WorkflowProposal.SubmittedByUserID, user.ID))
	assert.Check(t, submitRes.WorkflowProposal.ProposedHash != "")

	withdrawRes, err := resolver.Mutation().WithdrawWorkflowProposal(th.SetUserContext(user.UserCtx, suite.Client.DB), proposal.ID, nil)
	assert.NilError(t, err)
	assert.Check(t, withdrawRes != nil)
	assert.Check(t, withdrawRes.WorkflowProposal != nil)
	assert.Check(t, is.Equal(withdrawRes.WorkflowProposal.State, enums.WorkflowProposalStateSuperseded))
}

func TestWorkflowProposalPreview(t *testing.T) {
	ensureWorkflowEngine(t)
	t.Parallel()

	user := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := th.SetContext(user.UserCtx, suite.Client.DB)

	resolver := graphapi.NewResolver(suite.Client.DB, nil)

	control := createControlForWorkflow(t, ctx, user.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, user.OrganizationID)
	instance := createWorkflowInstance(t, ctx, user.OrganizationID, definition.ID, control)

	changes := map[string]any{"status": string(enums.ControlStatusApproved)}
	proposal := createWorkflowProposal(t, ctx, user.OrganizationID, instance, control, "Control:status", changes)

	_, err := suite.Client.DB.WorkflowInstance.UpdateOneID(instance.ID).
		SetWorkflowProposalID(proposal.ID).
		Save(ctx)
	assert.NilError(t, err)

	_ = createWorkflowAssignmentWithTarget(t, ctx, user.OrganizationID, instance.ID, user.ID)

	preview, err := resolver.WorkflowProposal().Preview(th.SetUserContext(user.UserCtx, suite.Client.DB), proposal)
	assert.NilError(t, err)
	assert.Check(t, preview != nil)
	assert.Check(t, is.Equal(preview.ProposalID, proposal.ID))
	assert.Check(t, is.Equal(preview.DomainKey, proposal.DomainKey))
	assert.Check(t, preview.Diffs != nil)
	assert.Check(t, len(preview.Diffs) > 0)
}

func TestWorkflowProposalAccess(t *testing.T) {
	ensureWorkflowEngine(t)
	t.Parallel()

	owner := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	member := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	suite.AddUserToOrganization(owner.UserCtx, t, &member, enums.RoleMember, owner.OrganizationID)
	otherOrgUser := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)

	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)
	resolver := graphapi.NewResolver(suite.Client.DB, nil)

	control := createControlForWorkflow(t, ctx, owner.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, owner.OrganizationID)
	instance := createWorkflowInstance(t, ctx, owner.OrganizationID, definition.ID, control)
	proposal := createWorkflowProposal(t, ctx, owner.OrganizationID, instance, control, "Control:status", map[string]any{"status": string(enums.ControlStatusApproved)})

	testCases := []struct {
		name        string
		ctx         context.Context
		proposalID  string
		submit      bool
		expectedErr string
	}{
		{
			name:       "happy path, owner views the proposal",
			ctx:        th.SetUserContext(owner.UserCtx, suite.Client.DB),
			proposalID: proposal.ID,
		},
		{
			name:        "member who is not an editor or approver is denied",
			ctx:         th.SetUserContext(member.UserCtx, suite.Client.DB),
			proposalID:  proposal.ID,
			expectedErr: rout.ErrPermissionDenied.Error(),
		},
		{
			name:        "other organization view is not found",
			ctx:         th.SetUserContext(otherOrgUser.UserCtx, suite.Client.DB),
			proposalID:  proposal.ID,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "other organization submit is not found",
			ctx:         th.SetUserContext(otherOrgUser.UserCtx, suite.Client.DB),
			proposalID:  proposal.ID,
			submit:      true,
			expectedErr: th.NotFoundErrorMsg,
		},
		{
			name:        "nonexistent proposal submit is not found",
			ctx:         th.SetUserContext(otherOrgUser.UserCtx, suite.Client.DB),
			proposalID:  ulids.New().String(),
			submit:      true,
			expectedErr: th.NotFoundErrorMsg,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				found *ent.WorkflowProposal
				err   error
			)

			if tc.submit {
				_, err = resolver.Mutation().SubmitWorkflowProposal(tc.ctx, tc.proposalID)
			} else {
				found, err = resolver.Query().WorkflowProposal(tc.ctx, tc.proposalID)
			}

			if tc.expectedErr != "" {
				assert.Check(t, is.ErrorContains(err, tc.expectedErr))
				assert.Check(t, found == nil)

				return
			}

			assert.NilError(t, err)
			assert.Check(t, is.Equal(tc.proposalID, found.ID))
		})
	}
}

func TestPreCommitApprovalRoutesMemberUpdateToProposal(t *testing.T) {
	ensureWorkflowEngine(t)

	owner := suite.UserBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	complianceManager := suite.OrgMemberWithFunctionalRoles(t, owner, "compliance_manager")

	ctx := th.SetContext(owner.UserCtx, suite.Client.DB)

	definition := createPreCommitApprovalDefinition(t, ctx, owner.OrganizationID, owner.ID)
	control := createControlForWorkflow(t, ctx, owner.OrganizationID)

	resp, err := suite.Client.API.UpdateControl(complianceManager.UserCtx, control.ID, testclient.UpdateControlInput{
		Status: &enums.ControlStatusApproved,
	})
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.ControlStatusNotImplemented, *resp.UpdateControl.Control.Status))

	reloaded, err := suite.Client.DB.Control.Get(ctx, control.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(enums.ControlStatusNotImplemented, reloaded.Status))

	instance, err := suite.Client.DB.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(definition.ID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Only(ctx)
	assert.NilError(t, err)
	assert.Assert(t, instance.WorkflowProposalID != "")

	proposal, err := suite.Client.DB.WorkflowProposal.Get(ctx, instance.WorkflowProposalID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(string(enums.ControlStatusApproved), proposal.Changes["status"]))
}

func createPreCommitApprovalDefinition(t *testing.T, ctx context.Context, ownerID, approverID string) *ent.WorkflowDefinition {
	t.Helper()

	params, err := json.Marshal(workflows.ApprovalActionParams{
		TargetedActionParams: workflows.TargetedActionParams{
			Targets: []workflows.TargetConfig{{Type: enums.WorkflowTargetTypeUser, ID: approverID}},
		},
		Label:  "Status Approval",
		Fields: []string{"status"},
	})
	assert.NilError(t, err)

	doc := models.WorkflowDefinitionDocument{
		ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
		ApprovalTiming:         enums.WorkflowApprovalTimingPreCommit,
		Triggers:               []models.WorkflowTrigger{{Operation: "UPDATE", Fields: []string{"status"}}},
		Conditions:             []models.WorkflowCondition{{Expression: "true"}},
		Actions: []models.WorkflowAction{{
			Type:   enums.WorkflowActionTypeApproval.String(),
			Key:    "status_pre_commit",
			Params: params,
		}},
	}

	operations, fields := workflows.DeriveTriggerPrefilter(doc)

	definition, err := suite.Client.DB.WorkflowDefinition.Create().
		SetName("Pre Commit Approval " + ulids.New().String()).
		SetWorkflowKind(enums.WorkflowKindApproval).
		SetSchemaType("Control").
		SetActive(true).
		SetDraft(false).
		SetOwnerID(ownerID).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		SetDefinitionJSON(doc).
		Save(ctx)
	assert.NilError(t, err)

	return definition
}
