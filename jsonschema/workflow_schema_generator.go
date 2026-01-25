//go:build generate

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/invopop/jsonschema"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/workflows"
)

const (
	workflowSchemaPath = "./jsonschema/workflow.definition.json"
	ownerReadWritePerm = 0600
)

// workflowSchemaTypes is the list of types to expose in the workflow schema
var workflowSchemaTypes = []any{
	models.WorkflowDefinitionDocument{},
	models.WorkflowTrigger{},
	models.WorkflowCondition{},
	models.WorkflowAction{},
	models.WorkflowSelector{},
	workflows.TargetConfig{},
	workflows.ApprovalActionParams{},
	workflows.NotificationActionParams{},
	workflows.WebhookActionParams{},
	workflows.FieldUpdateActionParams{},
	workflows.IntegrationActionParams{},
}

// main generates the workflow definition JSON schema
func main() {
	if err := generateWorkflowSchema(workflowSchemaPath); err != nil {
		panic(err)
	}
}

// generateWorkflowSchema creates the JSON schema file for workflow definitions
func generateWorkflowSchema(outputPath string) error {
	r := &jsonschema.Reflector{
		ExpandedStruct:             true,
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             false,
	}

	r.Namer = func(t reflect.Type) string {
		return t.Name()
	}

	// Add enum mappers for workflow-specific enums
	r.Mapper = workflowTypeMapper

	schema := r.Reflect(&models.WorkflowDefinitionDocument{})

	// Add $schema and $id
	schema.Version = "https://json-schema.org/draft/2020-12/schema"
	schema.ID = "https://theopenlane.io/schemas/workflow-definition.json"
	schema.Title = "Workflow Definition"
	schema.Description = "Schema for Openlane workflow definitions"

	// Enhance the schema with action params definitions
	if err := addActionParamsDefinitions(schema); err != nil {
		return fmt.Errorf("failed to add action params definitions: %w", err)
	}

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow schema: %w", err)
	}

	if err := os.WriteFile(outputPath, data, ownerReadWritePerm); err != nil {
		return fmt.Errorf("failed to write workflow schema: %w", err)
	}

	return nil
}

// workflowTypeMapper handles custom type mappings for enums
func workflowTypeMapper(t reflect.Type) *jsonschema.Schema {
	switch t {
	case reflect.TypeOf(enums.WorkflowKind("")):
		return &jsonschema.Schema{
			Type:        "string",
			Enum:        toInterfaceSlice(enums.WorkflowKinds),
			Description: "The kind of workflow (APPROVAL, LIFECYCLE, NOTIFICATION)",
		}
	case reflect.TypeOf(enums.WorkflowObjectType("")):
		return &jsonschema.Schema{
			Type:        "string",
			Enum:        toInterfaceSlice(enums.WorkflowObjectTypes),
			Description: "The object type the workflow applies to",
		}
	case reflect.TypeOf(enums.WorkflowApprovalSubmissionMode("")):
		return &jsonschema.Schema{
			Type:        "string",
			Enum:        toInterfaceSlice(enums.WorkflowApprovalSubmissionModes),
			Description: "Controls draft vs auto-submit behavior for approval domains",
		}
	case reflect.TypeOf(enums.WorkflowTargetType("")):
		return &jsonschema.Schema{
			Type:        "string",
			Enum:        toInterfaceSlice(enums.WorkflowTargetTypes),
			Description: "The type of target (USER, GROUP, ROLE, RESOLVER)",
		}
	case reflect.TypeOf(enums.WorkflowActionType("")):
		return &jsonschema.Schema{
			Type:        "string",
			Enum:        toInterfaceSlice(enums.WorkflowActionTypes),
			Description: "The type of workflow action",
		}
	case reflect.TypeOf(enums.Channel("")):
		return &jsonschema.Schema{
			Type:        "string",
			Enum:        toInterfaceSlice(enums.Channel("").Values()),
			Description: "Notification delivery channel",
		}
	case reflect.TypeOf(json.RawMessage{}):
		// For params field, we'll replace this with oneOf in post-processing
		return &jsonschema.Schema{
			Description: "Action-specific parameters (schema varies by action type)",
		}
	}

	return nil
}

