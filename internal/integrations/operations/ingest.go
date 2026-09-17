package operations

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// ingestPreloadMinRecords is the minimum total envelope count across a batch's payload sets that
// triggers preloading integration activity state and link-target lookups for the batch
const ingestPreloadMinRecords = 2

// ingestMaxRecordAttempts caps the batched runs a tracked failing record is retried before it drops
// out of exclusion tracking, mirroring gala's defaultMaxErrorStreak
const ingestMaxRecordAttempts = 5

// keyJoin separates the schema name and lookup field values that make up an exclusion-tracking map
// key, and separates the lookup field values stored on a FailedRecord
const keyJoin = "\x1f"

// ingestQueryChunk bounds the number of lookup key tuples pushed into a single QueryByLookup batch
const ingestQueryChunk = 500

// fieldID is the primary key field name on every marshaled entity row
const fieldID = "id"

// IngestOptions carries the minimal ingest-time metadata needed by persistence
type IngestOptions struct {
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
	// Persisted counts records this installation's definition created, changed, or left unchanged
	Persisted int
	// Filtered counts records excluded by configured filters
	Filtered int
	// Succeeded is the combined successful handling count
	Succeeded int
	// Changed counts records that created, modified, or durably queued rows, excluding unchanged rows
	Changed int
	// Skipped counts records left untouched because another definition's identity manages the row
	// or because a newer integration run already wrote the row
	Skipped int
	// Failed counts records that could not be imported
	Failed int
	// Excluded counts records that failed again while tracked from an earlier run and were not
	// requeued
	Excluded int
	// Removed counts rows marked removed by snapshot reconciliation
	Removed int
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
	// schema is the entityops schema resolved once per payload set, carried on the record so the
	// persist handle never re-resolves it per record
	schema *entityops.Schema
}

// preparedIngestRecord is one payload set envelope after mapping and filtering, staged for link
// resolution and persistence
type preparedIngestRecord struct {
	// resource is the provider resource identifier, carried for failure reporting
	resource string
	// record is the mapped record ready for link resolution and persistence
	record mappedIngestRecord
	// links are the mapping variant's cross-object link rules, carried so a group of records sharing
	// a variant resolve them to entityops link specs once instead of once per record
	links []types.LinkRule
}

// logCtx returns ctx carrying the record's schema and resource logging fields
func (p preparedIngestRecord) logCtx(ctx context.Context) context.Context {
	return logx.WithFields(ctx, map[string]any{"schema": p.record.Schema, "resource": p.resource})
}

// ingestOutcome captures one persisted record's identity and the upsert decision entityops made for
// it: changed reports whether the write was material, and managed reports whether this
// installation's definition owns the resolved row
type ingestOutcome struct {
	id      string
	changed bool
	managed bool
}

// ingestHandle persists one mapped record and reports the upsert decision
type ingestHandle func(context.Context, mappedIngestRecord) (ingestOutcome, error)

// ingestBatch is one batch of payload sets and the declarations governing its ingest
type ingestBatch struct {
	// OperationName is the operation whose ingest produced the batch
	OperationName string
	// Contracts declare the schemas the operation may ingest
	Contracts []types.IngestContract
	// Policy is the operation's execution policy
	Policy types.ExecutionPolicy
	// PayloadSets are the mapped payload sets to persist
	PayloadSets []types.IngestPayloadSet
	// Options carries the run metadata for persistence
	Options IngestOptions
}

// variantGroup batches one payload set's prepared records sharing a mapping variant, so their link
// rules resolve to entityops link specs once per group instead of once per record
type variantGroup struct {
	links   []types.LinkRule
	records []preparedIngestRecord
}

// snapshotSet tracks one payload set's snapshot-reconciliation bookkeeping across its persist pass:
// scope is every row entityops considers live for this owner, definition, instance, and managing
// installation before the pass began, and seen accumulates the ids this pass actually confirmed
type snapshotSet struct {
	schema *entityops.Schema
	scope  map[string]struct{}
	seen   map[string]struct{}
	// stale counts rows a newer run already wrote: scope rows carrying a greater run id at load time
	// plus rows the run guard rejected during this pass; any stale row makes this run too old to
	// infer removals for the set
	stale int
}

