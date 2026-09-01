package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorysyncrun"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/logx"
)

// IngestOptions carries the minimal ingest-time metadata needed by persistence
type IngestOptions struct {
	// DirectorySyncRunID groups all directory-related ingest records from one sync batch
	DirectorySyncRunID string
	// SkipDirectorySyncRunFinalization instructs the processor not to finalize the directory sync run after processing
	SkipDirectorySyncRunFinalization bool
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

// RecordFailure identifies one mapped record that could not be imported
type RecordFailure struct {
	// Schema is the mapping schema name
	Schema string
	// Resource is the provider resource identifier
	Resource string
	// Err is the underlying failure
	Err error
}

// IngestResult reports record-level work completed by a payload batch
type IngestResult struct {
	Attempted int
	// Persisted counts records written synchronously
	Persisted int
	// Accepted counts records queued durably; acceptance does not mean persistence
	Accepted int
	// Filtered counts records excluded by configured filters
	Filtered int
	// Succeeded is the combined successful handling count
	Succeeded int
	// Failed counts records that could not be imported
	Failed int
	// Failures lists each failed record with its cause
	Failures []RecordFailure
}

// IngestOptionsFromOperationContext derives ingest options from an integration operation context
func IngestOptionsFromOperationContext(oc gala.OperationContext) IngestOptions {
	src := types.IntegrationSourceFrom(oc)

	return IngestOptions{
		RunID:        src.RunID,
		Webhook:      src.Webhook,
		WebhookEvent: src.Event,
		DeliveryID:   src.DeliveryID,
		WorkflowMeta: src.Workflow,
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
	entityops.SchemaDirectoryAccount.Name:    {},
	entityops.SchemaDirectoryGroup.Name:      {},
	entityops.SchemaDirectoryMembership.Name: {},
}

// ProcessPayloadSets persists one batch of mapped payload sets synchronously; record
// failures are skipped and reported in the result, never the error
func ProcessPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) (IngestResult, error) {
	result, err := applyPayloadSets(ctx, ic, operationName, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		_, err := persistMappedRecord(handleCtx, ic.DB, ic.Integration, record.Schema, record.Payload)

		return err
	})

	result.Persisted = result.Succeeded

	return result, err
}

// EmitPayloadSets queues each non-directory record for durable schema ingest; directory syncs
// stay in-process because finalization and removal inference must follow persistence.
// Queued records report as Accepted, never Persisted
func EmitPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) (IngestResult, error) {
	if needsDirectorySyncRun(contracts) {
		return ProcessPayloadSets(ctx, ic, operationName, contracts, payloadSets, options)
	}

	if ic.Runtime == nil {
		return IngestResult{}, ErrGalaRequired
	}

	result, err := applyPayloadSets(ctx, ic, operationName, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		return emitMappedRecord(handleCtx, ic.Runtime, ic.Integration, operationName, record, options)
	})

	result.Accepted = result.Succeeded

	return result, err
}

