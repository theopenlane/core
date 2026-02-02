package graphapi_test

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/graphapi"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/internal/workflows/engine"
	"github.com/theopenlane/utils/ulids"
)

func ensureWorkflowEngine(t *testing.T) {
	t.Helper()

	prev := suite.client.db.WorkflowEngine
	if prev == nil {
		workflowEngine, err := engine.NewWorkflowEngine(suite.client.db, nil)
		requireNoError(t, err)
		suite.client.db.WorkflowEngine = workflowEngine
	}

	t.Cleanup(func() {
		suite.client.db.WorkflowEngine = prev
	})
}

func createWorkflowDefinition(t *testing.T, ctx context.Context, ownerID string) *ent.WorkflowDefinition {
	t.Helper()

	definition, err := suite.client.db.WorkflowDefinition.Create().
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

	control, err := suite.client.db.Control.Create().
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

	instance, err := suite.client.db.WorkflowInstance.Create().
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

	assignment, err := suite.client.db.WorkflowAssignment.Create().
		SetWorkflowInstanceID(instanceID).
		SetAssignmentKey("assignment-" + ulids.New().String()).
		SetOwnerID(ownerID).
		Save(ctx)
	assert.NilError(t, err)

	err = suite.client.db.WorkflowAssignmentTarget.Create().
		SetWorkflowAssignmentID(assignment.ID).
		SetTargetType(enums.WorkflowTargetTypeUser).
		SetTargetUserID(targetUserID).
		SetOwnerID(ownerID).
		Exec(ctx)
	assert.NilError(t, err)

	return assignment
}

func createWorkflowProposal(t *testing.T, ctx context.Context, ownerID string, instance *ent.WorkflowInstance, control *ent.Control, changes map[string]any) *ent.WorkflowProposal {
	t.Helper()

	objRef, err := suite.client.db.WorkflowObjectRef.Create().
		SetWorkflowInstanceID(instance.ID).
		SetControlID(control.ID).
		SetOwnerID(ownerID).
		Save(ctx)
	assert.NilError(t, err)

	proposal, err := suite.client.db.WorkflowProposal.Create().
		SetWorkflowObjectRefID(objRef.ID).
		SetDomainKey("status").
		SetChanges(changes).
		SetOwnerID(ownerID).
		Save(ctx)
	assert.NilError(t, err)

	return proposal
}

func TestRequestChangesWorkflowAssignment(t *testing.T) {
	ensureWorkflowEngine(t)

	user := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := setContext(user.UserCtx, suite.client.db)

	resolver := graphapi.NewResolver(suite.client.db, nil)

	control := createControlForWorkflow(t, ctx, user.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, user.OrganizationID)
	instance := createWorkflowInstance(t, ctx, user.OrganizationID, definition.ID, control)
	assignment := createWorkflowAssignmentWithTarget(t, ctx, user.OrganizationID, instance.ID, user.ID)

	reason := "needs more info"
	inputs := map[string]any{"status": "in_review"}

	res, err := resolver.Mutation().RequestChangesWorkflowAssignment(ctx, assignment.ID, &reason, inputs)
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
}

func TestReassignWorkflowAssignment(t *testing.T) {
	ensureWorkflowEngine(t)

	owner := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	newTarget := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	suite.addUserToOrganization(owner.UserCtx, t, &newTarget, enums.RoleAdmin, owner.OrganizationID)

	ctx := setContext(owner.UserCtx, suite.client.db)
	resolver := graphapi.NewResolver(suite.client.db, nil)

	control := createControlForWorkflow(t, ctx, owner.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, owner.OrganizationID)
	instance := createWorkflowInstance(t, ctx, owner.OrganizationID, definition.ID, control)
	assignment := createWorkflowAssignmentWithTarget(t, ctx, owner.OrganizationID, instance.ID, owner.ID)

	updated, err := resolver.Mutation().ReassignWorkflowAssignment(ctx, assignment.ID, newTarget.ID)
	assert.NilError(t, err)
	assert.Check(t, updated != nil)
	assert.Check(t, is.Equal(updated.Status, enums.WorkflowAssignmentStatusPending))

	exists, err := suite.client.db.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID),
			workflowassignmenttarget.TargetUserIDEQ(newTarget.ID),
		).
		Exist(ctx)
	assert.NilError(t, err)
	assert.Check(t, exists)
}