// payloadRun carries one batch's shared state across its prepare, link, persist, and finalize passes
type payloadRun struct {
	ic           IngestContext
	batch        ingestBatch
	handle       ingestHandle
	definition   types.Definition
	filterExpr   string
	installation types.MappingInstallation
	preload      bool
	tracked      map[string]*models.FailedRecord
	dirty        bool
	snapshots    []*snapshotSet
	result       IngestResult
}

// ProcessPayloadSets persists one batch of mapped payload sets synchronously inside the run job;
// record failures are skipped, requeued as durable per-record jobs when ic.Runtime is set, and
// reported in the result, never the error
func ProcessPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, policy types.ExecutionPolicy, payloadSets []types.IngestPayloadSet, options IngestOptions) (IngestResult, error) {
	batch := ingestBatch{OperationName: operationName, Contracts: contracts, Policy: policy, PayloadSets: payloadSets, Options: options}

	return applyPayloadSets(ctx, ic, batch, func(handleCtx context.Context, record mappedIngestRecord) (ingestOutcome, error) {
		id, changed, managed, err := record.schema.PersistIngest(handleCtx, ic.DB, ic.Integration, record.Payload)

		return ingestOutcome{id: id, changed: changed, managed: managed}, err
	})
}

// applyPayloadSets maps, filters, links, and hands every record in the batch to handle, accumulating
// the record-level result and the installation's failed-record tracking state
func applyPayloadSets(ctx context.Context, ic IngestContext, batch ingestBatch, handle ingestHandle) (IngestResult, error) {
	run, ctx, err := newPayloadRun(ctx, ic, batch, handle)
	if err != nil {
		return IngestResult{}, err
	}

	for _, payloadSet := range batch.PayloadSets {
		if err := run.applySet(ctx, payloadSet); err != nil {
			return run.result, err
		}
	}

	if err := run.finalize(ctx); err != nil {
		return run.result, err
	}

	return run.result, nil
}

// newPayloadRun validates the batch against the installation and stages the shared run state
func newPayloadRun(ctx context.Context, ic IngestContext, batch ingestBatch, handle ingestHandle) (*payloadRun, context.Context, error) {
	definition, ok := ic.Registry.Definition(ic.Integration.DefinitionID)
	if !ok {
		return nil, ctx, ErrIngestDefinitionNotFound
	}

	if ic.Integration.InstallationMetadata.Display.ExternalID == "" {
		return nil, ctx, ErrIngestInstanceIDRequired
	}

	filterExpr, err := resolveInstallationFilterExpr(ic.Integration, definition, batch.OperationName)
	if err != nil {
		return nil, ctx, ErrIngestInstallationFilterConfigInvalid
	}

	run := &payloadRun{
		ic:         ic,
		batch:      batch,
		handle:     handle,
		definition: definition,
		filterExpr: filterExpr,
		installation: types.MappingInstallation{
			ID:               ic.Integration.ID,
			Name:             ic.Integration.Name,
			DefinitionID:     definition.ID,
			DefinitionName:   definition.DisplayName,
			InstanceID:       ic.Integration.InstallationMetadata.Display.ExternalID,
			PrimaryDirectory: ic.Integration.PrimaryDirectory,
		},
		preload: lo.SumBy(batch.PayloadSets, func(payloadSet types.IngestPayloadSet) int { return len(payloadSet.Envelopes) }) >= ingestPreloadMinRecords,
		tracked: make(map[string]*models.FailedRecord, len(ic.Integration.Health.FailedRecords)),
	}

	for i := range ic.Integration.Health.FailedRecords {
		fr := ic.Integration.Health.FailedRecords[i]
		run.tracked[trackingKey(fr.Schema, fr.Key)] = &fr
	}

	if !run.preload {
		return run, ctx, nil
	}

	ids, err := ic.DB.Integration.Query().Where(integration.OwnerIDEQ(ic.Integration.OwnerID)).IDs(ctx)
	if err != nil {
		return nil, ctx, fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
	}

	return run, entityops.WithActiveIntegrations(ctx, ids), nil
}

// fail records one record failure on the result
func (run *payloadRun) fail(schema, resource string, err error) {
	run.result.Failed++
	run.result.Failures = append(run.result.Failures, RecordFailure{Schema: schema, Resource: resource, Err: err})
}

