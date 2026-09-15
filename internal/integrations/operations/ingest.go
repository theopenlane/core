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
	// ctx carries the envelope's logging fields into the link and persist passes
	ctx context.Context
	// resource is the provider resource identifier, carried for failure reporting
	resource string
	// record is the mapped record ready for link resolution and persistence
	record mappedIngestRecord
	// links are the mapping variant's cross-object link rules, carried so a group of records sharing
	// a variant resolve them to entityops link specs once instead of once per record
	links []types.LinkRule
}

// ingestOutcome captures one persisted record's identity and the upsert decision entityops made for
// it: changed reports whether the write was material, and managed reports whether this
// installation's definition owns the resolved row
type ingestOutcome struct {
	id      string
	changed bool
	managed bool
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

// ProcessPayloadSets persists one batch of mapped payload sets synchronously inside the run job;
// record failures are skipped, requeued as durable per-record jobs when ic.Runtime is set, and
// reported in the result, never the error
func ProcessPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, policy types.ExecutionPolicy, payloadSets []types.IngestPayloadSet, options IngestOptions) (IngestResult, error) {
	return applyPayloadSets(ctx, ic, operationName, contracts, policy, payloadSets, options, func(handleCtx context.Context, record mappedIngestRecord) (ingestOutcome, error) {
		id, changed, managed, err := record.schema.PersistIngest(handleCtx, ic.DB, ic.Integration, record.Payload)

		return ingestOutcome{id: id, changed: changed, managed: managed}, err
	})
}

