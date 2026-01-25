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
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/internal/workflows"
)

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

	userID, err := auth.GetSubjectIDFromContext(ctx)
	if err != nil || userID == "" {
		return nil, rout.ErrPermissionDenied
	}

	allow, err := r.db.Authz.CheckAccess(ctx, fgax.AccessCheck{
		ObjectType:  fgax.Kind(strcase.SnakeCase(objectType.String())),
		ObjectID:    objectID,
		Relation:    fgax.CanEdit,
		SubjectID:   userID,
		SubjectType: auth.GetAuthzSubjectType(ctx),
	})
	if err != nil {
		return nil, err
	}
	if !allow {
		return nil, rout.ErrPermissionDenied
	}

	allowCtx := workflows.AllowContext(ctx)

	proposal, err := r.db.WorkflowProposal.Get(allowCtx, instance.WorkflowProposalID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionGet, Object: "workflowproposal"})
	}

	fields := workflows.FieldsFromChanges(proposal.Changes)
	fieldMeta := workflowProposalFieldMetadata(objectType)

	currentValues := map[string]any{}
	if len(fields) > 0 {
		entity, err := workflows.LoadWorkflowObject(allowCtx, r.db, objectType.String(), objectID)
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

	data, err := json.Marshal(entity)
	if err != nil {
		return nil, err
	}

	var objectMap map[string]any
	if err := json.Unmarshal(data, &objectMap); err != nil {
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
		Context:  3,
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