// applySet prepares, links, and persists one payload set
func (run *payloadRun) applySet(ctx context.Context, payloadSet types.IngestPayloadSet) error {
	sourceSchema, err := run.resolveSetSchema(payloadSet.Schema)
	if err != nil {
		return err
	}

	prepared := run.prepare(ctx, payloadSet, sourceSchema)

	ready, err := run.link(ctx, sourceSchema, prepared)
	if err != nil {
		return err
	}

	setCtx := ctx

	if run.preload && sourceSchema.QueryByLookup != nil && len(sourceSchema.Lookup) > 0 {
		prefetched, err := prefetchLookupMatches(ctx, run.ic.DB, run.ic.Integration.OwnerID, sourceSchema, ready)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}

		setCtx = prefetched
	}

	set, err := run.snapshot(ctx, payloadSet, sourceSchema)
	if err != nil {
		return err
	}

	for _, p := range ready {
		if err := run.persist(setCtx, p, set); err != nil {
			return err
		}
	}

	return nil
}

// resolveSetSchema checks the payload set's schema is declared by the operation and ingest-capable
func (run *payloadRun) resolveSetSchema(name string) (*entityops.Schema, error) {
	if !contractIncludesSchema(run.batch.Contracts, name) {
		return nil, ErrIngestSchemaNotDeclared
	}

	sourceSchema, ok := lookupIngestSchema(name)
	if ok {
		return sourceSchema, nil
	}

	if _, exists := entityops.LookupSchema(name); exists {
		return nil, ErrIngestUnsupportedSchema
	}

	return nil, ErrIngestSchemaNotFound
}

// prepare maps and filters the payload set's envelopes, stamping provenance on every included record
func (run *payloadRun) prepare(ctx context.Context, payloadSet types.IngestPayloadSet, sourceSchema *entityops.Schema) []preparedIngestRecord {
	prepared := make([]preparedIngestRecord, 0, len(payloadSet.Envelopes))

	for _, envelope := range payloadSet.Envelopes {
		run.result.Attempted++
		envCtx := logx.WithFields(ctx, map[string]any{"schema": payloadSet.Schema, "resource": envelope.Resource})

		mapping, found := findMapping(run.definition.Mappings, payloadSet.Schema, envelope.Variant)
		if !found {
			logx.FromContext(envCtx).Error().Err(ErrIngestMappingNotFound).Msg("error mapping ingest record")
			run.fail(payloadSet.Schema, envelope.Resource, ErrIngestMappingNotFound)

			continue
		}

		record, include, err := mapIngestRecord(envCtx, mapping, payloadSet.Schema, envelope, run.filterExpr, run.installation)
		if err != nil {
			logx.FromContext(envCtx).Error().Err(err).Msg("error mapping ingest record")
			run.fail(payloadSet.Schema, envelope.Resource, err)

			continue
		}

		if !include {
			run.result.Filtered++
			continue
		}

		record.schema = sourceSchema
		record.Payload = entityops.StampProvenance(record.Payload, sourceSchema, run.ic.Integration, run.batch.Options.RunID)

		prepared = append(prepared, preparedIngestRecord{resource: envelope.Resource, record: record, links: mapping.Links})
	}

	return prepared
}

// link resolves each variant group's link rules once and injects the link targets into its records
func (run *payloadRun) link(ctx context.Context, sourceSchema *entityops.Schema, prepared []preparedIngestRecord) ([]preparedIngestRecord, error) {
	ready := make([]preparedIngestRecord, 0, len(prepared))

	for _, group := range groupByVariant(prepared) {
		specs, err := linkSpecs(sourceSchema, group.links)
		if err != nil {
			for _, p := range group.records {
				logx.FromContext(p.logCtx(ctx)).Error().Err(err).Msg("ingest link injection failed")
				run.fail(p.record.Schema, p.resource, err)
			}

			continue
		}

		linkCtx, err := run.prefetchLinkTargets(ctx, sourceSchema, group.records, specs)
		if err != nil {
			return ready, err
		}

		for _, p := range group.records {
			payload, err := entityops.InjectCreateLinks(linkCtx, run.ic.DB, run.ic.Integration.OwnerID, sourceSchema, p.record.Payload, specs)
			if err != nil {
				logx.FromContext(p.logCtx(ctx)).Error().Err(err).Str("schema", sourceSchema.Name).Msg("ingest link target resolution failed")
				run.fail(p.record.Schema, p.resource, fmt.Errorf("%w: %w", ErrLinkFailed, err))

				continue
			}

			p.record.Payload = payload
			ready = append(ready, p)
		}
	}

	return ready, nil
}

