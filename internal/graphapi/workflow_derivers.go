package graphapi

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/workflows"
)

// workflowEligibleFieldsForObjectType returns the set of fields eligible for workflow operations on the given object type
func workflowEligibleFieldsForObjectType(objectType enums.WorkflowObjectType) map[string]struct{} {
	return workflows.EligibleWorkflowFields(objectType)
}

// deriveWorkflowDefinitionPrefilter extracts trigger operations, fields, and tracked fields from a workflow definition
func deriveWorkflowDefinitionPrefilter(doc *models.WorkflowDefinitionDocument) ([]string, []string, []string) {
	if doc == nil {
		return nil, nil, nil
	}

	ops, fields := workflows.DeriveTriggerPrefilter(*doc)
	tracked := append([]string(nil), fields...)

	return ops, fields, tracked
}

// deriveWorkflowDefinitionApprovalFields extracts approval fields and edges from approval actions in the definition
func deriveWorkflowDefinitionApprovalFields(doc *models.WorkflowDefinitionDocument) ([]string, []string) {
	fieldSet := map[string]struct{}{}
	edgeSet := map[string]struct{}{}

	for _, action := range doc.Actions {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType == nil || *actionType != enums.WorkflowActionTypeApproval {
			continue
		}

		var params struct {
			Fields []string `json:"fields"`
			Edges  []string `json:"edges"`
		}

		if len(action.Params) == 0 {
			continue
		}

		if err := json.Unmarshal(action.Params, &params); err != nil {
			continue
		}

		for _, field := range params.Fields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			fieldSet[field] = struct{}{}
		}

		for _, edge := range params.Edges {
			edge = strings.TrimSpace(edge)
			if edge == "" {
				continue
			}

			edgeSet[edge] = struct{}{}
		}
	}

	fields := make([]string, 0, len(fieldSet))
	for field := range fieldSet {
		fields = append(fields, field)
	}

	sort.Strings(fields)

	edges := make([]string, 0, len(edgeSet))
	for edge := range edgeSet {
		edges = append(edges, edge)
	}

	sort.Strings(edges)

	return fields, edges
}

// deriveWorkflowDefinitionApprovalSubmissionMode extracts the approval submission mode from a workflow definition
func deriveWorkflowDefinitionApprovalSubmissionMode(doc *models.WorkflowDefinitionDocument) enums.WorkflowApprovalSubmissionMode {
	if doc == nil {
		return enums.WorkflowApprovalSubmissionModeManualSubmit
	}

	mode := doc.ApprovalSubmissionMode
	if mode == "" {
		return enums.WorkflowApprovalSubmissionModeManualSubmit
	}

	if parsed := enums.ToWorkflowApprovalSubmissionMode(mode.String()); parsed != nil {
		return *parsed
	}

	return enums.WorkflowApprovalSubmissionModeManualSubmit
}
