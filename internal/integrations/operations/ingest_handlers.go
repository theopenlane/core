package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/riverqueue/river"
	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
)

func withDirectorySyncRunID(ctx context.Context, id string) context.Context {
	return gala.DirectorySyncRunIDKey.Set(ctx, id)
}

func directorySyncRunIDFromContext(ctx context.Context) string {
	return gala.DirectorySyncRunIDKey.GetOr(ctx, "")
}

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

var (
	bindIngestOnce sync.Once
	bindIngestErr  error
)

// bindIngestPersistence attaches the supported operation-owned persistence implementations to
// their generated schema capabilities.
func bindIngestPersistence() error {
	bindIngestOnce.Do(func() {
		binds := []func() error{
			func() error { return entityops.BindIngest(entityops.SchemaActionPlan, persistActionPlanInput) },
			func() error { return entityops.BindIngest(entityops.SchemaAsset, persistAssetInput) },
			func() error { return entityops.BindIngest(entityops.SchemaCheckResult, persistCheckResultInput) },
			func() error { return entityops.BindIngest(entityops.SchemaContact, persistContactInput) },
			func() error {
				return entityops.BindIngest(entityops.SchemaDirectoryAccount, func(ctx context.Context, db *ent.Client, integration *ent.Integration, input ent.CreateDirectoryAccountInput) (string, error) {
					return persistDirectoryAccountInput(ctx, db, integration, prepareDirectoryAccountInput(ctx, input))
				})
			},
			func() error {
				return entityops.BindIngest(entityops.SchemaDirectoryGroup, func(ctx context.Context, db *ent.Client, integration *ent.Integration, input ent.CreateDirectoryGroupInput) (string, error) {
					return persistDirectoryGroupInput(ctx, db, integration, prepareDirectoryGroupInput(ctx, input))
				})
			},
			func() error {
				return entityops.BindIngest(entityops.SchemaDirectoryMembership, func(ctx context.Context, db *ent.Client, integration *ent.Integration, input ent.CreateDirectoryMembershipInput) (string, error) {
					return persistDirectoryMembershipInput(ctx, db, integration, prepareDirectoryMembershipInput(ctx, input))
				})
			},
			func() error { return entityops.BindIngest(entityops.SchemaEntity, persistEntityInput) },
			func() error { return entityops.BindIngest(entityops.SchemaFinding, persistFindingInput) },
			func() error { return entityops.BindIngest(entityops.SchemaInternalPolicy, persistInternalPolicyInput) },
			func() error { return entityops.BindIngest(entityops.SchemaProcedure, persistProcedureInput) },
			func() error { return entityops.BindIngest(entityops.SchemaRisk, persistRiskInput) },
			func() error { return entityops.BindIngest(entityops.SchemaVulnerability, persistVulnerabilityInput) },
		}

		for _, bind := range binds {
			if bindIngestErr = bind(); bindIngestErr != nil {
				return
			}
		}
	})

	return bindIngestErr
}

// lookupIngestSchema returns the entityops schema when it supports mapped integration ingestion
func lookupIngestSchema(schema string) (*entityops.Schema, bool) {
	target, ok := entityops.LookupSchema(schema)
	if !ok || target.Ingest == nil {
		return nil, false
	}

	return target, true
}

// emitMappedRecord queues one durable schema-ingest command.
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
	}

	return schema.EmitIngest(ctx, runtime, buildIngestHeaders(integration, operationName, record, options), request)
}

// persistMappedRecord is retained for synchronous callers (SCIM, directory syncs, and tests).
// It shares the schema's generated preparation and bound persistence with the durable listener.
func persistMappedRecord(ctx context.Context, db *ent.Client, integration *ent.Integration, schema string, payload json.RawMessage) (string, error) {
	target, ok := lookupIngestSchema(schema)
	if !ok {
		return "", ErrIngestUnsupportedSchema
	}
	if err := bindIngestPersistence(); err != nil {
		return "", err
	}

	return target.PersistIngest(ctx, db, integration, payload)
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

// prepareDirectoryAccountInput stamps the directory sync run attribution the generated preparation
// cannot supply; installation-derived defaults are applied by the schema's prepare
func prepareDirectoryAccountInput(ctx context.Context, input ent.CreateDirectoryAccountInput) ent.CreateDirectoryAccountInput {
	if runID := directorySyncRunIDFromContext(ctx); input.DirectorySyncRunID == nil && runID != "" {
		input.DirectorySyncRunID = &runID
	}
	return input
}

// prepareDirectoryGroupInput stamps the directory sync run attribution the generated preparation
// cannot supply; installation-derived defaults are applied by the schema's prepare
func prepareDirectoryGroupInput(ctx context.Context, input ent.CreateDirectoryGroupInput) ent.CreateDirectoryGroupInput {
	if runID := directorySyncRunIDFromContext(ctx); input.DirectorySyncRunID == "" && runID != "" {
		input.DirectorySyncRunID = runID
	}
	return input
}

// prepareDirectoryMembershipInput stamps the directory sync run attribution the generated
// preparation cannot supply; installation-derived defaults are applied by the schema's prepare
func prepareDirectoryMembershipInput(ctx context.Context, input ent.CreateDirectoryMembershipInput) ent.CreateDirectoryMembershipInput {
	if runID := directorySyncRunIDFromContext(ctx); input.DirectorySyncRunID == "" && runID != "" {
		input.DirectorySyncRunID = runID
	}
	return input
}