// prefetchLinkTargets caches the group's link targets on ctx when the batch is large enough to preload
func (run *payloadRun) prefetchLinkTargets(ctx context.Context, sourceSchema *entityops.Schema, records []preparedIngestRecord, specs []entityops.LinkSpec) (context.Context, error) {
	if !run.preload || len(specs) == 0 {
		return ctx, nil
	}

	payloads := lo.Map(records, func(p preparedIngestRecord, _ int) json.RawMessage { return p.record.Payload })

	prefetched, err := entityops.PrefetchLinkTargets(ctx, run.ic.DB, run.ic.Integration.OwnerID, sourceSchema, payloads, specs)
	if err != nil {
		return ctx, fmt.Errorf("%w: %w", ErrLinkFailed, err)
	}

	return prefetched, nil
}

// snapshot loads the payload set's snapshot scope when the run may infer removals for it
func (run *payloadRun) snapshot(ctx context.Context, payloadSet types.IngestPayloadSet, sourceSchema *entityops.Schema) (*snapshotSet, error) {
	if !run.batch.Policy.Snapshot || !payloadSet.SnapshotComplete || sourceSchema.SnapshotScope == nil {
		return nil, nil
	}

	rows, err := sourceSchema.SnapshotScope(ctx, run.ic.DB, run.ic.Integration.OwnerID, run.ic.Integration.DefinitionID, run.ic.Integration.InstallationMetadata.Display.ExternalID, run.ic.Integration.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
	}

	scope := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		scope[entityops.FieldValue(row, fieldID)] = struct{}{}
	}

	set := &snapshotSet{schema: sourceSchema, scope: scope, seen: map[string]struct{}{}}

	if runID := run.batch.Options.RunID; runID != "" {
		set.stale = lo.CountBy(rows, func(row json.RawMessage) bool {
			return entityops.FieldValue(row, entityops.FieldIntegrationRunID) > runID
		})
	}

	run.snapshots = append(run.snapshots, set)

	return set, nil
}

// persist hands one ready record to the handle unless its tracked failure excludes it, recording the outcome
func (run *payloadRun) persist(ctx context.Context, p preparedIngestRecord, set *snapshotSet) error {
	var recordKey, trackKey string

	if key, ok := failedRecordKey(p.record.schema, p.record.Payload); ok {
		recordKey = key
		trackKey = trackingKey(p.record.schema.Snake, key)
	}

	recordCtx := p.logCtx(ctx)

	if entry, excluded := run.tracked[trackKey]; excluded && trackKey != "" {
		skip, err := run.skipExcluded(ctx, recordCtx, p, set, trackKey, entry)
		if err != nil || skip {
			return err
		}
	}

	outcome, err := run.handle(recordCtx, p.record)

	switch {
	case err == nil:
		run.recordSuccess(outcome, set, trackKey)
	case errors.Is(err, entityops.ErrUpsertStaleRun):
		run.result.Skipped++

		if set != nil {
			set.stale++
		}

		logx.FromContext(recordCtx).Debug().Msg("ingest skipped stale run")
	default:
		run.recordFailure(recordCtx, p, recordKey, trackKey, err)
	}

	return nil
}

// skipExcluded applies the record's tracked failure, reporting whether the record is skipped this run
func (run *payloadRun) skipExcluded(ctx, recordCtx context.Context, p preparedIngestRecord, set *snapshotSet, trackKey string, entry *models.FailedRecord) (bool, error) {
	schema := p.record.schema
	resolvable := schema.QueryByLookup != nil && entry.RunID != ""

	var rows []json.RawMessage

	if set != nil || resolvable {
		fetched, err := excludedRecordRows(ctx, run.ic.DB, run.ic.Integration.OwnerID, schema, p.record.Payload)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}

		rows = fetched
	}

	run.dirty = true

	if resolvable && trackedFailureResolved(rows, entry.RunID) {
		delete(run.tracked, trackKey)

		logx.FromContext(recordCtx).Debug().Msg("ingest tracked failure resolved by durable retry")

		return false, nil
	}

	entry.Attempts++

	if set != nil {
		for _, row := range rows {
			set.seen[entityops.FieldValue(row, fieldID)] = struct{}{}
		}
	}

	run.result.Excluded++
	run.result.Failures = append(run.result.Failures, RecordFailure{Schema: p.record.Schema, Resource: p.resource, Err: fmt.Errorf("%w: %s", ErrIngestRecordExcluded, entry.LastError)})

	logx.FromContext(recordCtx).Warn().Int("attempts", entry.Attempts).Msg("ingest excluded record skipped")

	if entry.Attempts >= ingestMaxRecordAttempts {
		delete(run.tracked, trackKey)
	}

	return true, nil
}

