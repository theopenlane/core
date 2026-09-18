package operations

import (
	"context"
	"fmt"
	"strings"

	"github.com/riverqueue/river"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// resolveIngestIntegration loads the installation referenced by the durable operation context
func resolveIngestIntegration(ctx context.Context, client *ent.Client) (*ent.Integration, error) {
	oc, ok := gala.OperationContextFromContext(ctx)
	if !ok || oc.EntityID == "" {
		return nil, ErrIngestIntegrationUnresolved
	}

	integration, err := ResolveIntegration(ctx, client, oc.EntityID, "", "")
	if ent.IsNotFound(err) {
		return nil, river.JobCancel(fmt.Errorf("%w: %w", ErrIngestIntegrationRemoved, err))
	}

	return integration, err
}

// lookupIngestSchema returns the entityops schema when it supports mapped integration ingestion
func lookupIngestSchema(schema string) (*entityops.Schema, bool) {
	target, ok := entityops.LookupSchema(schema)
	if !ok || target.Ingest == nil {
		return nil, false
	}

	return target, true
}

// emitMappedRecord queues one durable schema-ingest command
func emitMappedRecord(ctx context.Context, runtime *gala.Gala, integration *ent.Integration, operationName string, record mappedIngestRecord, options IngestOptions) error {
	schema, ok := lookupIngestSchema(record.Schema)
	if !ok {
		return ErrIngestUnsupportedSchema
	}
	if runtime == nil {
		return ErrGalaRequired
	}

	request := entityops.IngestRequest{
		OperationContext: buildIngestOperationContext(integration, options),
		Input:            record.Payload,
		RunID:            options.RunID,
	}

	return schema.EmitIngest(ctx, runtime, buildIngestHeaders(integration, operationName, record, options), request)
}

func buildIngestOperationContext(integration *ent.Integration, options IngestOptions) gala.OperationContext {
	src := types.IntegrationSource{IntegrationID: integration.ID, DefinitionID: integration.DefinitionID, RunID: options.RunID, Webhook: options.Webhook, Event: options.WebhookEvent, DeliveryID: options.DeliveryID, Workflow: options.WorkflowMeta}
	return types.NewOperationContext(integration.OwnerID, "", src)
}

func buildIngestHeaders(integration *ent.Integration, operationName string, record mappedIngestRecord, options IngestOptions) gala.Headers {
	properties := map[string]string{"schema": record.Schema, "integration_id": integration.ID, "definition_id": integration.DefinitionID, "operation": operationName, "variant": record.Variant, "run_id": options.RunID, "webhook": options.Webhook, "webhook_event": options.WebhookEvent, "delivery_id": options.DeliveryID}
	if options.WorkflowMeta != nil {
		properties["workflow_instance_id"] = options.WorkflowMeta.InstanceID
		properties["workflow_action_key"] = options.WorkflowMeta.ActionKey
	}
	return gala.Headers{Properties: lo.PickBy(properties, func(_ string, value string) bool { return value != "" }), Tags: []string{integration.DefinitionID, "schema_" + strings.ToLower(record.Schema)}}
}
