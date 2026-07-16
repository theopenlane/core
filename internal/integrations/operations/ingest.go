package operations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/directorymembership"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// IngestOptions carries the minimal ingest-time metadata needed by persistence
type IngestOptions struct {
	// DirectorySyncRunID groups all directory-related ingest records from one sync batch
	DirectorySyncRunID string
	// SkipDirectorySyncRunFinalization instructs the processor not to finalize the directory sync run after processing
	SkipDirectorySyncRunFinalization bool
	// CompleteDirectorySnapshot signals that the batch carries the provider's complete directory
	// membership state, authorizing removal inference for memberships the batch did not confirm;
	// partial sources (webhooks, scim, single-record pushes) leave this false
	CompleteDirectorySnapshot bool
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

// EmitPayloadSets transforms one batch of mapped payload sets and dispatches them through the appropriate ingest path
func EmitPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) error {
	if needsDirectorySyncRun(contracts) {
		// directory sync runs must finalize in the same process that creates them
		return ProcessPayloadSets(ctx, ic, operationName, contracts, payloadSets, options)
	}

	if ic.Runtime == nil {
		return ErrGalaRequired
	}

	return applyPayloadSets(ctx, ic, operationName, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		return emitMappedRecord(handleCtx, ic.Runtime, ic.Integration, operationName, record, options)
	})
}

// ProcessPayloadSets persists one batch of mapped payload sets synchronously. Cross-object links are
// resolved and injected into each create input upstream in applyPayloadSets, so persistence creates
// each record with its edges already set
func ProcessPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) error {
	return applyPayloadSets(ctx, ic, operationName, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		_, err := persistMappedRecord(handleCtx, ic.DB, ic.Integration, record.Schema, record.Payload)

		return err
	})
}

// applyPayloadSets is the shared core for both async emit and sync persist paths
func applyPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions, handle func(context.Context, mappedIngestRecord) error) (err error) {
	definition, ok := ic.Registry.Definition(ic.Integration.DefinitionID)
	if !ok {
		return ErrIngestDefinitionNotFound
	}

	installationFilterExpr, err := resolveInstallationFilterExpr(ic.Integration, definition, operationName)
	if err != nil {
		return ErrIngestInstallationFilterConfigInvalid
	}

	directorySync := needsDirectorySyncRun(contracts)

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
		defer func() {
			if finalizeErr := finalizeDirectorySyncRun(ctx, ic.DB, directorySyncRunID, err); finalizeErr != nil {
				err = errors.Join(err, finalizeErr)
			}
		}()
	}

	var attempted, failed int
	var membershipSetSeen bool

	for _, payloadSet := range payloadSets {
		if !contractIncludesSchema(contracts, payloadSet.Schema) {
			return ErrIngestSchemaNotDeclared
		}

		if _, ok := entityops.LookupSchema(payloadSet.Schema); !ok {
			return ErrIngestSchemaNotFound
		}

		if payloadSet.Schema == entityops.SchemaDirectoryMembership.Name {
			membershipSetSeen = true
		}

		for _, envelope := range payloadSet.Envelopes {
			record, include, mapErr := mapIngestRecord(ctx, definition, payloadSet.Schema, envelope, installationFilterExpr)
			if mapErr != nil {
				logx.FromContext(ctx).Error().Err(mapErr).Str("schema", payloadSet.Schema).Str("resource", envelope.Resource).Msg("error mapping ingest record")

				attempted++
				failed++

				continue
			}

			if !include {
				continue
			}

			// resolve the effective cross-object link rules (installation override or definition
			// default) and inject the matched target ids into the create input, so the record is
			// created (or emitted for async creation) with its edges already set
			mapping, _ := findMapping(definition.Mappings, record.Schema, record.Variant)

			rules := resolveInstallationLinkRules(ic.Integration, definition, operationName)
			if len(rules) == 0 {
				rules = mapping.Links
			}

			record.Payload, err = injectLinks(ctx, ic.DB, ic.Integration.OwnerID, rules, record.Schema, record.Payload)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("schema", payloadSet.Schema).Str("resource", envelope.Resource).Msg("ingest link injection failed")

				attempted++
				failed++

				continue
			}

			attempted++

			if handleErr := handle(ctx, record); handleErr != nil {
				logx.FromContext(ctx).Error().Err(handleErr).Str("schema", payloadSet.Schema).Str("resource", envelope.Resource).Msg("ingest persist failed")

				failed++
			}
		}
	}

	if failed > 0 {
		return fmt.Errorf("%w: %d of %d records failed", ErrIngestRecordsFailed, failed, attempted)
	}

	// complete-snapshot directory syncs carry the full membership state, so any active membership
	// the run did not confirm was removed upstream; partial sources (scim, webhooks) and runs
	// finalized elsewhere never authorize removal inference
	if directorySync && membershipSetSeen && !options.SkipDirectorySyncRunFinalization && options.CompleteDirectorySnapshot {
		if err := markUnconfirmedDirectoryMembershipsRemoved(ctx, ic.DB, ic.Integration.ID, directorySyncRunID); err != nil {
			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
	}

	return nil
}

// markUnconfirmedDirectoryMembershipsRemoved records the removal side of the sync delta by
// stamping removed_at on active memberships for the integration that were not confirmed by
// the completed sync run
func markUnconfirmedDirectoryMembershipsRemoved(ctx context.Context, db *ent.Client, integrationID string, runID string) error {
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

// mapIngestRecord applies user-defined CEL expression filters over the data envelope so that we can filter out the data we import
func mapIngestRecord(ctx context.Context, definition types.Definition, schema string, envelope types.MappingEnvelope, installationFilterExpr string) (mappedIngestRecord, bool, error) {
	mapping, found := findMapping(definition.Mappings, schema, envelope.Variant)
	if !found {
		return mappedIngestRecord{}, false, ErrIngestMappingNotFound
	}

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

// installationLinkConfig holds the per-installation cross-object link overrides stored in the
// integration's client config; it mirrors the operation-scoped UserInput link rows
type installationLinkConfig struct {
	// Links overrides the definition's default link rules for the operation
	Links []types.LinkRule `json:"links,omitempty"`
}

// resolveInstallationLinkRules pulls the link-rule overrides for the current operation out of the
// installation config, mirroring resolveInstallationFilterExpr: the operation's ConfigResolver
// extracts the operation-specific section first, otherwise it falls back to a top-level links field.
// Returns nil when the installation supplies no override, so the caller uses the definition default
func resolveInstallationLinkRules(installation *ent.Integration, definition types.Definition, operationName string) []types.LinkRule {
	if operationName != "" {
		if op, ok := lo.Find(definition.Operations, func(o types.OperationRegistration) bool { return o.Name == operationName }); ok && op.ConfigResolver != nil {
			var cfg installationLinkConfig
			if err := jsonx.UnmarshalIfPresent(op.ConfigResolver(installation.Config.ClientConfig), &cfg); err == nil && len(cfg.Links) > 0 {
				return cfg.Links
			}
		}
	}

	var cfg installationLinkConfig
	if err := jsonx.UnmarshalIfPresent(installation.Config.ClientConfig, &cfg); err != nil {
		return nil
	}

	return cfg.Links
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