// addActionParamsDefinitions adds the action parameter type definitions to the schema
func addActionParamsDefinitions(schema *jsonschema.Schema) error {
	if schema.Definitions == nil {
		schema.Definitions = make(jsonschema.Definitions)
	}

	r := &jsonschema.Reflector{
		ExpandedStruct:             true,
		RequiredFromJSONSchemaTags: true,
		DoNotReference:             true,
	}
	r.Mapper = workflowTypeMapper

	// Add TargetConfig definition
	targetConfigSchema := r.Reflect(&workflows.TargetConfig{})
	schema.Definitions["TargetConfig"] = targetConfigSchema
	addTargetConfigDescription(targetConfigSchema)

	// Add action param definitions
	approvalSchema := r.Reflect(&workflows.ApprovalActionParams{})
	schema.Definitions["ApprovalActionParams"] = approvalSchema
	addApprovalParamsDescription(approvalSchema)

	notificationSchema := r.Reflect(&workflows.NotificationActionParams{})
	schema.Definitions["NotificationActionParams"] = notificationSchema
	addNotificationParamsDescription(notificationSchema)

	webhookSchema := r.Reflect(&workflows.WebhookActionParams{})
	schema.Definitions["WebhookActionParams"] = webhookSchema
	addWebhookParamsDescription(webhookSchema)

	fieldUpdateSchema := r.Reflect(&workflows.FieldUpdateActionParams{})
	schema.Definitions["FieldUpdateActionParams"] = fieldUpdateSchema
	addFieldUpdateParamsDescription(fieldUpdateSchema)

	integrationSchema := r.Reflect(&workflows.IntegrationActionParams{})
	schema.Definitions["IntegrationActionParams"] = integrationSchema
	addIntegrationParamsDescription(integrationSchema)

	// Add action type descriptions to the main schema
	addWorkflowActionDescription(schema)

	return nil
}

// addTargetConfigDescription adds descriptions to TargetConfig schema
func addTargetConfigDescription(schema *jsonschema.Schema) {
	schema.Title = "TargetConfig"
	schema.Description = "Defines who should receive workflow actions"

	if schema.Properties != nil {
		if prop, ok := schema.Properties.Get("type"); ok {
			prop.Description = "Target type: USER (specific user), GROUP (group members), ROLE (role holders), or RESOLVER (dynamic resolution)"
		}
		if prop, ok := schema.Properties.Get("id"); ok {
			prop.Description = "The ID of the target user, group, or role (required for USER, GROUP, ROLE types)"
		}
		if prop, ok := schema.Properties.Get("resolver_key"); ok {
			prop.Description = "The resolver key for dynamic target resolution (required for RESOLVER type)"
		}
	}
}

// addApprovalParamsDescription adds descriptions to ApprovalActionParams schema
func addApprovalParamsDescription(schema *jsonschema.Schema) {
	schema.Title = "ApprovalActionParams"
	schema.Description = "Parameters for REQUEST_APPROVAL actions that require user or group approval"

	if schema.Properties != nil {
		if prop, ok := schema.Properties.Get("targets"); ok {
			prop.Description = "List of users, groups, roles, or resolvers who can approve"
		}
		if prop, ok := schema.Properties.Get("required"); ok {
			prop.Description = "Whether this approval is required for workflow completion (defaults to true)"
		}
		if prop, ok := schema.Properties.Get("label"); ok {
			prop.Description = "Optional display label for the approval action"
		}
		if prop, ok := schema.Properties.Get("required_count"); ok {
			prop.Description = "Number of approvals needed (quorum threshold); 0 means all targets must approve"
		}
		if prop, ok := schema.Properties.Get("fields"); ok {
			prop.Description = "Fields that are gated by this approval action (used for domain derivation)"
		}
	}
}

// addNotificationParamsDescription adds descriptions to NotificationActionParams schema
func addNotificationParamsDescription(schema *jsonschema.Schema) {
	schema.Title = "NotificationActionParams"
	schema.Description = "Parameters for NOTIFY actions that send notifications to users"

	if schema.Properties != nil {
		if prop, ok := schema.Properties.Get("targets"); ok {
			prop.Description = "List of users, groups, roles, or resolvers to receive the notification"
		}
		if prop, ok := schema.Properties.Get("channels"); ok {
			prop.Description = "Notification delivery channels (IN_APP, SLACK, EMAIL)"
		}
		if prop, ok := schema.Properties.Get("topic"); ok {
			prop.Description = "Optional notification topic for categorization"
		}
		if prop, ok := schema.Properties.Get("title"); ok {
			prop.Description = "Notification title"
		}
		if prop, ok := schema.Properties.Get("body"); ok {
			prop.Description = "Notification body content"
		}
		if prop, ok := schema.Properties.Get("data"); ok {
			prop.Description = "Optional additional data to include in the notification payload"
		}
	}
}