// recordSuccess counts a handled record and clears any tracked failure it resolves
func (run *payloadRun) recordSuccess(outcome ingestOutcome, set *snapshotSet, trackKey string) {
	run.result.Succeeded++

	if outcome.managed {
		run.result.Persisted++
	} else {
		run.result.Skipped++
	}

	if outcome.changed {
		run.result.Changed++
	}

	if _, excluded := run.tracked[trackKey]; excluded && trackKey != "" {
		delete(run.tracked, trackKey)
		run.dirty = true
	}

	if set != nil {
		set.seen[outcome.id] = struct{}{}
	}
}

// recordFailure tracks a persist failure, requeues the record durably when a runtime is available, and counts it
func (run *payloadRun) recordFailure(recordCtx context.Context, p preparedIngestRecord, recordKey, trackKey string, err error) {
	wrapped := wrapIngestPersistError(err)

	if trackKey != "" {
		run.tracked[trackKey] = &models.FailedRecord{Schema: p.record.schema.Snake, Key: recordKey, RunID: run.batch.Options.RunID, Attempts: 1, LastError: wrapped.Error()}
		run.dirty = true
	}

	requeued := false

	if run.ic.Runtime != nil {
		if requeueErr := emitMappedRecord(recordCtx, run.ic.Runtime, run.ic.Integration, run.batch.OperationName, p.record, run.batch.Options); requeueErr != nil {
			logx.FromContext(recordCtx).Warn().Err(requeueErr).Msg("ingest failure requeue failed")
		} else {
			requeued = true
		}
	}

	logx.FromContext(recordCtx).Error().Err(wrapped).Bool("requeued", requeued).Msg("ingest persist failed")

	run.fail(p.record.Schema, p.resource, wrapped)
}

// finalize applies snapshot removals when every record imported and persists the failed-record tracking state
func (run *payloadRun) finalize(ctx context.Context) error {
	switch {
	case run.result.Failed > 0:
		logx.FromContext(ctx).Warn().Int("failed", run.result.Failed).Int("attempted", run.result.Attempted).Msg("ingest skipped records that could not be imported")
	default:
		if err := run.removeUnseen(ctx); err != nil {
			return err
		}
	}

	if run.result.Skipped > 0 {
		logx.FromContext(ctx).Info().Int("skipped", run.result.Skipped).Msg("ingest left records unmanaged by this installation")
	}

	if !run.dirty {
		return nil
	}

	if err := persistFailedRecords(ctx, run.ic, failedRecordsFromTracked(run.tracked)); err != nil {
		return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
	}

	return nil
}

// removeUnseen marks every snapshot-scoped row this run did not confirm as removed
func (run *payloadRun) removeUnseen(ctx context.Context) error {
	for _, set := range run.snapshots {
		if set.stale > 0 {
			logx.FromContext(ctx).Debug().Str("schema", set.schema.Snake).Int("stale", set.stale).Msg("ingest skipped snapshot removal for stale run")
			continue
		}

		removed := lo.FilterKeys(set.scope, func(id string, _ struct{}) bool {
			_, seen := set.seen[id]
			return !seen
		})

		if len(removed) == 0 {
			continue
		}

		if err := set.schema.MarkRemoved(ctx, run.ic.DB, removed, time.Now(), run.batch.Options.RunID); err != nil {
			return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
		}

		run.result.Removed += len(removed)
	}

	return nil
}

// persistFailedRecords writes the exclusion-tracking state onto the installation's health from a fresh read of the row, so concurrent health writers are not clobbered by the batch's stale snapshot
func persistFailedRecords(ctx context.Context, ic IngestContext, records []models.FailedRecord) error {
	systemCtx := privacy.DecisionContext(ctx, privacy.Allow)

	current, err := ic.DB.Integration.Get(systemCtx, ic.Integration.ID)
	if err != nil {
		return err
	}

	health := current.Health
	health.FailedRecords = records

	if err := ic.DB.Integration.UpdateOneID(ic.Integration.ID).SetHealth(health).Exec(systemCtx); err != nil {
		return err
	}

	ic.Integration.Health = health

	return nil
}