func TestAdminReassignWorkflowAssignment(t *testing.T) {
	ensureWorkflowEngine(t)

	owner := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	oldTarget := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	newTarget := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	suite.addUserToOrganization(owner.UserCtx, t, &oldTarget, enums.RoleAdmin, owner.OrganizationID)
	suite.addUserToOrganization(owner.UserCtx, t, &newTarget, enums.RoleAdmin, owner.OrganizationID)

	ctx := setContext(owner.UserCtx, suite.client.db)
	resolver := graphapi.NewResolver(suite.client.db, nil)

	control := createControlForWorkflow(t, ctx, owner.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, owner.OrganizationID)
	instance := createWorkflowInstance(t, ctx, owner.OrganizationID, definition.ID, control)
	assignment := createWorkflowAssignmentWithTarget(t, ctx, owner.OrganizationID, instance.ID, oldTarget.ID)

	_, err := suite.client.db.WorkflowAssignment.UpdateOneID(assignment.ID).
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

	res, err := resolver.Mutation().AdminReassignWorkflowAssignment(ctx, input)
	assert.NilError(t, err)
	assert.Check(t, res != nil)
	assert.Check(t, res.WorkflowAssignment != nil)

	updated := res.WorkflowAssignment
	assert.Check(t, is.Equal(updated.Status, enums.WorkflowAssignmentStatusPending))
	assert.Check(t, updated.DecidedAt == nil)
	assert.Check(t, is.Equal(updated.ActorUserID, ""))
	assert.Check(t, is.Equal(updated.RejectionMetadata.RejectionReason, ""))

	oldExists, err := suite.client.db.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDEQ(assignment.ID),
			workflowassignmenttarget.TargetUserIDEQ(oldTarget.ID),
		).
		Exist(ctx)
	assert.NilError(t, err)
	assert.Check(t, !oldExists)

	newExists, err := suite.client.db.WorkflowAssignmentTarget.Query().
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

	user := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := setContext(user.UserCtx, suite.client.db)

	resolver := graphapi.NewResolver(suite.client.db, nil)

	control := createControlForWorkflow(t, ctx, user.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, user.OrganizationID)
	instance := createWorkflowInstance(t, ctx, user.OrganizationID, definition.ID, control)

	changes := map[string]any{"status": string(enums.ControlStatusApproved)}
	proposal := createWorkflowProposal(t, ctx, user.OrganizationID, instance, control, changes)

	submitRes, err := resolver.Mutation().SubmitWorkflowProposal(ctx, proposal.ID)
	assert.NilError(t, err)
	assert.Check(t, submitRes != nil)
	assert.Check(t, submitRes.WorkflowProposal != nil)
	assert.Check(t, is.Equal(submitRes.WorkflowProposal.State, enums.WorkflowProposalStateSubmitted))
	assert.Check(t, submitRes.WorkflowProposal.SubmittedAt != nil)
	assert.Check(t, is.Equal(submitRes.WorkflowProposal.SubmittedByUserID, user.ID))
	assert.Check(t, submitRes.WorkflowProposal.ProposedHash != "")

	withdrawRes, err := resolver.Mutation().WithdrawWorkflowProposal(ctx, proposal.ID, nil)
	assert.NilError(t, err)
	assert.Check(t, withdrawRes != nil)
	assert.Check(t, withdrawRes.WorkflowProposal != nil)
	assert.Check(t, is.Equal(withdrawRes.WorkflowProposal.State, enums.WorkflowProposalStateSuperseded))
}

func TestWorkflowProposalPreview(t *testing.T) {
	ensureWorkflowEngine(t)

	user := suite.userBuilder(context.Background(), t, models.CatalogBaseModule, models.CatalogComplianceModule)
	ctx := setContext(user.UserCtx, suite.client.db)

	resolver := graphapi.NewResolver(suite.client.db, nil)

	control := createControlForWorkflow(t, ctx, user.OrganizationID)
	definition := createWorkflowDefinition(t, ctx, user.OrganizationID)
	instance := createWorkflowInstance(t, ctx, user.OrganizationID, definition.ID, control)

	changes := map[string]any{"status": string(enums.ControlStatusApproved)}
	proposal := createWorkflowProposal(t, ctx, user.OrganizationID, instance, control, changes)

	_, err := suite.client.db.WorkflowInstance.UpdateOneID(instance.ID).
		SetWorkflowProposalID(proposal.ID).
		Save(ctx)
	assert.NilError(t, err)

	_ = createWorkflowAssignmentWithTarget(t, ctx, user.OrganizationID, instance.ID, user.ID)

	preview, err := resolver.WorkflowProposal().Preview(ctx, proposal)
	assert.NilError(t, err)
	assert.Check(t, preview != nil)
	assert.Check(t, is.Equal(preview.ProposalID, proposal.ID))
	assert.Check(t, is.Equal(preview.DomainKey, proposal.DomainKey))
	assert.Check(t, preview.Diffs != nil)
	assert.Check(t, len(preview.Diffs) > 0)
}
