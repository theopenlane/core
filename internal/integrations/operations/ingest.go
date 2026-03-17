package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// IngestOptions carries the minimal ingest-time metadata needed by persistence
type IngestOptions struct {
	// DirectorySyncRunID is the database ID of the DirectorySyncRun record that groups all
	// directory-related ingest records (accounts, groups, memberships) from one sync batch;
	// created at the start of processing and finalized when the batch completes
	DirectorySyncRunID string
	// SkipDirectorySyncRunFinalization instructs the processor not to finalize the directory
	// sync run after processing; set when the caller is responsible for finalization (e.g. async path)
	SkipDirectorySyncRunFinalization bool
	// Source identifies the mechanism that produced the ingest data (e.g. webhook, poll, manual)
	Source integrationgenerated.IntegrationIngestSource
	// RunID is a caller-supplied correlation identifier for the overall operation run (e.g. a polling
	// job invocation); propagated into Gala message headers for distributed tracing — not a DB entity
	RunID string
	// Webhook is the webhook name or identifier that triggered this ingest, if applicable
	Webhook string
	// WebhookEvent is the event type reported by the webhook provider (e.g. "push", "member_added")
	WebhookEvent string
	// DeliveryID is the provider-assigned delivery identifier for webhook payloads, used for deduplication
	DeliveryID string
	// WorkflowMeta carries workflow instance context when ingest is triggered from a workflow action
	WorkflowMeta *WorkflowMeta
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

// ProcessIngestAsync decodes one operation response into payload sets and emits them via Gala
func ProcessIngestAsync(ctx context.Context, ic IngestContext, operation types.OperationRegistration, response json.RawMessage, options IngestOptions) error {
	var payloadSets []types.IngestPayloadSet
	if err := json.Unmarshal(response, &payloadSets); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestPayloadsInvalid, err)
	}

	return EmitPayloadSets(ctx, ic, operation.Name, operation.Ingest, payloadSets, options)
}

// EmitPayloadSets transforms one batch of mapped payload sets and emits typed second-stage Gala ingest requests
func EmitPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) error {
	if ic.Runtime == nil {
		return ErrGalaRequired
	}

	// Second-stage ingest is asynchronous; do not finalize directory sync runs before listeners finish.
	options.SkipDirectorySyncRunFinalization = true

	return processPayloadSets(ctx, ic, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		return emitMappedRecord(handleCtx, ic.Runtime, ic.Installation, operationName, record, options)
	})
}

// ProcessPayloadSets persists one batch of mapped payload sets synchronously
func ProcessPayloadSets(ctx context.Context, ic IngestContext, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions) error {
	return processPayloadSets(ctx, ic, contracts, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) error {
		return persistMappedRecord(handleCtx, ic.DB, ic.Installation, record.Schema, record.Payload)
	})
}

// processPayloadSets is the shared core for both async emit and sync persist paths
func processPayloadSets(ctx context.Context, ic IngestContext, contracts []types.IngestContract, payloadSets []types.IngestPayloadSet, options IngestOptions, handle func(context.Context, mappedIngestRecord) error) (err error) {
	definition, ok := ic.Registry.Definition(ic.Installation.DefinitionID)
	if !ok {
		return ErrIngestDefinitionNotFound
	}

	installationFilterExpr, err := resolveInstallationFilterExpr(ic.Installation)
	if err != nil {
		return ErrIngestInstallationFilterConfigInvalid
	}

	directorySyncRunID := options.DirectorySyncRunID
	if directorySyncRunID == "" && needsDirectorySyncRun(contracts) {
		directorySyncRunID, err = createDirectorySyncRun(ctx, ic.DB, ic.Installation)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}
	}

	if directorySyncRunID != "" {
		ctx = withDirectorySyncRunID(ctx, directorySyncRunID)
	}

	shouldFinalizeDirectorySyncRun := directorySyncRunID != "" && needsDirectorySyncRun(contracts) && !options.SkipDirectorySyncRunFinalization
	if shouldFinalizeDirectorySyncRun {
		defer func() {
			if finalizeErr := finalizeDirectorySyncRun(ctx, ic.DB, directorySyncRunID, err); finalizeErr != nil && err == nil {
				err = finalizeErr
			}
		}()
	}

	for _, payloadSet := range payloadSets {
		if !contractIncludesSchema(contracts, payloadSet.Schema) {
			return ErrIngestSchemaNotFound
		}
		if _, ok := integrationgenerated.IntegrationMappingSchemas[payloadSet.Schema]; !ok {
			return ErrIngestSchemaNotFound
		}

		for _, envelope := range payloadSet.Envelopes {
			record, include, recordErr := mapIngestRecord(ctx, definition, payloadSet.Schema, envelope, installationFilterExpr)
			if recordErr != nil {
				return recordErr
			}
			if !include {
				continue
			}

			if err := handle(ctx, record); err != nil {
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
		return mappedIngestRecord{}, false, ErrIngestFilterFailed
	}
	if !matched {
		return mappedIngestRecord{}, false, nil
	}

	mapped, err := providerkit.EvalMap(ctx, mapping.MapExpr, envelope)
	if err != nil {
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
	if installation == nil {
		return "", nil
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
	for _, mapping := range mappings {
		if mapping.Schema == schema && mapping.Variant == variant {
			return mapping.Spec, true
		}
	}

	return types.MappingOverride{}, false
}

// contractIncludesSchema checks whether the given list of contracts includes a contract for the given schema
func contractIncludesSchema(contracts []types.IngestContract, schema string) bool {
	for _, contract := range contracts {
		if contract.Schema == schema {
			return true
		}
	}

	return false
}

// needsDirectorySyncRun checks whether any of the given contracts require a directory sync run to be created
func needsDirectorySyncRun(contracts []types.IngestContract) bool {
	for _, contract := range contracts {
		switch contract.Schema {
		case integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			integrationgenerated.IntegrationMappingSchemaDirectoryMembership:
			return true
		}
	}

	return false
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

// finalizeDirectorySyncRun marks the directory sync run as completed or failed
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
