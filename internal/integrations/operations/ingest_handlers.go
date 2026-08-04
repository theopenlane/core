package operations

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/utils/contextx"
)

// directorySyncRunIDKey carries the directory sync run id to downstream ingest handlers across the
// durable dispatch hop (captured and restored via the codec registered in ContextCodecs)
var directorySyncRunIDKey = contextx.NewKey[string]()

// withDirectorySyncRunID returns ctx with the directory sync run ID stored for downstream ingest handlers
func withDirectorySyncRunID(ctx context.Context, id string) context.Context {
	return directorySyncRunIDKey.Set(ctx, id)
}

// directorySyncRunIDFromContext returns the directory sync run ID stored in ctx, or empty string if not set
func directorySyncRunIDFromContext(ctx context.Context) string {
	return directorySyncRunIDKey.GetOr(ctx, "")
}

// ingestIntegrationKey carries the resolved integration record to the schema handler's persist
// closure on the synchronous path, avoiding a redundant fetch when the caller already holds it
var ingestIntegrationKey = contextx.NewKey[*ent.Integration]()

// withIngestIntegration returns ctx with the integration record stored for the persist closure
func withIngestIntegration(ctx context.Context, integration *ent.Integration) context.Context {
	return ingestIntegrationKey.Set(ctx, integration)
}

// resolveIngestIntegration returns the integration record for the current ingest operation. On the
// synchronous path the record is read from ctx; on the asynchronous handler path it is fetched from
// the operation context entity id restored by the gala codec
func resolveIngestIntegration(ctx context.Context, client *ent.Client) (*ent.Integration, error) {
	if integration, ok := ingestIntegrationKey.Get(ctx); ok && integration != nil {
		return integration, nil
	}

	oc, ok := gala.OperationContextFromContext(ctx)
	if !ok || oc.EntityID == "" {
		return nil, ErrIngestIntegrationUnresolved
	}

	return client.Integration.Get(ctx, oc.EntityID)
}

// ingestSchemaOrder defines the registration order for ingest schema listeners
var ingestSchemaOrder = []string{
	entityops.SchemaActionPlan.Name,
	entityops.SchemaAsset.Name,
	entityops.SchemaCheckResult.Name,
	entityops.SchemaContact.Name,
	entityops.SchemaDirectoryAccount.Name,
	entityops.SchemaDirectoryGroup.Name,
	entityops.SchemaDirectoryMembership.Name,
	entityops.SchemaEntity.Name,
	entityops.SchemaFinding.Name,
	entityops.SchemaInternalPolicy.Name,
	entityops.SchemaProcedure.Name,
	entityops.SchemaRisk.Name,
	entityops.SchemaVulnerability.Name,
}

