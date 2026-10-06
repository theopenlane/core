package workflows

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/workflowobjectref"
	"github.com/theopenlane/iam/auth"
)

// ObjectRefIDs returns workflow object ref IDs matching the workflow object.
func ObjectRefIDs(ctx context.Context, client *generated.Client, obj *Object) ([]string, error) {
	if client == nil {
		return nil, ErrNilClient
	}
	if obj == nil || obj.ID == "" {
		return nil, ErrMissingObjectID
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, auth.ErrNoAuthUser
	}

	query := buildObjectRefQuery(client.WorkflowObjectRef.Query().Where(workflowobjectref.OwnerIDEQ(orgID)), obj)
	if query == nil {
		return nil, ErrUnsupportedObjectType
	}

	return query.IDs(ctx)
}

// buildObjectRefQuery applies the canonical schema predicate for the object type.
func buildObjectRefQuery(query *generated.WorkflowObjectRefQuery, obj *Object) *generated.WorkflowObjectRefQuery {
	if obj == nil {
		return nil
	}

	filtered, err := FilterWorkflowObjectRefs(query, obj.Type, obj.ID)
	if err != nil {
		return nil
	}

	return filtered
}
