package graphapi

import (
	"context"
	"encoding/json"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignment"
	"github.com/theopenlane/core/internal/ent/generated/workflowassignmenttarget"
	"github.com/theopenlane/core/internal/ent/generated/workflowinstance"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/jsonx"
)

// diffContextLines is the number of context lines to show around changes in unified diffs
const diffContextLines = 3

func (r *Resolver) workflowInstanceProposalPreview(ctx context.Context, instance *generated.WorkflowInstance) (*model.WorkflowProposalPreview, error) {
	if instance == nil || instance.WorkflowProposalID == "" {
		return nil, nil
	}

	objectType, objectID, err := workflowInstanceObjectContext(ctx, r.db, instance)
	if err != nil {
		return nil, err
	}
	if objectID == "" {
		return nil, nil
	}

	instanceCaller, ok := auth.CallerFromContext(ctx)
	if !ok || instanceCaller == nil || instanceCaller.SubjectID == "" {
		return nil, rout.ErrPermissionDenied
	}

	allow, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		ObjectType:  fgax.Kind(strcase.SnakeCase(objectType.String())),
		ObjectID:    objectID,
		Relation:    fgax.CanEdit,
		SubjectID:   instanceCaller.SubjectID,
		SubjectType: instanceCaller.SubjectType(),
	})
	if err != nil {
		return nil, err
	}
	if !allow {
		isApprover, err := workflowInstanceHasApprover(ctx, r.db, instance.ID, instanceCaller.SubjectID)
		if err != nil {
			return nil, err
		}
		if !isApprover {
			return nil, rout.ErrPermissionDenied
		}
	}

	allowCtx := workflows.AllowContext(ctx)

	proposal, err := r.db.WorkflowProposal.Get(allowCtx, instance.WorkflowProposalID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowproposal"})
	}

	return buildWorkflowProposalPreview(ctx, r.db, proposal, objectType, objectID)
}

// workflowProposalPreview is a resolver helper for creating previews of the proposal information
func (r *Resolver) workflowProposalPreview(ctx context.Context, proposal *generated.WorkflowProposal) (*model.WorkflowProposalPreview, error) {
	if proposal == nil || proposal.ID == "" {
		return nil, nil
	}

	objectType, objectID, err := workflowProposalObjectContext(ctx, r.db, proposal)
	if err != nil {
		return nil, err
	}
	if objectID == "" {
		return nil, nil
	}

	proposalCaller, ok := auth.CallerFromContext(ctx)
	if !ok || proposalCaller == nil || proposalCaller.SubjectID == "" {
		return nil, rout.ErrPermissionDenied
	}

	allow, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		ObjectType:  fgax.Kind(strcase.SnakeCase(objectType.String())),
		ObjectID:    objectID,
		Relation:    fgax.CanEdit,
		SubjectID:   proposalCaller.SubjectID,
		SubjectType: proposalCaller.SubjectType(),
	})
	if err != nil {
		return nil, err
	}
	if !allow {
		isApprover, err := workflowProposalHasApprover(ctx, r.db, proposal.ID, proposalCaller.SubjectID)
		if err != nil {
			return nil, err
		}
		if !isApprover {
			return nil, rout.ErrPermissionDenied
		}
	}

	return buildWorkflowProposalPreview(ctx, r.db, proposal, objectType, objectID)
}

