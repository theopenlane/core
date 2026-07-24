package types //nolint:revive

import (
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/gala"
)

// IntegrationSource is the integration-specific provenance carried in the
// Attributes of a gala.OperationContext for integration-originated events. It is
// decoded on the handling side via gala.DecodeAttributes
type IntegrationSource struct {
	// IntegrationID identifies the target integration installation
	IntegrationID string `json:"integrationId,omitempty" jsonschema:"description=Target integration installation identifier"`
	// DefinitionID identifies the integration definition
	DefinitionID string `json:"definitionId,omitempty" jsonschema:"description=Integration definition identifier"`
	// RunID identifies the persisted integration run when present
	RunID string `json:"runId,omitempty" jsonschema:"description=Persisted integration run identifier"`
	// RunType identifies the integration run classification
	RunType enums.IntegrationRunType `json:"runType,omitempty" jsonschema:"description=Integration run classification"`
	// Webhook identifies the triggering webhook when present
	Webhook string `json:"webhook,omitempty" jsonschema:"description=Triggering webhook identifier"`
	// Event identifies the triggering webhook event when present
	Event string `json:"event,omitempty" jsonschema:"description=Triggering webhook event name"`
	// DeliveryID identifies the provider delivery/event id when present
	DeliveryID string `json:"deliveryId,omitempty" jsonschema:"description=Provider delivery identifier"`
	// Runtime signals that this execution uses the runtime provider path
	Runtime bool `json:"runtime,omitempty" jsonschema:"description=Whether the execution uses the runtime provider path"`
	// SkipCampaignEmailSync bypasses campaign_email sibling clearing during hook maintenance updates
	SkipCampaignEmailSync bool `json:"skipCampaignEmailSync,omitempty" jsonschema:"description=Bypass campaign email sibling clearing"`
	// SkipPrimaryDirectorySync bypasses primary_directory sibling clearing during hook maintenance updates
	SkipPrimaryDirectorySync bool `json:"skipPrimaryDirectorySync,omitempty" jsonschema:"description=Bypass primary directory sibling clearing"`
	// Workflow captures workflow linkage when the execution originates from a workflow
	Workflow *WorkflowMeta `json:"workflow,omitempty" jsonschema:"description=Workflow linkage when present"`
}

// integrationEntityType is the OperationContext entity type for integration installations
const integrationEntityType = "integration"

// NewOperationContext builds a gala.OperationContext for an integration execution,
// promoting the integration installation as the queryable entity and marshaling the
// integration source provenance into its Attributes
func NewOperationContext(ownerID, operation string, src IntegrationSource) gala.OperationContext {
	oc := gala.OperationContext{OwnerID: ownerID, Operation: operation}

	if src.IntegrationID != "" {
		oc.EntityID = src.IntegrationID
		oc.EntityType = integrationEntityType
	}

	_ = gala.SetAttributes(&oc, src)

	return oc
}

// IntegrationSourceFrom decodes the integration source provenance from an operation context
func IntegrationSourceFrom(oc gala.OperationContext) IntegrationSource {
	src, _ := gala.DecodeAttributes[IntegrationSource](oc)

	return src
}
