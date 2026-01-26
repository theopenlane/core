package workflows

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
)

// LoadWorkflowObject loads an ent object that participates in workflows.
func LoadWorkflowObject(ctx context.Context, client *generated.Client, schemaType string, objectID string) (any, error) {
	return client.LoadWorkflowObject(ctx, schemaType, objectID)
}

// ObjectOwnerID resolves the owner ID for a workflow object via generated helpers.
func ObjectOwnerID(ctx context.Context, client *generated.Client, objectType enums.WorkflowObjectType, objectID string) (string, error) {
	return generated.GetObjectOwnerID(ctx, client, objectType, objectID)
}

// ApplyObjectFieldUpdates applies updates to a workflow object via generated helpers.
func ApplyObjectFieldUpdates(ctx context.Context, client *generated.Client, objectType enums.WorkflowObjectType, objectID string, updates map[string]any) error {
	return generated.ApplyObjectFieldUpdates(ctx, client, objectType, objectID, updates)
}

// WorkflowMetadata returns generated workflow-eligible schema metadata.
func WorkflowMetadata() []generated.WorkflowObjectTypeInfo {
	return generated.GetWorkflowMetadata()
}

// OrganizationOwnerIDs returns user IDs for owners of an organization.
func OrganizationOwnerIDs(ctx context.Context, client *generated.Client, orgID string) ([]string, error) {
	if client == nil {
		return nil, ErrNilClient
	}
	if orgID == "" {
		return nil, ErrMissingOrganizationID
	}

	memberships, err := client.OrgMembership.
		Query().
		Where(
			orgmembership.OrganizationIDEQ(orgID),
			orgmembership.RoleEQ(enums.RoleOwner),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(memberships))
	for _, membership := range memberships {
		if membership.UserID != "" {
			userIDs = append(userIDs, membership.UserID)
		}
	}

	return userIDs, nil
}
