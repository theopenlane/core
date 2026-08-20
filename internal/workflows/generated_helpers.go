package workflows

import (
	"context"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/orgmembership"
)

// WorkflowFieldInfo is the workflow-facing view of a canonical entity field.
type WorkflowFieldInfo struct {
	Name  string
	Label string
	Type  string
}

// WorkflowObjectTypeInfo is the workflow-facing view of a canonical entity schema.
type WorkflowObjectTypeInfo struct {
	Type           enums.WorkflowObjectType
	Label          string
	Description    string
	EligibleFields []WorkflowFieldInfo
	EligibleEdges  []string
}

// WorkflowSchema resolves a workflow object type through the canonical entity catalog.
func WorkflowSchema(objectType enums.WorkflowObjectType) (*entityops.Schema, error) {
	schema, ok := entityops.LookupSchema(objectType.String())
	if !ok || !schema.WorkflowEligible {
		return nil, ErrUnsupportedObjectType
	}

	return schema, nil
}

// workflowSchemaName resolves a generated schema name through the canonical entity catalog.
func workflowSchemaName(schemaType string) (*entityops.Schema, error) {
	schema, ok := entityops.LookupSchema(schemaType)
	if !ok || !schema.WorkflowEligible {
		return nil, ErrUnsupportedObjectType
	}

	return schema, nil
}

// LoadWorkflowObject loads an ent object that participates in workflows.
func LoadWorkflowObject(ctx context.Context, client *generated.Client, schemaType string, objectID string) (generated.Noder, error) {
	schema, err := workflowSchemaName(schemaType)
	if err != nil {
		return nil, err
	}

	return schema.LoadWorkflowObject(ctx, client, objectID)
}

// ObjectOwnerID resolves the owner ID for a workflow object via the canonical schema.
func ObjectOwnerID(ctx context.Context, client *generated.Client, objectType enums.WorkflowObjectType, objectID string) (string, error) {
	schema, err := WorkflowSchema(objectType)
	if err != nil {
		return "", err
	}

	return schema.WorkflowOwnerID(ctx, client, objectID)
}

// ApplyObjectFieldUpdates validates and applies updates through the canonical schema.
func ApplyObjectFieldUpdates(ctx context.Context, client *generated.Client, objectType enums.WorkflowObjectType, objectID string, updates map[string]any) error {
	schema, err := WorkflowSchema(objectType)
	if err != nil {
		return err
	}

	return schema.ApplyWorkflowFields(ctx, client, objectID, updates)
}

// EnrichWorkflowPayload adds schema-designated fields to a workflow webhook payload.
func EnrichWorkflowPayload(ctx context.Context, client *generated.Client, objectType enums.WorkflowObjectType, objectID string, payload map[string]any) error {
	schema, err := WorkflowSchema(objectType)
	if err != nil {
		return err
	}

	return schema.EnrichWorkflowPayload(ctx, client, objectID, payload)
}

// FilterWorkflowInstances restricts a workflow-instance query to one canonical object.
func FilterWorkflowInstances(query *generated.WorkflowInstanceQuery, objectType enums.WorkflowObjectType, objectID string) (*generated.WorkflowInstanceQuery, error) {
	schema, err := WorkflowSchema(objectType)
	if err != nil {
		return nil, err
	}

	return schema.FilterWorkflowInstances(query, objectID)
}

// FilterWorkflowObjectRefs restricts a workflow-object-ref query to one canonical object.
func FilterWorkflowObjectRefs(query *generated.WorkflowObjectRefQuery, objectType enums.WorkflowObjectType, objectID string) (*generated.WorkflowObjectRefQuery, error) {
	schema, err := WorkflowSchema(objectType)
	if err != nil {
		return nil, err
	}

	return schema.FilterWorkflowObjectRefs(query, objectID)
}

// WorkflowMetadata derives workflow object metadata from canonical schema descriptors.
func WorkflowMetadata() []WorkflowObjectTypeInfo {
	metadata := make([]WorkflowObjectTypeInfo, 0)
	for _, schema := range entityops.AllSchemas() {
		if schema == nil || !schema.WorkflowEligible {
			continue
		}

		objectType := enums.ToWorkflowObjectType(schema.Name)
		if objectType == nil {
			continue
		}

		fields := schema.WorkflowFields()
		fieldInfo := make([]WorkflowFieldInfo, 0, len(fields))
		for _, field := range fields {
			fieldInfo = append(fieldInfo, WorkflowFieldInfo{
				Name:  field.Name,
				Label: field.Label,
				Type:  field.Type,
			})
		}

		edges := schema.WorkflowEdges()
		edgeNames := make([]string, 0, len(edges))
		for _, edge := range edges {
			edgeNames = append(edgeNames, edge.Name)
		}

		metadata = append(metadata, WorkflowObjectTypeInfo{
			Type:           *objectType,
			Label:          schema.Name,
			Description:    schema.Name + " objects",
			EligibleFields: fieldInfo,
			EligibleEdges:  edgeNames,
		})
	}

	return metadata
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