// applyPayloadSets is the shared core for both async emit and sync persist paths
func applyPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions, handle func(context.Context, mappedIngestRecord) error) (result IngestResult, err error) {
	definition, ok := ic.Registry.Definition(ic.Integration.DefinitionID)
	if !ok {
		return result, ErrIngestDefinitionNotFound
	}

	installationFilterExpr, err := resolveInstallationFilterExpr(ic.Integration, definition, operationName)
	if err != nil {
		return result, ErrIngestInstallationFilterConfigInvalid
	}

	directorySync := needsDirectorySyncRun(contracts)

	directorySyncRunID := options.DirectorySyncRunID
	if directorySyncRunID == "" && directorySync {
		directorySyncRunID, err = createDirectorySyncRun(ctx, ic.DB, ic.Integration)
		if err != nil {
			return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
	}

	if directorySyncRunID != "" {
		ctx = withDirectorySyncRunID(ctx, directorySyncRunID)
	}

	if directorySyncRunID != "" && directorySync && !options.SkipDirectorySyncRunFinalization {
		defer func() {
			if finalizeErr := finalizeDirectorySyncRun(ctx, ic.DB, directorySyncRunID, err); finalizeErr != nil {
				err = errors.Join(err, finalizeErr)
			}
		}()
	}

	var attempted, failed int
	var membershipSetSeen bool
	var recordFailures []RecordFailure
	membershipsComplete := true

	for _, payloadSet := range payloadSets {
		if !contractIncludesSchema(contracts, payloadSet.Schema) {
			return result, ErrIngestSchemaNotDeclared
		}

		sourceSchema, ok := entityops.LookupSchema(payloadSet.Schema)
		if !ok {
			return result, ErrIngestSchemaNotFound
		}

		if payloadSet.Schema == entityops.SchemaDirectoryMembership.Name {
			membershipSetSeen = true
			membershipsComplete = membershipsComplete && payloadSet.SnapshotComplete
		}

		for _, envelope := range payloadSet.Envelopes {
			attempted++
			envCtx := logx.WithFields(ctx, map[string]any{"schema": payloadSet.Schema, "resource": envelope.Resource})

			mapping, found := findMapping(definition.Mappings, payloadSet.Schema, envelope.Variant)
			if !found {
				logx.FromContext(envCtx).Error().Err(ErrIngestMappingNotFound).Msg("error mapping ingest record")

				failed++
				recordFailures = append(recordFailures, RecordFailure{Schema: payloadSet.Schema, Resource: envelope.Resource, Err: ErrIngestMappingNotFound})

				continue
			}

			record, include, mapErr := mapIngestRecord(envCtx, mapping, payloadSet.Schema, envelope, installationFilterExpr)
			if mapErr != nil {
				logx.FromContext(envCtx).Error().Err(mapErr).Msg("error mapping ingest record")

				failed++
				recordFailures = append(recordFailures, RecordFailure{Schema: payloadSet.Schema, Resource: envelope.Resource, Err: mapErr})

				continue
			}

			if !include {
				result.Filtered++
				continue
			}

			// inject the mapping's cross-object links into the create input, so the record is
			// created (or emitted for async creation) with its edges already set; link rules are
			// declared on the definition's mapping and validated at registration
			record.Payload, err = injectLinks(envCtx, ic.DB, ic.Integration.OwnerID, mapping.Links, sourceSchema, record.Payload)
			if err != nil {
				logx.FromContext(envCtx).Error().Err(err).Msg("ingest link injection failed")

				attempted++
				failed++
				recordFailures = append(recordFailures, RecordFailure{Schema: payloadSet.Schema, Resource: envelope.Resource, Err: err})

				continue
			}

			if handleErr := handle(envCtx, record); handleErr != nil {
				logx.FromContext(envCtx).Error().Err(handleErr).Msg("ingest persist failed")

				failed++
				recordFailures = append(recordFailures, RecordFailure{Schema: payloadSet.Schema, Resource: envelope.Resource, Err: handleErr})
			} else {
				result.Succeeded++
			}
		}
	}

	if failed > 0 {
		// skipped records are not retried; failing the job would reprocess the whole batch
		logx.FromContext(ctx).Warn().Int("failed", failed).Int("attempted", attempted).Msg("ingest skipped records that could not be imported")
	}

	result.Attempted = attempted
	result.Failed = failed
	result.Failures = recordFailures

	// only a fully-confirmed complete snapshot authorizes removal inference: a skipped record
	// risks a false removal, and partial sources never carry full membership state
	if directorySync && membershipSetSeen && membershipsComplete && !options.SkipDirectorySyncRunFinalization && failed == 0 {
		if err := markUnconfirmedDirectoryMembershipsRemoved(ctx, ic.DB, ic.Integration.ID, directorySyncRunID); err != nil {
			return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
	}

	return result, nil
}

// markUnconfirmedDirectoryMembershipsRemoved records the removal side of the sync delta by
// stamping removed_at on active memberships for the integration that were not confirmed by
// the completed sync run
func markUnconfirmedDirectoryMembershipsRemoved(ctx context.Context, db *ent.Client, integrationID string, runID string) error {
	current, err := isCurrentDirectorySyncRun(ctx, db, runID)
	if err != nil {
		return err
	}
	if !current {
		logx.FromContext(ctx).Info().Str("directory_sync_run_id", runID).Msg("skipping membership removal for stale directory sync run")
		return nil
	}

	removed, err := db.DirectoryMembership.Update().
		Where(directorymembership.IntegrationID(integrationID),
			directorymembership.RemovedAtIsNil(),
			directorymembership.Or(
				directorymembership.LastConfirmedRunIDIsNil(),
				directorymembership.LastConfirmedRunIDNEQ(runID),
			)).
		SetRemovedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return err
	}

	if removed > 0 {
		logx.FromContext(ctx).Info().Int("removed_count", removed).Str("directory_sync_run_id", runID).Msg("marked directory memberships removed after sync run comparison")
	}

	return nil
}

// mapIngestRecord applies the resolved mapping's filters and map expression to one data envelope,
// returning the mapped record and whether the envelope passed the include filters
func mapIngestRecord(ctx context.Context, mapping types.MappingOverride, schema string, envelope types.MappingEnvelope, installationFilterExpr string) (mappedIngestRecord, bool, error) {
	matched, err := envelopeIncludedByFilters(ctx, installationFilterExpr, mapping.FilterExpr, envelope)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("ingest filter failed")
		return mappedIngestRecord{}, false, ErrIngestFilterFailed
	}
	if !matched {
		return mappedIngestRecord{}, false, nil
	}

	mapped, err := providerkit.EvalMap(ctx, mapping.MapExpr, envelope)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("ingest transform failed")

		return mappedIngestRecord{}, false, fmt.Errorf("%w: %w", ErrIngestTransformFailed, err)
	}

	return mappedIngestRecord{
		Schema:  schema,
		Variant: envelope.Variant,
		Payload: mapped,
	}, true, nil
}

