package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/riverqueue/river"

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

	integration, err := client.Integration.Get(ctx, oc.EntityID)
	if ent.IsNotFound(err) {
		// the installation was removed while record jobs were still queued; retrying can
		// never succeed, so cancel the job instead of burning River attempts
		return nil, river.JobCancel(fmt.Errorf("%w: %w", ErrIngestIntegrationRemoved, err))
	}

	return integration, err
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

// ingestHandlers maps each supported ingest schema to its generated entityops schema handler,
// built from the schema-specific input preparation and the hand-written upsert persistence closure
var ingestHandlers = map[string]entityops.SchemaHandler{
	entityops.SchemaActionPlan.Name:          entityops.ActionPlanIngestHandler(ingestPersistStatic(entityops.PrepareActionPlanInput, persistActionPlanInput)),
	entityops.SchemaAsset.Name:               entityops.AssetIngestHandler(ingestPersistStatic(entityops.PrepareAssetInput, persistAssetInput)),
	entityops.SchemaCheckResult.Name:         entityops.CheckResultIngestHandler(ingestPersistStatic(entityops.PrepareCheckResultInput, persistCheckResultInput)),
	entityops.SchemaContact.Name:             entityops.ContactIngestHandler(ingestPersistStatic(entityops.PrepareContactInput, persistContactInput)),
	entityops.SchemaDirectoryAccount.Name:    entityops.DirectoryAccountIngestHandler(ingestPersist(prepareDirectoryAccountInput, persistDirectoryAccountInput)),
	entityops.SchemaDirectoryGroup.Name:      entityops.DirectoryGroupIngestHandler(ingestPersist(prepareDirectoryGroupInput, persistDirectoryGroupInput)),
	entityops.SchemaDirectoryMembership.Name: entityops.DirectoryMembershipIngestHandler(ingestPersist(prepareDirectoryMembershipInput, persistDirectoryMembershipInput)),
	entityops.SchemaEntity.Name:              entityops.EntityIngestHandler(ingestPersistStatic(entityops.PrepareEntityInput, persistEntityInput)),
	entityops.SchemaFinding.Name:             entityops.FindingIngestHandler(ingestPersistStatic(entityops.PrepareFindingInput, persistFindingInput)),
	entityops.SchemaInternalPolicy.Name:      entityops.InternalPolicyIngestHandler(ingestPersistStatic(entityops.PrepareInternalPolicyInput, persistInternalPolicyInput)),
	entityops.SchemaProcedure.Name:           entityops.ProcedureIngestHandler(ingestPersistStatic(entityops.PrepareProcedureInput, persistProcedureInput)),
	entityops.SchemaRisk.Name:                entityops.RiskIngestHandler(ingestPersistStatic(entityops.PrepareRiskInput, persistRiskInput)),
	entityops.SchemaVulnerability.Name:       entityops.VulnerabilityIngestHandler(ingestPersistStatic(entityops.PrepareVulnerabilityInput, persistVulnerabilityInput)),
}

// ingestPersist adapts a schema's prepare and persist closures into the generated handler's
// persist shape, resolving the integration record from context before preparation
func ingestPersist[TInput any](prepare func(context.Context, TInput, *ent.Integration) TInput, persist func(context.Context, *ent.Client, *ent.Integration, TInput) (string, error)) func(context.Context, *ent.Client, TInput) (string, error) {
	return func(ctx context.Context, client *ent.Client, input TInput) (string, error) {
		integration, err := resolveIngestIntegration(ctx, client)
		if err != nil {
			return "", err
		}

		input = prepare(ctx, input, integration)

		return persist(ctx, client, integration, input)
	}
}

// ingestPersistStatic adapts context-free prepare functions so the generated Prepare functions pass directly
func ingestPersistStatic[TInput any](prepare func(TInput, *ent.Integration) TInput, persist func(context.Context, *ent.Client, *ent.Integration, TInput) (string, error)) func(context.Context, *ent.Client, TInput) (string, error) {
	return ingestPersist(func(_ context.Context, input TInput, integration *ent.Integration) TInput {
		return prepare(input, integration)
	}, persist)
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

// prepareDirectoryAccountInput applies integration-scoped defaults and the directory sync run id
// carried in context before emit or sync persistence
func prepareDirectoryAccountInput(ctx context.Context, input ent.CreateDirectoryAccountInput, integration *ent.Integration) ent.CreateDirectoryAccountInput {
	input = entityops.PrepareDirectoryAccountInput(input, integration)

	dirSyncRunID := directorySyncRunIDFromContext(ctx)

	if input.DirectorySyncRunID == nil && dirSyncRunID != "" {
		input.DirectorySyncRunID = &dirSyncRunID
	}

	return input
}

// prepareDirectoryGroupInput applies integration-scoped defaults and the directory sync run id
// carried in context before emit or sync persistence
func prepareDirectoryGroupInput(ctx context.Context, input ent.CreateDirectoryGroupInput, integration *ent.Integration) ent.CreateDirectoryGroupInput {
	input = entityops.PrepareDirectoryGroupInput(input, integration)

	dirSyncRunID := directorySyncRunIDFromContext(ctx)

	if input.DirectorySyncRunID == "" && dirSyncRunID != "" {
		input.DirectorySyncRunID = dirSyncRunID
	}

	return input
}

// prepareDirectoryMembershipInput applies integration-scoped defaults and the directory sync run id
// carried in context before emit or sync persistence
func prepareDirectoryMembershipInput(ctx context.Context, input ent.CreateDirectoryMembershipInput, integration *ent.Integration) ent.CreateDirectoryMembershipInput {
	input = entityops.PrepareDirectoryMembershipInput(input, integration)

	dirSyncRunID := directorySyncRunIDFromContext(ctx)

	if input.DirectorySyncRunID == "" && dirSyncRunID != "" {
		input.DirectorySyncRunID = dirSyncRunID
	}

	return input
}