// ingestHandlers maps each supported ingest schema to its shared entityops schema handler. Each
// handler is built from the generated typed topic, the schema-specific input preparation, and the
// hand-written upsert persistence closure
var ingestHandlers = map[string]entityops.SchemaHandler{
	entityops.SchemaActionPlan.Name: buildIngestHandler(entityops.SchemaActionPlan, entityops.TopicActionPlan,
		func(oc gala.OperationContext, in ent.CreateActionPlanInput, through map[string][]string) entityops.ActionPlanIngestRequested {
			return entityops.ActionPlanIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.ActionPlanIngestRequested) (gala.OperationContext, ent.CreateActionPlanInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareActionPlanInput, persistActionPlanInput),
	entityops.SchemaAsset.Name: buildIngestHandler(entityops.SchemaAsset, entityops.TopicAsset,
		func(oc gala.OperationContext, in ent.CreateAssetInput, through map[string][]string) entityops.AssetIngestRequested {
			return entityops.AssetIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.AssetIngestRequested) (gala.OperationContext, ent.CreateAssetInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareAssetInput, persistAssetInput),
	entityops.SchemaCheckResult.Name: buildIngestHandler(entityops.SchemaCheckResult, entityops.TopicCheckResult,
		func(oc gala.OperationContext, in ent.CreateCheckResultInput, through map[string][]string) entityops.CheckResultIngestRequested {
			return entityops.CheckResultIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.CheckResultIngestRequested) (gala.OperationContext, ent.CreateCheckResultInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareCheckResultInput, persistCheckResultInput),
	entityops.SchemaContact.Name: buildIngestHandler(entityops.SchemaContact, entityops.TopicContact,
		func(oc gala.OperationContext, in ent.CreateContactInput, through map[string][]string) entityops.ContactIngestRequested {
			return entityops.ContactIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.ContactIngestRequested) (gala.OperationContext, ent.CreateContactInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareContactInput, persistContactInput),
	entityops.SchemaDirectoryAccount.Name: buildIngestHandler(entityops.SchemaDirectoryAccount, entityops.TopicDirectoryAccount,
		func(oc gala.OperationContext, in ent.CreateDirectoryAccountInput, through map[string][]string) entityops.DirectoryAccountIngestRequested {
			return entityops.DirectoryAccountIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.DirectoryAccountIngestRequested) (gala.OperationContext, ent.CreateDirectoryAccountInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareDirectoryAccountInput, persistDirectoryAccountInput),
	entityops.SchemaDirectoryGroup.Name: buildIngestHandler(entityops.SchemaDirectoryGroup, entityops.TopicDirectoryGroup,
		func(oc gala.OperationContext, in ent.CreateDirectoryGroupInput, through map[string][]string) entityops.DirectoryGroupIngestRequested {
			return entityops.DirectoryGroupIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.DirectoryGroupIngestRequested) (gala.OperationContext, ent.CreateDirectoryGroupInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareDirectoryGroupInput, persistDirectoryGroupInput),
	entityops.SchemaDirectoryMembership.Name: buildIngestHandler(entityops.SchemaDirectoryMembership, entityops.TopicDirectoryMembership,
		func(oc gala.OperationContext, in ent.CreateDirectoryMembershipInput, through map[string][]string) entityops.DirectoryMembershipIngestRequested {
			return entityops.DirectoryMembershipIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.DirectoryMembershipIngestRequested) (gala.OperationContext, ent.CreateDirectoryMembershipInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareDirectoryMembershipInput, persistDirectoryMembershipInput),
	entityops.SchemaEntity.Name: buildIngestHandler(entityops.SchemaEntity, entityops.TopicEntity,
		func(oc gala.OperationContext, in ent.CreateEntityInput, through map[string][]string) entityops.EntityIngestRequested {
			return entityops.EntityIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.EntityIngestRequested) (gala.OperationContext, ent.CreateEntityInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareEntityInput, persistEntityInput),
	entityops.SchemaFinding.Name: buildIngestHandler(entityops.SchemaFinding, entityops.TopicFinding,
		func(oc gala.OperationContext, in ent.CreateFindingInput, through map[string][]string) entityops.FindingIngestRequested {
			return entityops.FindingIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.FindingIngestRequested) (gala.OperationContext, ent.CreateFindingInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareFindingInput, persistFindingInput),
	entityops.SchemaInternalPolicy.Name: buildIngestHandler(entityops.SchemaInternalPolicy, entityops.TopicInternalPolicy,
		func(oc gala.OperationContext, in ent.CreateInternalPolicyInput, through map[string][]string) entityops.InternalPolicyIngestRequested {
			return entityops.InternalPolicyIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.InternalPolicyIngestRequested) (gala.OperationContext, ent.CreateInternalPolicyInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareInternalPolicyInput, persistInternalPolicyInput),
	entityops.SchemaProcedure.Name: buildIngestHandler(entityops.SchemaProcedure, entityops.TopicProcedure,
		func(oc gala.OperationContext, in ent.CreateProcedureInput, through map[string][]string) entityops.ProcedureIngestRequested {
			return entityops.ProcedureIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.ProcedureIngestRequested) (gala.OperationContext, ent.CreateProcedureInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareProcedureInput, persistProcedureInput),
	entityops.SchemaRisk.Name: buildIngestHandler(entityops.SchemaRisk, entityops.TopicRisk,
		func(oc gala.OperationContext, in ent.CreateRiskInput, through map[string][]string) entityops.RiskIngestRequested {
			return entityops.RiskIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.RiskIngestRequested) (gala.OperationContext, ent.CreateRiskInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareRiskInput, persistRiskInput),
	entityops.SchemaVulnerability.Name: buildIngestHandler(entityops.SchemaVulnerability, entityops.TopicVulnerability,
		func(oc gala.OperationContext, in ent.CreateVulnerabilityInput, through map[string][]string) entityops.VulnerabilityIngestRequested {
			return entityops.VulnerabilityIngestRequested{OperationContext: oc, Input: in, ThroughEdgeIDs: through}
		},
		func(e entityops.VulnerabilityIngestRequested) (gala.OperationContext, ent.CreateVulnerabilityInput, map[string][]string) {
			return e.OperationContext, e.Input, e.ThroughEdgeIDs
		},
		prepareVulnerabilityInput, persistVulnerabilityInput),
}

// buildIngestHandler assembles one entityops schema handler from the catalog entry, the generated
// typed topic, and the schema-specific preparation and persistence closures. Integration-scoped
// preparation runs inside the persist closure, where the integration record is resolved from context
func buildIngestHandler[TInput any, TEvent any](
	schema *entityops.Schema,
	topic gala.Topic[TEvent],
	wrap func(gala.OperationContext, TInput, map[string][]string) TEvent,
	unwrap func(TEvent) (gala.OperationContext, TInput, map[string][]string),
	prepare func(context.Context, TInput, *ent.Integration) TInput,
	persist func(context.Context, *ent.Client, *ent.Integration, TInput) (string, error),
) entityops.SchemaHandler {
	return entityops.BuildSchemaHandler(entityops.SchemaHandlerConfig[TInput, TEvent]{
		Schema:  schema,
		Topic:   topic,
		Prepare: func(_ context.Context, input TInput) TInput { return input },
		Wrap:    wrap,
		Unwrap:  unwrap,
		Persist: func(ctx context.Context, client *ent.Client, input TInput) (string, error) {
			integration, err := resolveIngestIntegration(ctx, client)
			if err != nil {
				return "", err
			}

			input = prepare(ctx, input, integration)

			return persist(ctx, client, integration, input)
		},
	})
}

// lookupIngestHandler returns the schema handler for one schema name
func lookupIngestHandler(schema string) (entityops.SchemaHandler, bool) {
	handler, ok := ingestHandlers[schema]

	return handler, ok
}

// RegisterIngestListeners attaches second-stage ingest listeners for all supported ingest schemas
func RegisterIngestListeners(runtime *gala.Gala) error {
	if runtime == nil {
		return ErrGalaRequired
	}

	for _, schema := range ingestSchemaOrder {
		handler, ok := lookupIngestHandler(schema)
		if !ok {
			return ErrIngestUnsupportedSchema
		}

		if err := handler.Register(runtime); err != nil {
			return err
		}
	}

	return nil
}

// emitMappedRecord looks up the schema handler and emits one mapped ingest record via Gala
func emitMappedRecord(ctx context.Context, runtime *gala.Gala, integration *ent.Integration, operationName string, record mappedIngestRecord, options IngestOptions) error {
	handler, ok := lookupIngestHandler(record.Schema)
	if !ok {
		return ErrIngestUnsupportedSchema
	}

	oc := buildIngestOperationContext(integration, options)
	headers := buildIngestHeaders(integration, operationName, record, options)

	return handler.Emit(ctx, runtime, oc, headers, record.Payload)
}

// persistMappedRecord looks up the schema handler and persists one mapped ingest record synchronously
func persistMappedRecord(ctx context.Context, db *ent.Client, integration *ent.Integration, schema string, payload json.RawMessage) (string, error) {
	handler, ok := lookupIngestHandler(schema)
	if !ok {
		return "", ErrIngestUnsupportedSchema
	}

	ctx = withIngestIntegration(ctx, integration)

	return handler.Persist(ctx, db, payload)
}

// buildIngestOperationContext builds the durable operation context for one ingest record, promoting
// the integration installation as the queryable entity and carrying integration provenance
func buildIngestOperationContext(integration *ent.Integration, options IngestOptions) gala.OperationContext {
	src := types.IntegrationSource{
		IntegrationID: integration.ID,
		DefinitionID:  integration.DefinitionID,
		RunID:         options.RunID,
		Webhook:       options.Webhook,
		Event:         options.WebhookEvent,
		DeliveryID:    options.DeliveryID,
		Workflow:      options.WorkflowMeta,
	}

	return types.NewOperationContext(integration.OwnerID, "", src)
}

// buildIngestHeaders assembles Gala message headers for one ingest record
func buildIngestHeaders(integration *ent.Integration, operationName string, record mappedIngestRecord, options IngestOptions) gala.Headers {
	tags := []string{integration.DefinitionID, "schema_" + strings.ToLower(record.Schema)}

	properties := map[string]string{
		"schema":         record.Schema,
		"integration_id": integration.ID,
		"definition_id":  integration.DefinitionID,
		"operation":      operationName,
		"variant":        record.Variant,
		"run_id":         options.RunID,
		"webhook":        options.Webhook,
		"webhook_event":  options.WebhookEvent,
		"delivery_id":    options.DeliveryID,
	}

	if options.WorkflowMeta != nil {
		properties["workflow_instance_id"] = options.WorkflowMeta.InstanceID
		properties["workflow_action_key"] = options.WorkflowMeta.ActionKey
	}

	return gala.Headers{
		Properties: lo.PickBy(properties, func(_ string, value string) bool {
			return value != ""
		}),
		Tags: tags,
	}
}

// prepareActionPlanInput applies integration-scoped defaults before emit or sync persistence
func prepareActionPlanInput(_ context.Context, input ent.CreateActionPlanInput, integration *ent.Integration) ent.CreateActionPlanInput {
	if input.OwnerID == nil && integration.OwnerID != "" {
		input.OwnerID = &integration.OwnerID
	}

	return input
}

// prepareAssetInput applies integration-scoped defaults before emit or sync persistence
func prepareAssetInput(_ context.Context, input ent.CreateAssetInput, integration *ent.Integration) ent.CreateAssetInput {
	return entityops.PrepareAssetInput(input, integration)
}

// prepareCheckResultInput applies integration-scoped defaults before emit or sync persistence
func prepareCheckResultInput(_ context.Context, input ent.CreateCheckResultInput, integration *ent.Integration) ent.CreateCheckResultInput {
	return entityops.PrepareCheckResultInput(input, integration)
}

// prepareContactInput applies integration-scoped defaults before emit or sync persistence
func prepareContactInput(_ context.Context, input ent.CreateContactInput, integration *ent.Integration) ent.CreateContactInput {
	input = entityops.PrepareContactInput(input, integration)

	if input.OwnerID == nil && integration.OwnerID != "" {
		input.OwnerID = &integration.OwnerID
	}

	return input
}

// prepareDirectoryAccountInput applies integration-scoped defaults before emit or sync persistence
func prepareDirectoryAccountInput(ctx context.Context, input ent.CreateDirectoryAccountInput, integration *ent.Integration) ent.CreateDirectoryAccountInput {
	input = entityops.PrepareDirectoryAccountInput(input, integration)

	dirSyncRunID := directorySyncRunIDFromContext(ctx)

	if input.DirectorySyncRunID == nil && dirSyncRunID != "" {
		input.DirectorySyncRunID = &dirSyncRunID
	}

	return input
}

// prepareDirectoryGroupInput applies integration-scoped defaults before emit or sync persistence
func prepareDirectoryGroupInput(ctx context.Context, input ent.CreateDirectoryGroupInput, integration *ent.Integration) ent.CreateDirectoryGroupInput {
	input = entityops.PrepareDirectoryGroupInput(input, integration)

	if input.OwnerID == nil && integration.OwnerID != "" {
		input.OwnerID = &integration.OwnerID
	}

	dirSyncRunID := directorySyncRunIDFromContext(ctx)

	if input.DirectorySyncRunID == "" && dirSyncRunID != "" {
		input.DirectorySyncRunID = dirSyncRunID
	}

	return input
}

// prepareDirectoryMembershipInput applies integration-scoped defaults before emit or sync persistence
func prepareDirectoryMembershipInput(ctx context.Context, input ent.CreateDirectoryMembershipInput, integration *ent.Integration) ent.CreateDirectoryMembershipInput {
	input = entityops.PrepareDirectoryMembershipInput(input, integration)

	if input.OwnerID == nil && integration.OwnerID != "" {
		input.OwnerID = &integration.OwnerID
	}

	dirSyncRunID := directorySyncRunIDFromContext(ctx)

	if input.DirectorySyncRunID == "" && dirSyncRunID != "" {
		input.DirectorySyncRunID = dirSyncRunID
	}

	return input
}

// prepareEntityInput applies integration-scoped defaults before emit or sync persistence
func prepareEntityInput(_ context.Context, input ent.CreateEntityInput, integration *ent.Integration) ent.CreateEntityInput {
	return entityops.PrepareEntityInput(input, integration)
}

// prepareFindingInput applies integration-scoped defaults before emit or sync persistence
func prepareFindingInput(_ context.Context, input ent.CreateFindingInput, integration *ent.Integration) ent.CreateFindingInput {
	return entityops.PrepareFindingInput(input, integration)
}

// prepareInternalPolicyInput applies integration-scoped defaults before emit or sync persistence
func prepareInternalPolicyInput(_ context.Context, input ent.CreateInternalPolicyInput, integration *ent.Integration) ent.CreateInternalPolicyInput {
	if input.OwnerID == nil && integration.OwnerID != "" {
		input.OwnerID = &integration.OwnerID
	}

	return input
}

// prepareProcedureInput applies integration-scoped defaults before emit or sync persistence
func prepareProcedureInput(_ context.Context, input ent.CreateProcedureInput, integration *ent.Integration) ent.CreateProcedureInput {
	if input.OwnerID == nil && integration.OwnerID != "" {
		input.OwnerID = &integration.OwnerID
	}

	return input
}

// prepareRiskInput applies integration-scoped defaults before emit or sync persistence
func prepareRiskInput(_ context.Context, input ent.CreateRiskInput, integration *ent.Integration) ent.CreateRiskInput {
	return entityops.PrepareRiskInput(input, integration)
}

// prepareVulnerabilityInput applies integration-scoped defaults before emit or sync persistence
func prepareVulnerabilityInput(_ context.Context, input ent.CreateVulnerabilityInput, integration *ent.Integration) ent.CreateVulnerabilityInput {
	return entityops.PrepareVulnerabilityInput(input, integration)
}