// groupByVariant batches a payload set's prepared records by mapping variant, in first-seen order,
// so their shared link rules resolve to entityops link specs once per group instead of once per record
func groupByVariant(prepared []preparedIngestRecord) []variantGroup {
	var groups []variantGroup

	index := make(map[string]int, len(prepared))

	for _, p := range prepared {
		i, ok := index[p.record.Variant]
		if !ok {
			i = len(groups)
			index[p.record.Variant] = i
			groups = append(groups, variantGroup{links: p.links})
		}

		groups[i].records = append(groups[i].records, p)
	}

	return groups
}

// lookupKeyFor resolves the first lookup alternative whose fields are all present and non-empty in
// the payload, returning its index and extracted values
func lookupKeyFor(schema *entityops.Schema, payload json.RawMessage) (alternative int, values entityops.LookupValues, ok bool) {
	for i, alt := range schema.Lookup {
		candidate := make(entityops.LookupValues, len(alt.Fields))
		complete := true

		for _, name := range alt.Fields {
			field, found := schema.FieldByName(name)
			if !found {
				complete = false
				break
			}

			value := entityops.FieldValue(payload, field.InputKey)
			if value == "" {
				complete = false
				break
			}

			candidate[name] = value
		}

		if complete {
			return i, candidate, true
		}
	}

	return 0, nil, false
}

// failedRecordKey renders a mapped record's exclusion-tracking key from its first complete lookup
// alternative, joining the alternative's values in declared field order by keyJoin. ok is false when
// no alternative is complete, meaning the record cannot be tracked
func failedRecordKey(schema *entityops.Schema, payload json.RawMessage) (key string, ok bool) {
	alternative, values, ok := lookupKeyFor(schema, payload)
	if !ok {
		return "", false
	}

	fields := schema.Lookup[alternative].Fields
	ordered := make([]string, len(fields))

	for i, name := range fields {
		ordered[i] = values[name]
	}

	return strings.Join(ordered, keyJoin), true
}

// recordLookup pairs one ready record's resolved ingest lookup alternative and key values, staged
// for grouping into per-alternative prefetch batches
type recordLookup struct {
	alternative int
	values      entityops.LookupValues
}

// prefetchLookupMatches batch-queries a payload set's ready records against their ingest lookup
// alternatives and installs the results as a ctx-carried match cache (see
// entityops.WithLookupMatches), letting Upsert resolve each record's existing row from the
// prefetched batch instead of issuing QueryByLookup per record
func prefetchLookupMatches(ctx context.Context, db *ent.Client, ownerID string, schema *entityops.Schema, ready []preparedIngestRecord) (context.Context, error) {
	lookups := lo.FilterMap(ready, func(p preparedIngestRecord, _ int) (recordLookup, bool) {
		alternative, values, ok := lookupKeyFor(schema, p.record.Payload)

		return recordLookup{alternative: alternative, values: values}, ok
	})

	matches := make(map[int]map[string][]json.RawMessage, len(schema.Lookup))

	for alternative, group := range lo.GroupBy(lookups, func(l recordLookup) int { return l.alternative }) {
		alt := schema.Lookup[alternative]
		values := lo.Map(group, func(l recordLookup, _ int) entityops.LookupValues { return l.values })

		keyed := make(map[string][]json.RawMessage, len(values))
		for _, v := range values {
			keyed[entityops.EncodeLookupKey(alt, v)] = nil
		}

		for _, chunk := range lo.Chunk(values, ingestQueryChunk) {
			rows, err := schema.QueryByLookup(ctx, db, ownerID, alternative, chunk)
			if err != nil {
				return ctx, err
			}

			keyLookupRows(alt, rows, keyed)
		}

		matches[alternative] = keyed
	}

	return entityops.WithLookupMatches(ctx, schema, matches), nil
}

// keyLookupRows indexes rows by their lookup alternative key
func keyLookupRows(alt entityops.LookupAlternative, rows []json.RawMessage, keyed map[string][]json.RawMessage) {
	for _, row := range rows {
		rowValues := make(entityops.LookupValues, len(alt.Fields))
		for _, name := range alt.Fields {
			rowValues[name] = entityops.FieldValue(row, name)
		}

		key := entityops.EncodeLookupKey(alt, rowValues)
		keyed[key] = append(keyed[key], row)
	}
}