// applyPayloadSets maps, filters, links, and hands every record in the batch to handle, accumulating
// the record-level result and the installation's failed-record tracking state
func applyPayloadSets(ctx context.Context, ic IngestContext, operationName string, contracts []types.IngestContract, policy types.ExecutionPolicy, payloadSets []types.IngestPayloadSet, options IngestOptions, handle func(context.Context, mappedIngestRecord) (ingestOutcome, error)) (result IngestResult, err error) {
	definition, ok := ic.Registry.Definition(ic.Integration.DefinitionID)
	if !ok {
		return result, ErrIngestDefinitionNotFound
	}

	if ic.Integration.InstallationMetadata.Display.ExternalID == "" {
		return result, ErrIngestInstanceIDRequired
	}

	installationFilterExpr, err := resolveInstallationFilterExpr(ic.Integration, definition, operationName)
	if err != nil {
		return result, ErrIngestInstallationFilterConfigInvalid
	}

	mappingInstallation := types.MappingInstallation{
		ID:               ic.Integration.ID,
		Name:             ic.Integration.Name,
		DefinitionID:     definition.ID,
		DefinitionName:   definition.DisplayName,
		InstanceID:       ic.Integration.InstallationMetadata.Display.ExternalID,
		PrimaryDirectory: ic.Integration.PrimaryDirectory,
	}

	preload := lo.SumBy(payloadSets, func(payloadSet types.IngestPayloadSet) int { return len(payloadSet.Envelopes) }) >= ingestPreloadMinRecords

	tracked := make(map[string]*models.FailedRecord)
	dirty := false

	for i := range ic.Integration.Health.FailedRecords {
		fr := ic.Integration.Health.FailedRecords[i]
		tracked[trackingKey(fr.Schema, fr.Key)] = &fr
	}

	if preload {
		ids, activeErr := ic.DB.Integration.Query().Where(integration.OwnerIDEQ(ic.Integration.OwnerID)).IDs(ctx)
		if activeErr != nil {
			return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, activeErr)
		}

		ctx = entityops.WithActiveIntegrations(ctx, ids)
	}

	var snapshotSets []*snapshotSet

	for _, payloadSet := range payloadSets {
		if !contractIncludesSchema(contracts, payloadSet.Schema) {
			return result, ErrIngestSchemaNotDeclared
		}

		sourceSchema, ok := lookupIngestSchema(payloadSet.Schema)
		if !ok {
			if _, exists := entityops.LookupSchema(payloadSet.Schema); exists {
				return result, ErrIngestUnsupportedSchema
			}

			return result, ErrIngestSchemaNotFound
		}

		prepared := make([]preparedIngestRecord, 0, len(payloadSet.Envelopes))

		for _, envelope := range payloadSet.Envelopes {
			result.Attempted++
			envCtx := logx.WithFields(ctx, map[string]any{"schema": payloadSet.Schema, "resource": envelope.Resource})

			mapping, found := findMapping(definition.Mappings, payloadSet.Schema, envelope.Variant)
			if !found {
				logx.FromContext(envCtx).Error().Err(ErrIngestMappingNotFound).Msg("error mapping ingest record")

				result.Failed++
				result.Failures = append(result.Failures, RecordFailure{Schema: payloadSet.Schema, Resource: envelope.Resource, Err: ErrIngestMappingNotFound})

				continue
			}

			record, include, mapErr := mapIngestRecord(envCtx, mapping, payloadSet.Schema, envelope, installationFilterExpr, mappingInstallation)
			if mapErr != nil {
				logx.FromContext(envCtx).Error().Err(mapErr).Msg("error mapping ingest record")

				result.Failed++
				result.Failures = append(result.Failures, RecordFailure{Schema: payloadSet.Schema, Resource: envelope.Resource, Err: mapErr})

				continue
			}

			if !include {
				result.Filtered++
				continue
			}

			record.schema = sourceSchema
			record.Payload = entityops.StampProvenance(record.Payload, sourceSchema, ic.Integration, options.RunID)

			prepared = append(prepared, preparedIngestRecord{ctx: envCtx, resource: envelope.Resource, record: record, links: mapping.Links})
		}

		ready := make([]preparedIngestRecord, 0, len(prepared))

		for _, group := range groupByVariant(prepared) {
			specs, specsErr := linkSpecs(sourceSchema, group.links)
			if specsErr != nil {
				for _, p := range group.records {
					logx.FromContext(p.ctx).Error().Err(specsErr).Msg("ingest link injection failed")

					result.Failed++
					result.Failures = append(result.Failures, RecordFailure{Schema: payloadSet.Schema, Resource: p.resource, Err: specsErr})
				}

				continue
			}

			linkCtx := ctx

			if preload && len(specs) > 0 {
				payloads := lo.Map(group.records, func(p preparedIngestRecord, _ int) json.RawMessage { return p.record.Payload })

				prefetchCtx, prefetchErr := entityops.PrefetchLinkTargets(ctx, ic.DB, ic.Integration.OwnerID, sourceSchema, payloads, specs)
				if prefetchErr != nil {
					return result, fmt.Errorf("%w: %w", ErrLinkFailed, prefetchErr)
				}

				linkCtx = prefetchCtx
			}

			for _, p := range group.records {
				payload, linkErr := entityops.InjectCreateLinks(linkCtx, ic.DB, ic.Integration.OwnerID, sourceSchema, p.record.Payload, specs)
				if linkErr != nil {
					logx.FromContext(p.ctx).Error().Err(linkErr).Str("schema", sourceSchema.Name).Msg("ingest link target resolution failed")

					result.Failed++
					result.Failures = append(result.Failures, RecordFailure{Schema: payloadSet.Schema, Resource: p.resource, Err: fmt.Errorf("%w: %w", ErrLinkFailed, linkErr)})

					continue
				}

				p.record.Payload = payload
				ready = append(ready, p)
			}
		}

		if preload && sourceSchema.QueryByLookup != nil && len(sourceSchema.Lookup) > 0 {
			setCtx, prefetchErr := prefetchLookupMatches(ctx, ic.DB, ic.Integration.OwnerID, sourceSchema, ready)
			if prefetchErr != nil {
				return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, prefetchErr)
			}

			for i := range ready {
				ready[i].ctx = logx.WithFields(setCtx, map[string]any{"schema": payloadSet.Schema, "resource": ready[i].resource})
			}
		}

		var set *snapshotSet

		if policy.Snapshot && payloadSet.SnapshotComplete && sourceSchema.SnapshotScope != nil {
			rows, scopeErr := sourceSchema.SnapshotScope(ctx, ic.DB, ic.Integration.OwnerID, ic.Integration.DefinitionID, ic.Integration.InstallationMetadata.Display.ExternalID, ic.Integration.ID)
			if scopeErr != nil {
				return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, scopeErr)
			}

			scope := make(map[string]struct{}, len(rows))
			for _, row := range rows {
				scope[entityops.FieldValue(row, fieldID)] = struct{}{}
			}

			set = &snapshotSet{schema: sourceSchema, scope: scope, seen: map[string]struct{}{}}

			if options.RunID != "" {
				set.stale = lo.CountBy(rows, func(row json.RawMessage) bool {
					return entityops.FieldValue(row, entityops.FieldIntegrationRunID) > options.RunID
				})
			}

			snapshotSets = append(snapshotSets, set)
		}

		for _, p := range ready {
			var recordKey, trackKey string

			if key, ok := failedRecordKey(sourceSchema, p.record.Payload); ok {
				recordKey = key
				trackKey = trackingKey(sourceSchema.Snake, key)
			}

			if entry, excluded := tracked[trackKey]; excluded && trackKey != "" {
				resolvable := sourceSchema.QueryByLookup != nil && entry.RunID != ""

				var rows []json.RawMessage

				if set != nil || resolvable {
					fetched, lookupErr := excludedRecordRows(ctx, ic.DB, ic.Integration.OwnerID, sourceSchema, p.record.Payload)
					if lookupErr != nil {
						return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, lookupErr)
					}

					rows = fetched
				}

				if resolvable && trackedFailureResolved(rows, entry.RunID) {
					delete(tracked, trackKey)
					dirty = true

					logx.FromContext(p.ctx).Debug().Msg("ingest tracked failure resolved by durable retry")
				} else {
					entry.Attempts++
					dirty = true

					if set != nil {
						for _, row := range rows {
							set.seen[entityops.FieldValue(row, fieldID)] = struct{}{}
						}
					}

					result.Excluded++
					result.Failures = append(result.Failures, RecordFailure{Schema: payloadSet.Schema, Resource: p.resource, Err: fmt.Errorf("%w: %s", ErrIngestRecordExcluded, entry.LastError)})

					logx.FromContext(p.ctx).Warn().Int("attempts", entry.Attempts).Msg("ingest excluded record skipped")

					if entry.Attempts >= ingestMaxRecordAttempts {
						delete(tracked, trackKey)
					}

					continue
				}
			}

			outcome, handleErr := handle(p.ctx, p.record)

			switch {
			case handleErr == nil:
				result.Succeeded++

				if outcome.managed {
					result.Persisted++
				} else {
					result.Skipped++
				}

				if outcome.changed {
					result.Changed++
				}

				if trackKey != "" {
					if _, excluded := tracked[trackKey]; excluded {
						delete(tracked, trackKey)
						dirty = true
					}
				}

				if set != nil {
					set.seen[outcome.id] = struct{}{}
				}
			case errors.Is(handleErr, entityops.ErrUpsertStaleRun):
				result.Skipped++

				if set != nil {
					set.stale++
				}

				logx.FromContext(p.ctx).Debug().Msg("ingest skipped stale run")
			default:
				wrapped := wrapIngestPersistError(handleErr)

				if trackKey != "" {
					tracked[trackKey] = &models.FailedRecord{Schema: sourceSchema.Snake, Key: recordKey, RunID: options.RunID, Attempts: 1, LastError: wrapped.Error()}
					dirty = true
				}

				requeued := false

				if ic.Runtime != nil {
					if requeueErr := emitMappedRecord(p.ctx, ic.Runtime, ic.Integration, operationName, p.record, options); requeueErr != nil {
						logx.FromContext(p.ctx).Warn().Err(requeueErr).Msg("ingest failure requeue failed")
					} else {
						requeued = true
					}
				}

				logx.FromContext(p.ctx).Error().Err(wrapped).Bool("requeued", requeued).Msg("ingest persist failed")

				result.Failed++
				result.Failures = append(result.Failures, RecordFailure{Schema: payloadSet.Schema, Resource: p.resource, Err: wrapped})
			}
		}
	}

	switch {
	case result.Failed > 0:
		logx.FromContext(ctx).Warn().Int("failed", result.Failed).Int("attempted", result.Attempted).Msg("ingest skipped records that could not be imported")
	default:
		for _, set := range snapshotSets {
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

			if removeErr := set.schema.MarkRemoved(ctx, ic.DB, removed, time.Now(), options.RunID); removeErr != nil {
				return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, removeErr)
			}

			result.Removed += len(removed)
		}
	}

	if result.Skipped > 0 {
		logx.FromContext(ctx).Info().Int("skipped", result.Skipped).Msg("ingest left records unmanaged by this installation")
	}

	if dirty {
		if healthErr := persistFailedRecords(ctx, ic, failedRecordsFromTracked(tracked)); healthErr != nil {
			return result, fmt.Errorf("%w: %w", ErrIngestPersistFailed, healthErr)
		}
	}

	return result, nil
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
	lookups := make([]recordLookup, 0, len(ready))

	for _, p := range ready {
		alternative, values, ok := lookupKeyFor(schema, p.record.Payload)
		if !ok {
			continue
		}

		lookups = append(lookups, recordLookup{alternative: alternative, values: values})
	}

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

			for _, row := range rows {
				rowValues := make(entityops.LookupValues, len(alt.Fields))
				for _, name := range alt.Fields {
					rowValues[name] = entityops.FieldValue(row, name)
				}

				key := entityops.EncodeLookupKey(alt, rowValues)
				keyed[key] = append(keyed[key], row)
			}
		}

		matches[alternative] = keyed
	}

	return entityops.WithLookupMatches(ctx, schema, matches), nil
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
