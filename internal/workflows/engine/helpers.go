package engine

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/workflows"
)

// computeHMACSignature computes an HMAC-SHA256 signature for webhook authentication
func computeHMACSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)

	return hex.EncodeToString(mac.Sum(nil))
}

// applyStringTemplates walks arbitrary maps/slices and replaces {{key}} tokens in string values
func applyStringTemplates(input map[string]any, replacements map[string]string) map[string]any {
	if input == nil {
		return map[string]any{}
	}

	out := make(map[string]any, len(input))
	for k, v := range input {
		out[k] = replaceInValue(v, replacements)
	}
	return out
}

// replaceInValue recursively replaces tokens in nested values
func replaceInValue(v any, replacements map[string]string) any {
	switch val := v.(type) {
	case string:
		return replaceTokens(val, replacements)
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, nested := range val {
			result[k] = replaceInValue(nested, replacements)
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, nested := range val {
			result[i] = replaceInValue(nested, replacements)
		}
		return result
	default:
		return v
	}
}

// replaceTokens replaces {{key}} tokens in a string with their replacement values
func replaceTokens(s string, replacements map[string]string) string {
	out := s
	for k, v := range replacements {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

// loadWorkflowInstance loads a workflow instance by ID and organization ID
func loadWorkflowInstance(ctx context.Context, client *generated.Client, instanceID, orgID string) (*generated.WorkflowInstance, error) {
	return client.WorkflowInstance.Query().Where(workflowinstance.IDEQ(instanceID), workflowinstance.OwnerIDEQ(orgID)).Only(ctx)
}

// loadWorkflowObjectRef loads a workflow object reference by instance ID and organization ID
func loadWorkflowObjectRef(ctx context.Context, client *generated.Client, instanceID, orgID string) (*generated.WorkflowObjectRef, error) {
	return client.WorkflowObjectRef.Query().Where(workflowobjectref.WorkflowInstanceIDEQ(instanceID), workflowobjectref.OwnerIDEQ(orgID)).First(ctx)
}

// assignmentTargetUserIDs collects unique user IDs for assignment targets
func assignmentTargetUserIDs(ctx context.Context, client *generated.Client, assignmentIDs []string, orgID string) ([]string, error) {
	if len(assignmentIDs) == 0 {
		return nil, nil
	}

	targets, err := client.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.WorkflowAssignmentIDIn(assignmentIDs...),
			workflowassignmenttarget.OwnerIDEQ(orgID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	userIDs := lo.Uniq(lo.FilterMap(targets, func(target *generated.WorkflowAssignmentTarget, _ int) (string, bool) {
		if target.TargetUserID == "" {
			return "", false
		}
		return target.TargetUserID, true
	}))

	return userIDs, nil
}

// AssignmentStatusCounts holds counts of assignment statuses for quorum evaluation
type AssignmentStatusCounts struct {
	// Approved is the count of approved assignments
	Approved int
	// Pending is the count of pending assignments
	Pending int
	// Rejected is the count of rejected assignments
	Rejected int
	// ChangesRequested is the count of changes requested assignments
	ChangesRequested int
	// RejectedRequired indicates if any required assignment was rejected
	RejectedRequired bool
	// ChangesRequestedRequired indicates if any required assignment requested changes
	ChangesRequestedRequired bool
}

// approvalResolution captures how approvals should advance a workflow
type approvalResolution int

const (
	// approvalPending indicates approvals are still outstanding
	approvalPending approvalResolution = iota
	// approvalSatisfied indicates approvals satisfy quorum
	approvalSatisfied
	// approvalFailed indicates approvals failed due to rejection or lack of quorum
	approvalFailed
)

// CountAssignmentStatus counts the status of assignments for quorum evaluation
func CountAssignmentStatus(assignments []*generated.WorkflowAssignment) AssignmentStatusCounts {
	counts := AssignmentStatusCounts{}

	for _, a := range assignments {
		switch a.Status {
		case enums.WorkflowAssignmentStatusApproved:
			counts.Approved++
		case enums.WorkflowAssignmentStatusPending:
			counts.Pending++
		case enums.WorkflowAssignmentStatusRejected:
			counts.Rejected++
			if a.Required {
				counts.RejectedRequired = true
			}
		case enums.WorkflowAssignmentStatusChangesRequested:
			counts.ChangesRequested++
			if a.Required {
				counts.ChangesRequestedRequired = true
			}
		}
	}

	return counts
}

// isGatedActionType returns true if the action type requires approval gating
func isGatedActionType(actionType enums.WorkflowActionType) bool {
	return actionType == enums.WorkflowActionTypeApproval || actionType == enums.WorkflowActionTypeReview
}

// actionIndexForKey returns the index of the action matching the key
func actionIndexForKey(actions []models.WorkflowAction, key string) int {
	for i, action := range actions {
		if action.Key == key {
			return i
		}
	}

	return -1
}

// resolveApproval determines whether approvals are pending, satisfied, or failed
func resolveApproval(requiredCount int, statusCounts AssignmentStatusCounts) approvalResolution {
	if statusCounts.RejectedRequired || statusCounts.ChangesRequestedRequired {
		return approvalFailed
	}

	if requiredCount > 0 {
		if statusCounts.Approved < requiredCount {
			if statusCounts.Pending > 0 {
				return approvalPending
			}
			return approvalFailed
		}
		return approvalSatisfied
	}

	if statusCounts.Pending > 0 {
		return approvalPending
	}

	return approvalSatisfied
}

// buildTriggerContext constructs the workflow instance context for a new trigger
func buildTriggerContext(defID string, obj *workflows.Object, input TriggerInput, userID string) models.WorkflowInstanceContext {
	return models.WorkflowInstanceContext{
		WorkflowDefinitionID:   defID,
		ObjectType:             obj.Type,
		ObjectID:               obj.ID,
		Version:                1,
		Assignments:            []models.WorkflowAssignmentContext{},
		TriggerEventType:       input.EventType,
		TriggerChangedFields:   input.ChangedFields,
		TriggerChangedEdges:    input.ChangedEdges,
		TriggerAddedIDs:        input.AddedIDs,
		TriggerRemovedIDs:      input.RemovedIDs,
		TriggerUserID:          userID,
		TriggerProposedChanges: input.ProposedChanges,
		Data:                   nil,
	}
}

// applyTriggerContext updates an existing instance context with trigger metadata
func applyTriggerContext(existing models.WorkflowInstanceContext, defID string, obj *workflows.Object, input TriggerInput, userID string) models.WorkflowInstanceContext {
	if existing.Version == 0 {
		existing.Version = 1
	}
	existing.WorkflowDefinitionID = defID
	existing.ObjectType = obj.Type
	existing.ObjectID = obj.ID
	existing.TriggerEventType = input.EventType
	existing.TriggerChangedFields = input.ChangedFields
	existing.TriggerChangedEdges = input.ChangedEdges
	existing.TriggerAddedIDs = input.AddedIDs
	existing.TriggerRemovedIDs = input.RemovedIDs
	existing.TriggerUserID = userID
	existing.TriggerProposedChanges = input.ProposedChanges

	return existing
}
