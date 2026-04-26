package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// IngestOptions carries the minimal ingest-time metadata needed by persistence
type IngestOptions struct {
	// DirectorySyncRunID groups all directory-related ingest records from one sync batch
	DirectorySyncRunID string
	// SkipDirectorySyncRunFinalization instructs the processor not to finalize the directory sync run after processing
	SkipDirectorySyncRunFinalization bool
	// Source identifies the mechanism that produced the ingest data (e.g. webhook, poll, manual)
	Source integrationgenerated.IntegrationIngestSource
	// RunID is a caller-supplied correlation identifier for the overall operation run
	RunID string
	// Webhook is the webhook name or identifier that triggered this ingest
	Webhook string
	// WebhookEvent is the event type reported by the webhook provider
	WebhookEvent string
	// DeliveryID is the provider-assigned delivery identifier; used for deduplication
	DeliveryID string
	// WorkflowMeta carries workflow instance context
	WorkflowMeta *types.WorkflowMeta
}

// IngestOptionsFromMetadata derives ingest options from execution metadata
func IngestOptionsFromMetadata(source integrationgenerated.IntegrationIngestSource, m types.ExecutionMetadata) IngestOptions {
	return IngestOptions{
		Source:       source,
		RunID:        m.RunID,
		Webhook:      m.Webhook,
		WebhookEvent: m.Event,
		DeliveryID:   m.DeliveryID,
		WorkflowMeta: m.Workflow,
	}
}

// installationFilterConfig holds per-installation CEL filter configuration stored in the integration's client config
type installationFilterConfig struct {
	// FilterExpr is a CEL expression evaluated against each ingest envelope; non-matching envelopes are dropped
	FilterExpr string `json:"filterExpr,omitempty"`
}

// mappedIngestRecord is the result of applying a mapping expression to one ingest envelope
type mappedIngestRecord struct {
	// Schema is the integration mapping schema name identifying the target ent type
	Schema string
	// Variant is the provider-specific sub-type within the schema (e.g. "user" vs "service_account")
	Variant string
	// Payload is the mapped JSON document ready for unmarshaling into the ent create input type
	Payload json.RawMessage
}

// directorySyncRunSchemas is the set of mapping schemas that require a directory sync run record
var directorySyncRunSchemas = map[string]struct{}{
	integrationgenerated.IntegrationMappingSchemaDirectoryAccount:    {},
	integrationgenerated.IntegrationMappingSchemaDirectoryGroup:      {},
	integrationgenerated.IntegrationMappingSchemaDirectoryMembership: {},
}

// EmitPayloadSets transforms one batch of mapped payload sets and dispatches them through the appropriate ingest path
func EmitPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) error {
	if needsDirectorySyncRun(contracts) {
		// directory sync runs must finalize in the same process that creates them
		return ProcessPayloadSets(ctx, ic, contracts, payloadSets, options)
	}

	if ic.Runtime == nil {
		return ErrGalaRequired
	}

	return applyPayloadSets(ctx, ic, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		return emitMappedRecord(handleCtx, ic.Runtime, ic.Integration, operationName, record, options)
	})
}

// ProcessPayloadSets persists one batch of mapped payload sets synchronously
func ProcessPayloadSets(ctx context.Context, ic IngestContext, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) error {
	return applyPayloadSets(ctx, ic, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		return persistMappedRecord(handleCtx, ic.DB, ic.Integration, record.Schema, record.Payload)
	})
}

// applyPayloadSets is the shared core for both async emit and sync persist paths
func applyPayloadSets(ctx context.Context, ic IngestContext, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions, handle func(context.Context, mappedIngestRecord) error) (err error) {
	definition, ok := ic.Registry.Definition(ic.Integration.DefinitionID)
	if !ok {
		return ErrIngestDefinitionNotFound
	}

	installationFilterExpr, err := resolveInstallationFilterExpr(ic.Integration)
	if err != nil {
		return ErrIngestInstallationFilterConfigInvalid
	}

	directorySync := needsDirectorySyncRun(contracts)
	markRemoved := hasExhaustiveDirectoryAccountContract(contracts)
	syncStartedAt := time.Now()

	directorySyncRunID := options.DirectorySyncRunID
	if directorySyncRunID == "" && directorySync {
		directorySyncRunID, err = createDirectorySyncRun(ctx, ic.DB, ic.Integration)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
	}

	if directorySyncRunID != "" {
		ctx = withDirectorySyncRunID(ctx, directorySyncRunID)
	}

	if directorySyncRunID != "" && directorySync && !options.SkipDirectorySyncRunFinalization {
		integrationID := ic.Integration.ID
		defer func() {
			if finalizeErr := finalizeDirectorySyncRun(ctx, ic.DB, directorySyncRunID, integrationID, syncStartedAt, markRemoved, err); finalizeErr != nil {
				err = errors.Join(err, finalizeErr)
			}
		}()
	}

	for _, payloadSet := range payloadSets {
		if !contractIncludesSchema(contracts, payloadSet.Schema) {
			return ErrIngestSchemaNotDeclared
		}

		if _, ok := integrationgenerated.IntegrationMappingSchemas[payloadSet.Schema]; !ok {
			return ErrIngestSchemaNotFound
		}

		for _, envelope := range payloadSet.Envelopes {
			record, include, err := mapIngestRecord(ctx, definition, payloadSet.Schema, envelope, installationFilterExpr)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("schema", payloadSet.Schema).Msg("integration: error mapping ingest record")
				return err
			}

			if !include {
				continue
			}

			if err = handle(ctx, record); err != nil {
				return err
			}
		}
	}

	return nil
}