// addWebhookParamsDescription adds descriptions to WebhookActionParams schema
func addWebhookParamsDescription(schema *jsonschema.Schema) {
	schema.Title = "WebhookActionParams"
	schema.Description = "Parameters for WEBHOOK actions that call external HTTP endpoints"

	if schema.Properties != nil {
		if prop, ok := schema.Properties.Get("url"); ok {
			prop.Description = "The webhook endpoint URL"
		}
		if prop, ok := schema.Properties.Get("method"); ok {
			prop.Description = "HTTP method (GET, POST, PUT, DELETE, etc.)"
		}
		if prop, ok := schema.Properties.Get("headers"); ok {
			prop.Description = "Additional HTTP headers to include in the request"
		}
		if prop, ok := schema.Properties.Get("payload"); ok {
			prop.Description = "Additional data to merge into the webhook payload"
		}
		if prop, ok := schema.Properties.Get("timeout_ms"); ok {
			prop.Description = "Request timeout in milliseconds"
		}
		if prop, ok := schema.Properties.Get("secret"); ok {
			prop.Description = "Secret key for signing the webhook payload (HMAC-SHA256)"
		}
		if prop, ok := schema.Properties.Get("retries"); ok {
			prop.Description = "Number of retry attempts on failure"
		}
		if prop, ok := schema.Properties.Get("idempotency_key"); ok {
			prop.Description = "Optional idempotency key header override"
		}
	}
}

// addFieldUpdateParamsDescription adds descriptions to FieldUpdateActionParams schema
func addFieldUpdateParamsDescription(schema *jsonschema.Schema) {
	schema.Title = "FieldUpdateActionParams"
	schema.Description = "Parameters for UPDATE_FIELD actions that modify object fields"

	if schema.Properties != nil {
		if prop, ok := schema.Properties.Get("updates"); ok {
			prop.Description = "Map of field names to new values to apply"
		}
	}
}

// addIntegrationParamsDescription adds descriptions to IntegrationActionParams schema
func addIntegrationParamsDescription(schema *jsonschema.Schema) {
	schema.Title = "IntegrationActionParams"
	schema.Description = "Parameters for INTEGRATION actions that interact with external systems"

	if schema.Properties != nil {
		if prop, ok := schema.Properties.Get("integration"); ok {
			prop.Description = "Integration identifier"
		}
		if prop, ok := schema.Properties.Get("provider"); ok {
			prop.Description = "Provider override for the integration"
		}
		if prop, ok := schema.Properties.Get("operation"); ok {
			prop.Description = "The integration operation to perform"
		}
		if prop, ok := schema.Properties.Get("config"); ok {
			prop.Description = "Integration-specific configuration payload"
		}
		if prop, ok := schema.Properties.Get("timeout_ms"); ok {
			prop.Description = "Operation timeout in milliseconds"
		}
		if prop, ok := schema.Properties.Get("retries"); ok {
			prop.Description = "Number of retry attempts on failure"
		}
		if prop, ok := schema.Properties.Get("force_refresh"); ok {
			prop.Description = "Request a provider-side refresh"
		}
		if prop, ok := schema.Properties.Get("client_force"); ok {
			prop.Description = "Request a client-side refresh"
		}
	}
}

// addWorkflowActionDescription enhances the WorkflowAction schema with action type descriptions
func addWorkflowActionDescription(schema *jsonschema.Schema) {
	if schema.Definitions == nil {
		return
	}

	actionSchema, ok := schema.Definitions["WorkflowAction"]
	if !ok {
		return
	}

	actionSchema.Description = "Represents an action performed by the workflow"

	if actionSchema.Properties != nil {
		if prop, ok := actionSchema.Properties.Get("key"); ok {
			prop.Description = "Unique key identifying this action within the workflow"
		}
		if prop, ok := actionSchema.Properties.Get("type"); ok {
			prop.Description = "Action type: REQUEST_APPROVAL, NOTIFY, WEBHOOK, UPDATE_FIELD, or INTEGRATION"
			prop.Enum = toInterfaceSlice(enums.WorkflowActionTypes)
		}
		if prop, ok := actionSchema.Properties.Get("params"); ok {
			prop.Description = "Action-specific parameters. Schema depends on the action type:\n" +
				"- REQUEST_APPROVAL: ApprovalActionParams\n" +
				"- NOTIFY: NotificationActionParams\n" +
				"- WEBHOOK: WebhookActionParams\n" +
				"- UPDATE_FIELD: FieldUpdateActionParams\n" +
				"- INTEGRATION: IntegrationActionParams"
		}
		if prop, ok := actionSchema.Properties.Get("when"); ok {
			prop.Description = "Optional CEL expression to conditionally execute this action"
		}
		if prop, ok := actionSchema.Properties.Get("description"); ok {
			prop.Description = "Human-readable description of what this action does"
		}
	}
}

// toInterfaceSlice converts a string slice to an interface slice for enum values
func toInterfaceSlice(strings []string) []any {
	result := make([]any, len(strings))
	for i, s := range strings {
		result[i] = s
	}
	return result
}