func buildWorkflowProposalPreview(ctx context.Context, client *generated.Client, proposal *generated.WorkflowProposal, objectType enums.WorkflowObjectType, objectID string) (*model.WorkflowProposalPreview, error) {
	if proposal == nil {
		return nil, nil
	}

	fields := workflows.FieldsFromChanges(proposal.Changes)
	fieldMeta := workflowProposalFieldMetadata(objectType)

	currentValues := map[string]any{}
	if len(fields) > 0 {
		allowCtx := workflows.AllowContext(ctx)
		entity, err := workflows.LoadWorkflowObject(allowCtx, client, objectType.String(), objectID)
		if err != nil {
			return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowobject"})
		}

		currentValues, err = workflowProposalCurrentValues(entity, fields)
		if err != nil {
			return nil, err
		}
	}

	diffs := make([]*model.WorkflowFieldDiff, 0, len(fields))
	for _, field := range fields {
		meta := fieldMeta[field]
		currentValue := currentValues[field]
		proposedValue := proposal.Changes[field]
		diffText := workflowProposalDiff(currentValue, proposedValue)

		diffs = append(diffs, &model.WorkflowFieldDiff{
			Field:         field,
			Label:         workflowProposalOptionalString(meta.Label),
			Type:          workflowProposalOptionalString(meta.Type),
			CurrentValue:  currentValue,
			ProposedValue: proposedValue,
			Diff:          workflowProposalOptionalString(diffText),
		})
	}

	preview := &model.WorkflowProposalPreview{
		ProposalID:      proposal.ID,
		DomainKey:       proposal.DomainKey,
		State:           proposal.State,
		ProposedChanges: proposal.Changes,
		CurrentValues:   currentValues,
		Diffs:           diffs,
	}

	if proposal.SubmittedAt != nil {
		dt := models.DateTime(*proposal.SubmittedAt)
		preview.SubmittedAt = &dt
	}

	if proposal.SubmittedByUserID != "" {
		preview.SubmittedByUserID = &proposal.SubmittedByUserID
	}

	return preview, nil
}

func workflowProposalOptionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func workflowInstanceObjectContext(ctx context.Context, client *generated.Client, instance *generated.WorkflowInstance) (enums.WorkflowObjectType, string, error) {
	if instance == nil {
		return "", "", nil
	}

	if instance.Context.ObjectID != "" && instance.Context.ObjectType != "" {
		return instance.Context.ObjectType, instance.Context.ObjectID, nil
	}

	if obj, ok := workflowObjectFromRefs(instance.Edges.WorkflowObjectRefs); ok {
		return obj.Type, obj.ID, nil
	}

	if client == nil {
		return "", "", nil
	}

	allowCtx := workflows.AllowContext(ctx)
	query := client.WorkflowObjectRef.Query().
		Where(workflowobjectref.WorkflowInstanceIDEQ(instance.ID))
	if instance.OwnerID != "" {
		query = query.Where(workflowobjectref.OwnerIDEQ(instance.OwnerID))
	}

	ref, err := query.First(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return "", "", nil
		}
		return "", "", parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowobjectref"})
	}

	if obj, ok := workflowObjectFromRefs([]*generated.WorkflowObjectRef{ref}); ok {
		return obj.Type, obj.ID, nil
	}

	return "", "", nil
}

func workflowProposalObjectContext(ctx context.Context, client *generated.Client, proposal *generated.WorkflowProposal) (enums.WorkflowObjectType, string, error) {
	if proposal == nil {
		return "", "", nil
	}

	if proposal.Edges.WorkflowObjectRef != nil {
		obj, err := workflows.ObjectFromRef(proposal.Edges.WorkflowObjectRef)
		if err == nil && obj != nil {
			return obj.Type, obj.ID, nil
		}
	}

	if proposal.WorkflowObjectRefID == "" || client == nil {
		return "", "", nil
	}

	allowCtx := workflows.AllowContext(ctx)
	ref, err := client.WorkflowObjectRef.Get(allowCtx, proposal.WorkflowObjectRefID)
	if err != nil {
		if generated.IsNotFound(err) {
			return "", "", nil
		}
		return "", "", parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowobjectref"})
	}

	obj, err := workflows.ObjectFromRef(ref)
	if err == nil && obj != nil {
		return obj.Type, obj.ID, nil
	}

	return "", "", nil
}

func workflowObjectFromRefs(refs []*generated.WorkflowObjectRef) (*workflows.Object, bool) {
	for _, ref := range refs {
		if ref == nil {
			continue
		}
		obj, err := workflows.ObjectFromRef(ref)
		if err == nil && obj != nil {
			return obj, true
		}
	}

	return nil, false
}

