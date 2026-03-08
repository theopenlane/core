package workflows

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowobjectref"
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

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return nil, auth.ErrNoAuthUser
	}

	query := buildObjectRefQuery(client.WorkflowObjectRef.Query().Where(workflowobjectref.OwnerIDEQ(caller.OrganizationID)), obj)
	if query == nil {
		return nil, ErrUnsupportedObjectType
	}

	return query.IDs(ctx)
}

// buildObjectRefQuery applies registered query builders to match the object type
func buildObjectRefQuery(query *generated.WorkflowObjectRefQuery, obj *Object) *generated.WorkflowObjectRefQuery {
	for i := len(objectRefQueryBuilders) - 1; i >= 0; i-- {
		if next, ok := objectRefQueryBuilders[i](query, obj); ok && next != nil {
			return next
		}
	}

	return nil
}
