//go:build test

package engine_test

import (
	"testing"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/v2/internal/workflows"
	"github.com/theopenlane/utils/ulids"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

// TestProposalManagerComputeHashUsesDomainKey ensures that when multiple approval workflows
// trigger for the same object, each workflow's approval assignments reference the correct
// proposal hash for their specific domain (field set), not a combined hash.
//
// Workflow Definitions (Plain English):
//
//	Definition 1: "Require approval for Control.reference_id changes"
//	Definition 2: "Require approval for Control.title changes"
//
// Test Flow:
//  1. Creates two approval workflow definitions targeting different fields
//  2. Creates a Control with initial reference_id and title
//  3. Updates BOTH reference_id AND title in a single mutation
//  4. Verifies two separate workflow instances were created (one per definition)
//  5. Each instance has its own proposal with domain-specific changes
//  6. Submits the reference_id proposal
//  7. Verifies the resulting assignments have a ProposedHash matching the reference_id proposal's changes
//
// Why This Matters:
//
//	Different approval workflows may cover different field sets. Each workflow's proposal
//	should only contain the changes relevant to its domain, and the hash used for approval
//	verification should match that scoped set of changes. This prevents cross-contamination
//	of approval hashes between unrelated workflows.
func TestProposalManagerComputeHashUsesDomainKey(t *testing.T) {
	s := suite.setupTest(t)

	userID, orgID, _ := s.SetupTestUser()
	seedCtx := s.SeedContext(userID, orgID)

	buildApprovalAction := func(key string, fields []string) models.WorkflowAction {
		params := workflows.ApprovalActionParams{
			TargetedActionParams: workflows.TargetedActionParams{
				Targets: []workflows.TargetConfig{
					{Type: enums.WorkflowTargetTypeUser, ID: userID},
				},
			},
			Required: boolPtr(true),
			Label:    "Approval",
			Fields:   fields,
		}
		paramsBytes, err := json.Marshal(params)
		assert.NilError(t, err)

		return models.WorkflowAction{
			Type:   enums.WorkflowActionTypeApproval.String(),
			Key:    key,
			Params: paramsBytes,
		}
	}

	refAction := buildApprovalAction("reference_id_approval", []string{"reference_id"})
	titleAction := buildApprovalAction("title_approval", []string{"title"})

	buildDefinition := func(name string, triggerField string, action models.WorkflowAction) *generated.WorkflowDefinition {
		doc := models.WorkflowDefinitionDocument{
			ApprovalSubmissionMode: enums.WorkflowApprovalSubmissionModeAutoSubmit,
			Triggers: []models.WorkflowTrigger{
				{Operation: "UPDATE", Fields: []string{triggerField}},
			},
			Conditions: []models.WorkflowCondition{
				{Expression: "true"},
			},
			Actions: []models.WorkflowAction{action},
		}

		operations, fields := workflows.DeriveTriggerPrefilter(doc)

		def, err := s.client.WorkflowDefinition.Create().
			SetName(name + " " + ulids.New().String()).
			SetWorkflowKind(enums.WorkflowKindApproval).
			SetSchemaType("Control").
			SetActive(true).
			SetDraft(false).
			SetOwnerID(orgID).
			SetTriggerOperations(operations).
			SetTriggerFields(fields).
			SetApprovalSubmissionMode(enums.WorkflowApprovalSubmissionModeAutoSubmit).
			SetDefinitionJSON(doc).
			Save(seedCtx)
		assert.NilError(t, err)

		return def
	}

	refDef := buildDefinition("Reference ID Approval", "reference_id", refAction)
	titleDef := buildDefinition("Title Approval", "title", titleAction)

	oldRef := "REF-OLD-" + ulid.Make().String()
	newRef := "REF-NEW-" + ulid.Make().String()
	oldTitle := "Old Title " + ulid.Make().String()
	newTitle := "New Title " + ulid.Make().String()

	control, err := s.client.Control.Create().
		SetRefCode("CTL-HASH-" + ulid.Make().String()).
		SetOwnerID(orgID).
		SetReferenceID(oldRef).
		SetTitle(oldTitle).
		Save(seedCtx)
	assert.NilError(t, err)

	updated, err := s.client.Control.UpdateOneID(control.ID).
		SetReferenceID(newRef).
		SetTitle(newTitle).
		Save(seedCtx)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(oldRef, updated.ReferenceID))
	assert.Check(t, is.DeepEqual(oldTitle, updated.Title))

	refInstance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(refDef.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(refInstance.WorkflowProposalID) > 0)

	refProposal, err := s.client.WorkflowProposal.Get(seedCtx, refInstance.WorkflowProposalID)
	assert.NilError(t, err)

	titleInstance, err := s.client.WorkflowInstance.Query().
		Where(
			workflowinstance.WorkflowDefinitionIDEQ(titleDef.ID),
			workflowinstance.OwnerIDEQ(orgID),
			workflowinstance.ControlIDEQ(control.ID),
		).
		Order(generated.Desc(workflowinstance.FieldCreatedAt)).
		First(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(titleInstance.WorkflowProposalID) > 0)

	_, err = s.client.WorkflowProposal.Get(seedCtx, titleInstance.WorkflowProposalID)
	assert.NilError(t, err)

	_, err = s.client.WorkflowProposal.UpdateOneID(refProposal.ID).
		SetState(enums.WorkflowProposalStateSubmitted).
		SetSubmittedAt(time.Now()).
		SetSubmittedByUserID(userID).
		Save(seedCtx)
	assert.NilError(t, err)

	s.WaitForEvents()

	instance, err := s.client.WorkflowInstance.Query().
		Where(workflowinstance.WorkflowProposalIDEQ(refProposal.ID)).
		Only(seedCtx)
	assert.NilError(t, err)

	assignments, err := s.client.WorkflowAssignment.Query().
		Where(workflowassignment.WorkflowInstanceIDEQ(instance.ID)).
		All(seedCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(assignments) > 0)

	expectedHash, err := workflows.ComputeProposalHash(refProposal.Changes)
	assert.NilError(t, err)

	for _, assignment := range assignments {
		assert.Check(t, is.DeepEqual(expectedHash, assignment.ApprovalMetadata.ProposedHash))
	}
}