func workflowInstanceHasApprover(ctx context.Context, client *generated.Client, instanceID string, userID string) (bool, error) {
	if client == nil || instanceID == "" || userID == "" {
		return false, nil
	}

	allowCtx := workflows.AllowContext(ctx)

	direct, err := client.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.TargetUserIDEQ(userID),
			workflowassignmenttarget.HasWorkflowAssignmentWith(
				workflowassignment.WorkflowInstanceIDEQ(instanceID),
			),
		).
		Exist(allowCtx)
	if err != nil {
		return false, err
	}
	if direct {
		return true, nil
	}

	groupIDs, err := client.GroupMembership.Query().
		Where(groupmembership.UserIDEQ(userID)).
		Select(groupmembership.FieldGroupID).
		Strings(allowCtx)
	if err != nil {
		return false, err
	}
	if len(groupIDs) == 0 {
		return false, nil
	}

	groupTarget, err := client.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.TargetGroupIDIn(groupIDs...),
			workflowassignmenttarget.HasWorkflowAssignmentWith(
				workflowassignment.WorkflowInstanceIDEQ(instanceID),
			),
		).
		Exist(allowCtx)
	if err != nil {
		return false, err
	}

	return groupTarget, nil
}

func workflowProposalHasApprover(ctx context.Context, client *generated.Client, proposalID string, userID string) (bool, error) {
	if client == nil || proposalID == "" || userID == "" {
		return false, nil
	}

	allowCtx := workflows.AllowContext(ctx)

	direct, err := client.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.TargetUserIDEQ(userID),
			workflowassignmenttarget.HasWorkflowAssignmentWith(
				workflowassignment.HasWorkflowInstanceWith(
					workflowinstance.WorkflowProposalIDEQ(proposalID),
				),
			),
		).
		Exist(allowCtx)
	if err != nil {
		return false, err
	}
	if direct {
		return true, nil
	}

	groupIDs, err := client.GroupMembership.Query().
		Where(groupmembership.UserIDEQ(userID)).
		Select(groupmembership.FieldGroupID).
		Strings(allowCtx)
	if err != nil {
		return false, err
	}
	if len(groupIDs) == 0 {
		return false, nil
	}

	groupTarget, err := client.WorkflowAssignmentTarget.Query().
		Where(
			workflowassignmenttarget.TargetGroupIDIn(groupIDs...),
			workflowassignmenttarget.HasWorkflowAssignmentWith(
				workflowassignment.HasWorkflowInstanceWith(
					workflowinstance.WorkflowProposalIDEQ(proposalID),
				),
			),
		).
		Exist(allowCtx)
	if err != nil {
		return false, err
	}

	return groupTarget, nil
}

func workflowProposalFieldMetadata(objectType enums.WorkflowObjectType) map[string]generated.WorkflowFieldInfo {
	fields := map[string]generated.WorkflowFieldInfo{}
	for _, meta := range workflows.WorkflowMetadata() {
		if meta.Type != objectType {
			continue
		}

		for _, field := range meta.EligibleFields {
			fields[field.Name] = field
		}

		break
	}

	return fields
}

func workflowProposalCurrentValues(entity any, fields []string) (map[string]any, error) {
	values := make(map[string]any, len(fields))
	if entity == nil {
		return values, nil
	}

	objectMap, err := jsonx.ToMap(entity)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		if value, ok := objectMap[field]; ok {
			values[field] = value
			continue
		}

		values[field] = nil
	}

	return values, nil
}

func workflowProposalDiff(currentValue any, proposedValue any) string {
	currentStr, currentOK := workflowProposalDiffString(currentValue)
	proposedStr, proposedOK := workflowProposalDiffString(proposedValue)
	if !currentOK && !proposedOK {
		return ""
	}

	if !currentOK {
		currentStr = ""
	}

	if !proposedOK {
		proposedStr = ""
	}

	if currentStr == proposedStr {
		return ""
	}

	ud := difflib.UnifiedDiff{
		A:        difflib.SplitLines(currentStr),
		B:        difflib.SplitLines(proposedStr),
		FromFile: "current",
		ToFile:   "proposed",
		Context:  diffContextLines,
	}

	diff, _ := difflib.GetUnifiedDiffString(ud)

	return diff
}

func workflowProposalDiffString(value any) (string, bool) {
	if value == nil {
		return "", true
	}

	switch v := value.(type) {
	case string:
		return v, true
	case []byte:
		return string(v), true
	case json.RawMessage:
		return string(v), true
	default:
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", false
		}
		return string(data), true
	}
}
