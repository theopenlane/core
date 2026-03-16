package workflows

import (
	"reflect"

	"github.com/theopenlane/core/common/models"
)

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
	// Type is the reflected Go type represented by this schema definition
	Type reflect.Type
}

// WorkflowDefinitionSchemaRootType is the root Go type for workflow definition schema generation
var WorkflowDefinitionSchemaRootType = WorkflowSchemaTypeDefinition{
	Name: "WorkflowDefinitionDocument", Type: reflect.TypeFor[models.WorkflowDefinitionDocument](),
}

// WorkflowDefinitionSchemaModelTypes are model definition types expected in workflow schema $defs
var WorkflowDefinitionSchemaModelTypes = []WorkflowSchemaTypeDefinition{
	{Name: "WorkflowTrigger", Type: reflect.TypeFor[models.WorkflowTrigger]()},
	{Name: "WorkflowCondition", Type: reflect.TypeFor[models.WorkflowCondition]()},
	{Name: "WorkflowAction", Type: reflect.TypeFor[models.WorkflowAction]()},
	{Name: "WorkflowSelector", Type: reflect.TypeFor[models.WorkflowSelector]()},
}

// WorkflowDefinitionSchemaExtensionTypes are additional workflow schema definitions beyond the root model graph
var WorkflowDefinitionSchemaExtensionTypes = []WorkflowSchemaTypeDefinition{
	{Name: "TargetConfig", Type: reflect.TypeFor[TargetConfig]()},
	{Name: "ApprovalActionParams", Type: reflect.TypeFor[ApprovalActionParams]()},
	{Name: "ReviewActionParams", Type: reflect.TypeFor[ReviewActionParams]()},
	{Name: "NotificationActionParams", Type: reflect.TypeFor[NotificationActionParams]()},
	{Name: "WebhookActionParams", Type: reflect.TypeFor[WebhookActionParams]()},
	{Name: "FieldUpdateActionParams", Type: reflect.TypeFor[FieldUpdateActionParams]()},
	{Name: "IntegrationActionParams", Type: reflect.TypeFor[IntegrationActionParams]()},
	{Name: "CreateObjectActionParams", Type: reflect.TypeFor[CreateObjectActionParams]()},
}
