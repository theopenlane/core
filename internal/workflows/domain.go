package workflows

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/mutations"
)

// DeriveDomainKey generates a stable domain key from a sorted list of field names.
// A "domain key" is a canonicalized list of approval fields scoped to the object type.
// Format: "ObjectType:field1,field2" (fields are sorted); if fields are empty, returns the object type string.
func DeriveDomainKey(objectType enums.WorkflowObjectType, fields []string) string {
	sorted := make([]string, len(fields))
	copy(sorted, fields)
	sort.Strings(sorted)

	prefix := objectType.String()
	if prefix == "" {
		return strings.Join(sorted, ",")
	}

	if len(sorted) == 0 {
		return prefix
	}

	return prefix + ":" + strings.Join(sorted, ",")
}

// DomainChanges represents changes grouped by approval domain
type DomainChanges struct {
	// DomainKey is the derived key for the approval field set
	DomainKey string
	// Fields lists the approval fields included in the domain
	Fields []string
	// Changes is the subset of proposed changes for the domain
	Changes map[string]any
}

// FilterChangesForDomain filters a mutation's changes to only include fields that belong to the specified approval domain
func FilterChangesForDomain(changes map[string]any, approvalFields []string) map[string]any {
	if len(changes) == 0 {
		return nil
	}

	domainChanges := make(map[string]any)
	for _, field := range approvalFields {
		if value, exists := changes[field]; exists {
			domainChanges[field] = value
		}
	}

	return domainChanges
}

// ComputeProposalHash computes a SHA256 hash of a changes map with sorted keys
func ComputeProposalHash(changes map[string]any) (string, error) {
	keys := lo.Keys(changes)

	sort.Strings(keys)

	sorted := make(map[string]any, len(keys))
	for _, k := range keys {
		sorted[k] = changes[k]
	}

	data, err := json.Marshal(sorted)
	if err != nil {
		return "", fmt.Errorf("failed to marshal changes for hashing: %w", err)
	}

	hash := sha256.Sum256(data)

	return fmt.Sprintf("%x", hash), nil
}

// DefinitionHasApprovalAction returns true if the workflow definition contains at least one approval action
func DefinitionHasApprovalAction(doc models.WorkflowDefinitionDocument) bool {
	return lo.SomeBy(doc.Actions, func(action models.WorkflowAction) bool {
		actionType := enums.ToWorkflowActionType(action.Type)
		return actionType != nil && *actionType == enums.WorkflowActionTypeApproval
	})
}

// DefinitionHasReviewAction returns true if the workflow definition contains at least one review action.
func DefinitionHasReviewAction(doc models.WorkflowDefinitionDocument) bool {
	return lo.SomeBy(doc.Actions, func(action models.WorkflowAction) bool {
		actionType := enums.ToWorkflowActionType(action.Type)
		return actionType != nil && *actionType == enums.WorkflowActionTypeReview
	})
}

// ApprovalTimingOrDefault resolves the approval timing, defaulting to PRE_COMMIT.
func ApprovalTimingOrDefault(doc models.WorkflowDefinitionDocument) enums.WorkflowApprovalTiming {
	if doc.ApprovalTiming == "" {
		return enums.WorkflowApprovalTimingPreCommit
	}
	if parsed := enums.ToWorkflowApprovalTiming(doc.ApprovalTiming.String()); parsed != nil {
		return *parsed
	}

	return enums.WorkflowApprovalTimingPreCommit
}

// DefinitionUsesPreCommitApprovals returns true if approvals should block changes.
func DefinitionUsesPreCommitApprovals(doc models.WorkflowDefinitionDocument) bool {
	return DefinitionHasApprovalAction(doc) && ApprovalTimingOrDefault(doc) == enums.WorkflowApprovalTimingPreCommit
}

// DefinitionUsesPostCommitApprovals returns true if approvals should run after commit.
func DefinitionUsesPostCommitApprovals(doc models.WorkflowDefinitionDocument) bool {
	return DefinitionHasApprovalAction(doc) && ApprovalTimingOrDefault(doc) == enums.WorkflowApprovalTimingPostCommit
}

// ConvertApprovalActionsToReview returns a copy of the definition with approval actions
// converted to review actions. This is used to reuse review handling for POST_COMMIT flows.
func ConvertApprovalActionsToReview(doc models.WorkflowDefinitionDocument) models.WorkflowDefinitionDocument {
	if len(doc.Actions) == 0 {
		return doc
	}

	actions := make([]models.WorkflowAction, len(doc.Actions))
	copy(actions, doc.Actions)

	updated := false
	for i, action := range actions {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType == nil || *actionType != enums.WorkflowActionTypeApproval {
			continue
		}

		actions[i].Type = string(enums.WorkflowActionTypeReview)
		updated = true
	}

	if !updated {
		return doc
	}

	out := doc
	out.Actions = actions
	return out
}

// ApprovalDomains returns the distinct field sets used by approval actions in a definition
// It returns an error when approval action params cannot be parsed
func ApprovalDomains(doc models.WorkflowDefinitionDocument) ([][]string, error) {
	seen := map[string]struct{}{}
	domains := make([][]string, 0)

	for _, action := range doc.Actions {
		actionType := enums.ToWorkflowActionType(action.Type)
		if actionType == nil || *actionType != enums.WorkflowActionTypeApproval {
			continue
		}

		var params ApprovalActionParams
		if len(action.Params) == 0 {
			continue
		}
		if err := json.Unmarshal(action.Params, &params); err != nil {
			return nil, fmt.Errorf("%w: action %q: %v", ErrApprovalActionParamsInvalid, action.Key, err)
		}

		fields := mutations.NormalizeStrings(params.Fields)
		if len(fields) == 0 {
			continue
		}

		sort.Strings(fields)
		key := DeriveDomainKey("", fields)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		domains = append(domains, fields)
	}

	return domains, nil
}

// DomainChangesForDefinition splits proposed changes by approval domains for a workflow definition
func DomainChangesForDefinition(doc models.WorkflowDefinitionDocument, objectType enums.WorkflowObjectType, proposedChanges map[string]any) ([]DomainChanges, error) {
	domains, err := ApprovalDomains(doc)
	if err != nil {
		return nil, err
	}

	domainChanges := SplitChangesByDomains(proposedChanges, objectType, domains)
	if len(domainChanges) == 0 {
		fields := lo.Keys(proposedChanges)
		if len(fields) > 0 {
			sort.Strings(fields)
			domainChanges = []DomainChanges{{
				DomainKey: DeriveDomainKey(objectType, fields),
				Fields:    fields,
				Changes:   proposedChanges,
			}}
		}
	}

	return domainChanges, nil
}

// SplitChangesByDomains filters changes into per-domain buckets based on approval field sets
func SplitChangesByDomains(changes map[string]any, objectType enums.WorkflowObjectType, domains [][]string) []DomainChanges {
	out := make([]DomainChanges, 0, len(domains))
	for _, domainFields := range domains {
		domainChanges := FilterChangesForDomain(changes, domainFields)
		if len(domainChanges) == 0 {
			continue
		}

		fields := make([]string, len(domainFields))
		copy(fields, domainFields)
		out = append(out, DomainChanges{
			DomainKey: DeriveDomainKey(objectType, fields),
			Fields:    fields,
			Changes:   domainChanges,
		})
	}

	return out
}

// FieldsFromChanges returns the sorted field names present in a changes map
func FieldsFromChanges(changes map[string]any) []string {
	fields := lo.Keys(changes)

	sort.Strings(fields)

	return fields
}