// excludedRecordRows resolves the rows an excluded record's lookup key currently matches, so a
// snapshot pass can mark them seen (never marking a record removed only because its write was
// skipped) and a tracked failure can be checked for resolution by a durable retry
func excludedRecordRows(ctx context.Context, db *ent.Client, ownerID string, schema *entityops.Schema, payload json.RawMessage) ([]json.RawMessage, error) {
	alternative, values, ok := lookupKeyFor(schema, payload)
	if !ok || schema.QueryByLookup == nil {
		return nil, nil
	}

	return schema.QueryByLookup(ctx, db, ownerID, alternative, []entityops.LookupValues{values})
}

// trackedFailureResolved reports whether any row matching an excluded record's lookup key carries
// an integration run id at or after the tracked failure's run id, meaning a durable per-record
// retry already wrote the row after the batched run that recorded the failure
func trackedFailureResolved(rows []json.RawMessage, runID string) bool {
	return lo.ContainsBy(rows, func(row json.RawMessage) bool {
		return entityops.FieldValue(row, entityops.FieldIntegrationRunID) >= runID
	})
}

// trackingKey builds the exclusion-tracking map key from a schema name and a failed record's
// lookup key, so records sharing lookup values across different schemas track independently
func trackingKey(schemaName, key string) string {
	return schemaName + keyJoin + key
}

// failedRecordsFromTracked renders the exclusion-tracking map into a deterministically ordered
// slice, sorted by schema then key, for persistence on the integration's health
func failedRecordsFromTracked(tracked map[string]*models.FailedRecord) []models.FailedRecord {
	records := make([]models.FailedRecord, 0, len(tracked))

	for _, fr := range tracked {
		records = append(records, *fr)
	}

	slices.SortFunc(records, func(a, b models.FailedRecord) int {
		return cmp.Or(strings.Compare(a.Schema, b.Schema), strings.Compare(a.Key, b.Key))
	})

	return records
}

// mapIngestRecord applies the resolved mapping's filters and map expression to one data envelope,
// returning the mapped record and whether the envelope passed the include filters
func mapIngestRecord(ctx context.Context, mapping types.MappingOverride, schema string, envelope types.MappingEnvelope, installationFilterExpr string, installation types.MappingInstallation) (mappedIngestRecord, bool, error) {
	matched, err := envelopeIncludedByFilters(ctx, installationFilterExpr, mapping.FilterExpr, envelope, installation)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("ingest filter failed")
		return mappedIngestRecord{}, false, ErrIngestFilterFailed
	}
	if !matched {
		return mappedIngestRecord{}, false, nil
	}

	mapped, err := providerkit.EvalMap(ctx, mapping.MapExpr, envelope, installation)
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
func envelopeIncludedByFilters(ctx context.Context, installationFilterExpr string, mappingFilterExpr string, envelope types.MappingEnvelope, installation types.MappingInstallation) (bool, error) {
	matched, err := providerkit.EvalFilter(ctx, installationFilterExpr, envelope, installation)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, nil
	}

	return providerkit.EvalFilter(ctx, mappingFilterExpr, envelope, installation)
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

// RecordFailureSummary renders a compact description of a run's failed records for the run's error text
func RecordFailureSummary(result IngestResult) string {
	first := result.Failures[0]

	return fmt.Sprintf("%d of %d records failed to import; first failure: %s %s: %v", result.Failed, result.Attempted, first.Schema, first.Resource, first.Err)
}

// wrapIngestPersistError wraps the known errors from persistence operations so we don't need the same boilerplate in multiple functions
func wrapIngestPersistError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case ent.IsValidationError(err):
		return fmt.Errorf("%w: %w", ErrIngestMappedDocumentInvalid, err)
	case ent.IsNotSingular(err), ent.IsConstraintError(err), errors.Is(err, entityops.ErrUpsertConflict):
		return fmt.Errorf("%w: %w", ErrIngestUpsertConflict, err)
	case errors.Is(err, entityops.ErrUpsertKeyMissing):
		return fmt.Errorf("%w: %w", ErrIngestUpsertKeyMissing, err)
	default:
		return fmt.Errorf("%w: %w", ErrIngestPersistFailed, err)
	}
}