// resolveInstallationFilterExpr pulls the filter expression for the current operation out of the
// installation config. When the operation declares a ConfigResolver, it extracts the
// operation-specific config section first (supporting nested UserInput structures like
// directorySync.filterExpr); otherwise it falls back to a top-level filterExpr in ClientConfig
func resolveInstallationFilterExpr(installation *ent.Integration, definition types.Definition, operationName string) (string, error) {
	if operationName != "" {
		if op, ok := lo.Find(definition.Operations, func(o types.OperationRegistration) bool { return o.Name == operationName }); ok && op.ConfigResolver != nil {
			var cfg installationFilterConfig
			if err := jsonx.UnmarshalIfPresent(op.ConfigResolver(installation.Config.ClientConfig), &cfg); err != nil {
				return "", err
			}

			if cfg.FilterExpr != "" {
				return cfg.FilterExpr, nil
			}
		}
	}

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
func finalizeDirectorySyncRun(ctx context.Context, db *ent.Client, directorySyncRunID string, ingestErr error) error {
	current, err := isCurrentDirectorySyncRun(ctx, db, directorySyncRunID)
	if err != nil {
		return err
	}
	if !current {
		logx.FromContext(ctx).Info().Str("directory_sync_run_id", directorySyncRunID).Msg("skipping stale directory sync run finalization")
		return nil
	}

	update := db.DirectorySyncRun.UpdateOneID(directorySyncRunID).
		SetCompletedAt(time.Now())

	if ingestErr != nil {
		update.SetStatus(enums.DirectorySyncRunStatusFailed)
		update.SetError(ingestErr.Error())
	} else {
		update.SetStatus(enums.DirectorySyncRunStatusCompleted)
		update.ClearError()
	}

	return update.Exec(ctx)
}

func isCurrentDirectorySyncRun(ctx context.Context, db *ent.Client, runID string) (bool, error) {
	run, err := db.DirectorySyncRun.Query().Where(directorysyncrun.ID(runID)).Only(ctx)
	if err != nil {
		return false, err
	}

	latest, err := db.DirectorySyncRun.Query().
		Where(directorysyncrun.IntegrationID(run.IntegrationID)).
		Order(directorysyncrun.ByStartedAt(sql.OrderDesc())).
		First(ctx)
	if err != nil {
		return false, err
	}

	return sameDirectorySyncRun(run.ID, latest.ID), nil
}

func sameDirectorySyncRun(runID, latestRunID string) bool {
	return runID != "" && runID == latestRunID
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
