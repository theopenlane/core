package workflows

import "github.com/theopenlane/core/common/models"

const (
	// WorkflowDefinitionJSONSchemaVersion is the JSON Schema draft used for workflow definition schemas
	WorkflowDefinitionJSONSchemaVersion = "https://json-schema.org/draft/2020-12/schema"
	// WorkflowDefinitionJSONSchemaID is the canonical schema identifier for workflow definitions
	WorkflowDefinitionJSONSchemaID = "https://theopenlane.io/schemas/workflow-definition.json"
	// WorkflowDefinitionJSONSchemaTitle is the display title for workflow definition schemas
	WorkflowDefinitionJSONSchemaTitle = "Workflow Definition"
	// WorkflowDefinitionJSONSchemaDescription is the top-level description for workflow definition schemas
	WorkflowDefinitionJSONSchemaDescription = "Schema for Openlane workflow definitions"
)

// WorkflowSchemaTypeDefinition identifies a named Go type used in workflow schema generation
type WorkflowSchemaTypeDefinition struct {
	// Name is the schema definition name under $defs
	Name string
	// Value is a zero value of the Go type represented by this schema definition
	Value any
}

// WorkflowDefinitionSchemaRootType is the root Go type for workflow definition schema generation
var WorkflowDefinitionSchemaRootType = WorkflowSchemaTypeDefinition{
	Name:  "WorkflowDefinitionDocument",
	Value: models.WorkflowDefinitionDocument{},
}

// WorkflowDefinitionSchemaModelTypes are model definition types expected in workflow schema $defs
var WorkflowDefinitionSchemaModelTypes = []WorkflowSchemaTypeDefinition{
	{Name: "WorkflowTrigger", Value: models.WorkflowTrigger{}},
	{Name: "WorkflowCondition", Value: models.WorkflowCondition{}},
	{Name: "WorkflowAction", Value: models.WorkflowAction{}},
	{Name: "WorkflowSelector", Value: models.WorkflowSelector{}},
}

// WorkflowDefinitionSchemaExtensionTypes are additional workflow schema definitions beyond the root model graph
var WorkflowDefinitionSchemaExtensionTypes = []WorkflowSchemaTypeDefinition{
	{Name: "TargetConfig", Value: TargetConfig{}},
	{Name: "ApprovalActionParams", Value: ApprovalActionParams{}},
	{Name: "ReviewActionParams", Value: ReviewActionParams{}},
	{Name: "NotificationActionParams", Value: NotificationActionParams{}},
	{Name: "WebhookActionParams", Value: WebhookActionParams{}},
	{Name: "FieldUpdateActionParams", Value: FieldUpdateActionParams{}},
	{Name: "IntegrationActionParams", Value: IntegrationActionParams{}},
	{Name: "CreateObjectActionParams", Value: CreateObjectActionParams{}},
}
