package operations

import (
	"context"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	intobvs "github.com/theopenlane/core/v2/internal/integrations/observability"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

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
	properties := map[string]string{"schema": record.Schema, intobvs.FieldIntegrationID: integration.ID, intobvs.FieldDefinitionID: integration.DefinitionID, "operation": operationName, "variant": record.Variant, intobvs.FieldRunID: options.RunID, "webhook": options.Webhook, "webhook_event": options.WebhookEvent, "delivery_id": options.DeliveryID}
	if options.WorkflowMeta != nil {
		properties["workflow_instance_id"] = options.WorkflowMeta.InstanceID
		properties["workflow_action_key"] = options.WorkflowMeta.ActionKey
	}
	return gala.Headers{Properties: lo.PickBy(properties, func(_ string, value string) bool { return value != "" }), Tags: []string{integration.DefinitionID, "schema_" + strings.ToLower(record.Schema)}}
}