// mapIngestRecord applies user-defined CEL expression filters over the data envelope so that we can filter out the data we import
func mapIngestRecord(ctx context.Context, definition types.Definition, schema string, envelope types.MappingEnvelope, installationFilterExpr string) (mappedIngestRecord, bool, error) {
	mapping, found := findMapping(definition.Mappings, schema, envelope.Variant)
	if !found {
		return mappedIngestRecord{}, false, ErrIngestMappingNotFound
	}

	matched, err := envelopeIncludedByFilters(ctx, installationFilterExpr, mapping.FilterExpr, envelope)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("integration: ingest filter failed")
		return mappedIngestRecord{}, false, ErrIngestFilterFailed
	}
	if !matched {
		return mappedIngestRecord{}, false, nil
	}

	mapped, err := providerkit.EvalMap(ctx, mapping.MapExpr, envelope)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("integration: error transforming ingest")

		return mappedIngestRecord{}, false, fmt.Errorf("%w: %w", ErrIngestTransformFailed, err)
	}

	return mappedIngestRecord{
		Schema:  schema,
		Variant: envelope.Variant,
		Payload: mapped,
	}, true, nil
}

// resolveInstallationFilterExpr pulls the per-installation filter expression out of the config
func resolveInstallationFilterExpr(installation *ent.Integration) (string, error) {
	var cfg installationFilterConfig
	if err := jsonx.UnmarshalIfPresent(installation.Config.ClientConfig, &cfg); err != nil {
		return "", err
	}

	return cfg.FilterExpr, nil
}

// envelopeIncludedByFilters evaluates the installation-level and mapping-level filter expressions against the data envelope
func envelopeIncludedByFilters(ctx context.Context, installationFilterExpr string, mappingFilterExpr string, envelope types.MappingEnvelope) (bool, error) {
	matched, err := providerkit.EvalFilter(ctx, installationFilterExpr, envelope)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, nil
	}

	return providerkit.EvalFilter(ctx, mappingFilterExpr, envelope)
}

// findMapping looks up the mapping spec for the given schema and variant
func findMapping(mappings []types.MappingRegistration, schema string, variant string) (types.MappingOverride, bool) {
	mapping, ok := lo.Find(mappings, func(mapping types.MappingRegistration) bool {
		return mapping.Schema == schema && mapping.Variant == variant
	})
	if !ok {
		return types.MappingOverride{}, false
	}

	return mapping.Spec, true
}

// contractIncludesSchema checks whether the given list of contracts includes a contract for the given schema
func contractIncludesSchema(contracts []types.IngestContract, schema string) bool {
	return lo.ContainsBy(contracts, func(contract types.IngestContract) bool {
		return contract.Schema == schema
	})
}

// needsDirectorySyncRun checks whether any of the given contracts require a directory sync run to be created
func needsDirectorySyncRun(contracts []types.IngestContract) bool {
	return lo.ContainsBy(contracts, func(contract types.IngestContract) bool {
		_, ok := directorySyncRunSchemas[contract.Schema]
		return ok
	})
}

// createDirectorySyncRun creates a new directory sync run in the database and returns its ID so that we can pass it down into the ingest context
func createDirectorySyncRun(ctx context.Context, db *ent.Client, installation *ent.Integration) (string, error) {
	create := db.DirectorySyncRun.Create().
		SetIntegrationID(installation.ID).
		SetStatus(enums.DirectorySyncRunStatusRunning)

	if installation.OwnerID != "" {
		create.SetOwnerID(installation.OwnerID)
	}
	if installation.PlatformID != "" {
		create.SetPlatformID(installation.PlatformID)
	}

	run, err := create.Save(ctx)
	if err != nil {
		return "", err
	}

	return run.ID, nil
}

// finalizeDirectorySyncRun marks the directory sync run as completed or failed, and when markRemoved
// is true and the sync succeeded, marks any directory accounts not seen during this sync as deleted
func finalizeDirectorySyncRun(ctx context.Context, db *ent.Client, directorySyncRunID string, integrationID string, syncStartedAt time.Time, markRemoved bool, ingestErr error) error {
	update := db.DirectorySyncRun.UpdateOneID(directorySyncRunID).
		SetCompletedAt(time.Now())

	if ingestErr != nil {
		update.SetStatus(enums.DirectorySyncRunStatusFailed)
		update.SetError(ingestErr.Error())
	} else {
		update.SetStatus(enums.DirectorySyncRunStatusCompleted)
		update.ClearError()
	}

	if err := update.Exec(ctx); err != nil {
		return err
	}

	if ingestErr == nil && markRemoved {
		return markRemovedDirectoryAccounts(ctx, db, integrationID, syncStartedAt)
	}

	return nil
}

// hasExhaustiveDirectoryAccountContract reports whether any contract in the list declares an
// exhaustive directory account sync, meaning records absent from the sync should be marked removed
func hasExhaustiveDirectoryAccountContract(contracts []types.IngestContract) bool {
	return lo.ContainsBy(contracts, func(c types.IngestContract) bool {
		return c.Schema == integrationgenerated.IntegrationMappingSchemaDirectoryAccount && c.ExhaustiveSync
	})
}

// wrapIngestPersistError wraps the known errors from persistence operations so we don't need the same boilerplate in multiple functions
func wrapIngestPersistError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case ent.IsValidationError(err):
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	case ent.IsNotSingular(err), ent.IsConstraintError(err):
		return fmt.Errorf("%w: %w", ErrIngestUpsertConflict, err)
	default:
		return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
	}
}
