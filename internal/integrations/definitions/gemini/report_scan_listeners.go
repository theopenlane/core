package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/scan"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	// PartMinInterval is the wait before the first retry of a failed part, doubled on each attempt
	PartMinInterval = 30 * time.Second
	// PartMaxInterval caps the backoff between part attempts
	PartMaxInterval = 5 * time.Minute
	// PartMaxAttempts bounds how many times a section is parsed before it is marked failed
	PartMaxAttempts = 3
)

// ReportScanPartEnvelope carries one section parse job through its attempts
type ReportScanPartEnvelope struct {
	// OrganizationID is the organization that owns the scan
	OrganizationID string `json:"organizationId"`
	// ScanID is the Scan record the part result is stored on
	ScanID string `json:"scanId"`
	// Part is the section name to parse
	Part string `json:"part"`
	// Attempt is the number of parse attempts already made for this part
	Attempt int `json:"attempt"`
	// RefCodes scopes a batch of a control-dependent part to these control ref codes
	RefCodes []string `json:"refCodes,omitempty"`
	// BatchIndex is this job's position among the part's batches, meaningful when BatchCount is set
	BatchIndex int `json:"batchIndex,omitempty"`
	// BatchCount is the total number of batches for the part; zero means the job covers the whole part
	BatchCount int `json:"batchCount,omitempty"`
	// Cache names the cached content holding the report; it rides the job rather than the scan
	// record because the resource name identifies the operator's cloud project
	Cache string `json:"cache,omitempty"`
}

// isBatch reports whether the envelope covers one batch of a fanned-out section
func (e ReportScanPartEnvelope) isBatch() bool {
	return e.BatchCount > 0
}

// reportScanTopics is the namespace for report scan saga topics
var reportScanTopics = gala.IntegrationRun.At("reportscan.part")

// reportScanPartTopic is the durable per-section parse topic; per-attempt keys dedup crash-retry re-emissions
var reportScanPartTopic = gala.NamespacedTopicFor(reportScanTopics, gala.WithUniqueKey(func(e ReportScanPartEnvelope) string {
	return reportScanTopics.Key(e.ScanID, e.Part, strconv.Itoa(e.BatchIndex), strconv.Itoa(e.Attempt))
}))

// reportScanSaga orchestrates the durable report scan flow: one parse job per section, assembled by the last to finish
type reportScanSaga struct {
	services types.RuntimeServices
}

// reportScanListeners declares the standalone gala listeners implementing the report scan saga
func reportScanListeners() types.GalaListenerRegistration {
	return types.GalaListenerRegistration{
		Name: "gemini.reportscan",
		Register: func(g *gala.Gala, services types.RuntimeServices) ([]gala.ListenerID, error) {
			saga := reportScanSaga{services: services}

			return gala.Register(g, gala.Definition[ReportScanPartEnvelope]{
				Topic: reportScanPartTopic,
				LogFields: func(envelope ReportScanPartEnvelope) map[string]any {
					fields := map[string]any{
						"organization_id": envelope.OrganizationID,
						"scan_id":         envelope.ScanID,
						"part":            envelope.Part,
						"part_attempt":    envelope.Attempt,
					}

					if envelope.isBatch() {
						fields["batch"] = fmt.Sprintf("%d/%d", envelope.BatchIndex+1, envelope.BatchCount)
						fields["batch_controls"] = len(envelope.RefCodes)
					}

					return fields
				},
				Handle: func(hc gala.HandlerContext, envelope ReportScanPartEnvelope) error {
					return saga.handlePart(hc.Context, envelope)
				},
			})
		},
	}
}

// handlePart parses one section, re-emitting itself with backoff on failure until the attempt
// budget runs out, then stores the outcome on the scan and finalizes the scan once every part is done
func (s reportScanSaga) handlePart(ctx context.Context, envelope ReportScanPartEnvelope) error {
	config, err := json.Marshal(ReportScanPartRequest{
		ScanID:         envelope.ScanID,
		OrganizationID: envelope.OrganizationID,
		Part:           envelope.Part,
		RefCodes:       envelope.RefCodes,
		Cache:          envelope.Cache,
	})
	if err != nil {
		return err
	}

	response, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), ReportScanPartOp.Name(), config)

	// the parse may have consumed the job deadline; follow-up writes and re-emits must still land
	ctx = context.WithoutCancel(ctx)

	if err != nil {
		return s.retryOrFailPart(ctx, envelope, err)
	}

	var result ReportScanPartResult
	if err := json.Unmarshal(response, &result); err != nil {
		return s.retryOrFailPart(ctx, envelope, err)
	}

	section, ok := result.Sections[envelope.Part]
	if !ok {
		section = json.RawMessage("[]")
	}

	// a section far below its expected size is treated like a failed pass while attempts remain
	thin := thinSectionError(envelope, section)
	if thin != nil && envelope.Attempt+1 < PartMaxAttempts {
		return s.retryOrFailPart(ctx, envelope, thin)
	}

	if thin != nil {
		logx.FromContext(ctx).Warn().Err(thin).Msg("report scan: storing thin section after exhausting attempts")
	}

	if envelope.isBatch() {
		if err := s.storeBatchResult(ctx, envelope, section); err != nil {
			return err
		}

		logx.FromContext(ctx).Debug().Int("items", sectionItemCount(section)).Msg("report scan: batch stored")

		return s.completeBatchedPart(ctx, envelope)
	}

	logx.FromContext(ctx).Debug().Int("items", sectionItemCount(section)).Msg("report scan: section stored")

	if err := s.storePartResult(ctx, envelope, section, thin); err != nil {
		return err
	}

	if envelope.Part == soc2.ControlsPart {
		refCodes := soc2.ControlRefCodes(section)

		for _, part := range soc2.ControlScopedParts {
			if err := s.fanOutDependentPart(ctx, envelope, part, refCodes, controlScopedBatchSizes[part]); err != nil {
				return err
			}
		}
	}

	return s.maybeFinalize(ctx, envelope)
}

// fanOutDependentPart schedules one job per batch of control ref codes for a control-scoped part
// that is still pending
func (s reportScanSaga) fanOutDependentPart(ctx context.Context, envelope ReportScanPartEnvelope, part string, refCodes []string, batchSize int) error {
	pending, err := s.partPending(ctx, envelope.OrganizationID, envelope.ScanID, part)
	if err != nil || !pending {
		return err
	}

	batches := lo.Chunk(refCodes, batchSize)
	if len(batches) == 0 {
		batches = [][]string{nil}
	}

	progress, err := json.Marshal(map[string]any{"total": len(batches), "completed": 0, "failed": 0})
	if err != nil {
		return err
	}

	if err := s.setMetadataPaths(ctx, envelope, metadataWrite{path: jsonPath(SummaryMetadataKey, part, summaryBatchesKey), value: progress}); err != nil {
		return err
	}

	for index, batch := range batches {
		if _, err := s.services.Gala().EmitWithHeaders(ctx, reportScanPartTopic.Name, ReportScanPartEnvelope{
			OrganizationID: envelope.OrganizationID,
			ScanID:         envelope.ScanID,
			Part:           part,
			RefCodes:       batch,
			BatchIndex:     index,
			BatchCount:     len(batches),
			Cache:          envelope.Cache,
		}, gala.Headers{UniqueOnce: true}); err != nil {
			return err
		}
	}

	logx.FromContext(ctx).Debug().Str("dependent_part", part).Int("batches", len(batches)).Int("controls", len(refCodes)).Msg("report scan: batches scheduled")

	return nil
}

// partPending reports whether the named part is still waiting on a result
func (s reportScanSaga) partPending(ctx context.Context, organizationID, scanID, part string) (bool, error) {
	scanRecord, err := s.services.DB().Scan.Get(scanSystemContext(ctx, organizationID), scanID)
	if err != nil {
		return false, err
	}

	return summaryStatus(scanRecord.Metadata, part) == PartStatePending, nil
}

// summaryStatus reads a part's state from the report summary
func summaryStatus(metadata map[string]any, part string) string {
	entry, _ := metadata[SummaryMetadataKey].(map[string]any)[part].(map[string]any)
	status, _ := entry["status"].(string)

	return status
}

// completeBatchedPart marks the batched part completed or failed once every batch has reported,
// then finalizes the scan. A part with at least one successful batch counts as completed
func (s reportScanSaga) completeBatchedPart(ctx context.Context, envelope ReportScanPartEnvelope) error {
	systemCtx := scanSystemContext(ctx, envelope.OrganizationID)

	scanRecord, err := s.services.DB().Scan.Get(systemCtx, envelope.ScanID)
	if err != nil {
		return err
	}

	entry, _ := scanRecord.Metadata[SummaryMetadataKey].(map[string]any)[envelope.Part].(map[string]any)
	batches, _ := entry[summaryBatchesKey].(map[string]any)
	total, _ := batches["total"].(float64)
	completed, _ := batches["completed"].(float64)
	failed, _ := batches["failed"].(float64)
	reported := int(completed + failed)

	if reported < int(total) {
		logx.FromContext(ctx).Debug().Int("reported", reported).Int("total", int(total)).Msg("report scan: waiting on remaining batches")

		return nil
	}

	state := PartStateCompleted
	if completed == 0 {
		state = PartStateFailed
	}

	var writes []metadataWrite

	count := 0

	// every batch has reported, so the appended section can be rewritten without racing a sibling
	if report, ok := scanRecord.Metadata[ReportMetadataKey].(map[string]any); ok {
		items, _ := report[envelope.Part].([]any)
		count = len(items)

		if deduped, removed, err := dedupeByExternalID(report[envelope.Part]); err == nil && removed > 0 {
			logx.FromContext(ctx).Debug().Int("removed", removed).Msg("report scan: dropped duplicate items across batches")

			count -= removed

			writes = append(writes, metadataWrite{path: jsonPath(ReportMetadataKey, envelope.Part), value: deduped})
		}
	}

	var partWarning error

	if failed > 0 {
		partWarning = fmt.Errorf("%d of %d batches of the %s section could not be extracted, %w", int(failed), reported, envelope.Part, ErrResultsIncomplete)
	}

	summaryWrites, err := summaryEntryWrites(envelope.Part, state, count, partWarning)
	if err != nil {
		return err
	}

	writes = append(writes, summaryWrites...)

	// batch progress only exists to detect the last batch, so it is dropped once the part is done
	writes = append(writes, metadataWrite{path: jsonPath(SummaryMetadataKey, envelope.Part, summaryBatchesKey), remove: true})

	if err := s.setMetadataPaths(ctx, envelope, writes...); err != nil {
		return err
	}

	return s.maybeFinalize(ctx, envelope)
}

// dedupeByExternalID drops items that repeat an earlier item's externalID, keeping the first
// occurrence, and reports how many were removed
func dedupeByExternalID(section any) (json.RawMessage, int, error) {
	items, ok := section.([]any)
	if !ok {
		return nil, 0, nil
	}

	seen := make(map[string]bool, len(items))
	kept := make([]any, 0, len(items))

	for _, item := range items {
		object, _ := item.(map[string]any)
		externalID, _ := object["externalID"].(string)

		if externalID != "" && seen[externalID] {
			continue
		}

		seen[externalID] = true

		kept = append(kept, item)
	}

	encoded, err := json.Marshal(kept)
	if err != nil {
		return nil, 0, err
	}

	return encoded, len(items) - len(kept), nil
}

// storeBatchResult appends the batch's items to the section and counts the batch as completed
func (s reportScanSaga) storeBatchResult(ctx context.Context, envelope ReportScanPartEnvelope, section json.RawMessage) error {
	return s.setMetadataPaths(ctx, envelope,
		metadataWrite{path: jsonPath(ReportMetadataKey, envelope.Part), value: section, appendArray: true},
		metadataWrite{path: jsonPath(SummaryMetadataKey, envelope.Part, summaryBatchesKey, "completed"), increment: true},
	)
}

// storeBatchFailure counts the batch as failed so the part can still complete from its siblings
func (s reportScanSaga) storeBatchFailure(ctx context.Context, envelope ReportScanPartEnvelope) error {
	return s.setMetadataPaths(ctx, envelope,
		metadataWrite{path: jsonPath(SummaryMetadataKey, envelope.Part, summaryBatchesKey, "failed"), increment: true},
	)
}

// sectionItemCount returns how many items a parsed section array holds, zero when it is not an array
func sectionItemCount(section json.RawMessage) int {
	var items []json.RawMessage
	if err := json.Unmarshal(section, &items); err != nil {
		return 0
	}

	return len(items)
}

// retryOrFailPart schedules another attempt while budget remains, otherwise records the failure
func (s reportScanSaga) retryOrFailPart(ctx context.Context, envelope ReportScanPartEnvelope, cause error) error {
	if envelope.Attempt+1 < PartMaxAttempts {
		scheduledAt := time.Now().Add(partBackoff(envelope.Attempt))

		logx.FromContext(ctx).Warn().Err(cause).Time("scheduled_at", scheduledAt).Msg("report scan: part failed, scheduling retry")

		// the retry keeps the batch scope so a failed batch does not come back as the whole part
		retry := envelope
		retry.Attempt = envelope.Attempt + 1

		_, err := s.services.Gala().EmitWithHeaders(ctx, reportScanPartTopic.Name, retry, gala.Headers{ScheduledAt: &scheduledAt, UniqueOnce: true})

		return err
	}

	// the technical cause stays in the logs; the stored reason is what the user reads
	logx.FromContext(ctx).Error().Err(cause).Msg("report scan: part exhausted attempts, marking failed")

	if envelope.isBatch() {
		if err := s.storeBatchFailure(ctx, envelope); err != nil {
			return err
		}

		return s.completeBatchedPart(ctx, envelope)
	}

	if err := s.storePartFailure(ctx, envelope, fmt.Errorf("the %s section %w after %d attempts", envelope.Part, ErrSectionFailed, PartMaxAttempts)); err != nil {
		return err
	}

	// dependent parts are scoped by the controls, so a failed controls section fails them too
	if envelope.Part == soc2.ControlsPart {
		for _, part := range soc2.ControlScopedParts {
			if err := s.failDependentPart(ctx, envelope, part); err != nil {
				return err
			}
		}
	}

	return s.maybeFinalize(ctx, envelope)
}

// failDependentPart marks a still-pending part failed because the part it depends on gave up
func (s reportScanSaga) failDependentPart(ctx context.Context, envelope ReportScanPartEnvelope, part string) error {
	pending, err := s.partPending(ctx, envelope.OrganizationID, envelope.ScanID, part)
	if err != nil || !pending {
		return err
	}

	dependent := envelope
	dependent.Part = part

	return s.storePartFailure(ctx, dependent, fmt.Errorf("the %s section was skipped because the %s section %w", part, envelope.Part, ErrSectionFailed))
}

// scanFailedReason is the user-facing reason recorded when no section could be extracted
const scanFailedReason = "no sections could be extracted from the report, please try again"

// metadataWrite is one nested jsonb path and the value to set there; appendArray concatenates the
// value onto the existing array at the path instead of replacing it
type metadataWrite struct {
	path string
	// value is the jsonb to set; ignored when increment is set
	value json.RawMessage
	// appendArray concatenates value onto the existing array at path instead of replacing it
	appendArray bool
	// increment adds one to the integer at path, treating a missing value as zero
	increment bool
	// remove deletes the value at path
	remove bool
}

// storePartResult writes the section and marks the part completed in one atomic jsonb update, so
// concurrent part jobs never overwrite each other's results; a warning is recorded alongside
// when the section was accepted despite looking thin
func (s reportScanSaga) storePartResult(ctx context.Context, envelope ReportScanPartEnvelope, section json.RawMessage, warning error) error {
	writes, err := summaryEntryWrites(envelope.Part, PartStateCompleted, sectionItemCount(section), warning)
	if err != nil {
		return err
	}

	writes = append(writes, metadataWrite{path: jsonPath(ReportMetadataKey, envelope.Part), value: section})

	return s.setMetadataPaths(ctx, envelope, writes...)
}

// summaryEntryWrites builds the path writes for one section's summary: its state, item count, and
// error if any; writing the fields individually preserves batch progress recorded alongside them
func summaryEntryWrites(part, state string, count int, cause error) ([]metadataWrite, error) {
	stateValue, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}

	countValue, err := json.Marshal(count)
	if err != nil {
		return nil, err
	}

	writes := []metadataWrite{
		{path: jsonPath(SummaryMetadataKey, part, "status"), value: stateValue},
		{path: jsonPath(SummaryMetadataKey, part, "count"), value: countValue},
	}

	if cause != nil {
		message, err := json.Marshal(cause.Error())
		if err != nil {
			return nil, err
		}

		writes = append(writes, metadataWrite{path: jsonPath(SummaryMetadataKey, part, "error"), value: message})
	}

	return writes, nil
}

// summaryBatchesKey is the summary entry field holding batch progress for a control-scoped section
const summaryBatchesKey = "batches"

// storePartFailure records the cause and marks the part failed in one atomic jsonb update
func (s reportScanSaga) storePartFailure(ctx context.Context, envelope ReportScanPartEnvelope, cause error) error {
	writes, err := summaryEntryWrites(envelope.Part, PartStateFailed, 0, cause)
	if err != nil {
		return err
	}

	return s.setMetadataPaths(ctx, envelope, writes...)
}

// setMetadataPaths applies nested jsonb_set writes to the scan metadata in a single statement
func (s reportScanSaga) setMetadataPaths(ctx context.Context, envelope ReportScanPartEnvelope, writes ...metadataWrite) error {
	systemCtx := scanSystemContext(ctx, envelope.OrganizationID)

	return s.services.DB().Scan.UpdateOneID(envelope.ScanID).Modify(func(u *sql.UpdateBuilder) {
		u.Set(scan.FieldMetadata, sql.ExprFunc(func(b *sql.Builder) {
			removes, sets := lo.FilterReject(writes, func(write metadataWrite, _ int) bool { return write.remove })

			for range sets {
				b.WriteString("jsonb_set(")
			}

			// removals apply to the base document before any set, so a removed path cannot be re-created
			b.WriteString("(coalesce(" + scan.FieldMetadata + ", '{}'::jsonb)")

			for _, write := range removes {
				b.WriteString(" #- ")
				b.Arg(write.path).WriteString("::text[]")
			}

			b.WriteString(")")

			for _, write := range sets {
				b.WriteString(", ")
				b.Arg(write.path).WriteString("::text[], ")

				switch {
				case write.increment:
					b.WriteString("to_jsonb(coalesce((" + scan.FieldMetadata + " #>> ")
					b.Arg(write.path).WriteString("::text[])::int, 0) + 1), true)")
				case write.appendArray:
					b.WriteString("coalesce(" + scan.FieldMetadata + " #> ")
					b.Arg(write.path).WriteString("::text[], '[]'::jsonb) || ")
					b.Arg(string(write.value)).WriteString("::jsonb, true)")
				default:
					b.Arg(string(write.value)).WriteString("::jsonb, true)")
				}
			}
		}))
	}).Exec(systemCtx)
}

// maybeFinalize completes the scan once no part is still pending; the conditional status update
// guarantees only one of several concurrently finishing parts performs the transition
func (s reportScanSaga) maybeFinalize(ctx context.Context, envelope ReportScanPartEnvelope) error {
	systemCtx := scanSystemContext(ctx, envelope.OrganizationID)

	scanRecord, err := s.services.DB().Scan.Get(systemCtx, envelope.ScanID)
	if err != nil {
		return err
	}

	summary, _ := scanRecord.Metadata[SummaryMetadataKey].(map[string]any)

	completed := 0

	for part := range summary {
		switch summaryStatus(scanRecord.Metadata, part) {
		case PartStatePending:
			return nil
		case PartStateCompleted:
			completed++
		}
	}

	status := enums.ScanStatusCompleted

	// no part is pending here, so rewriting the whole metadata map cannot race a sibling write
	metadata := scanRecord.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	if completed == 0 {
		status = enums.ScanStatusFailed
		metadata[ErrorMetadataKey] = scanFailedReason
	}

	// a not-found here means a sibling part finalized first, which is the expected race outcome
	err = s.services.DB().Scan.UpdateOneID(envelope.ScanID).
		Where(scan.StatusEQ(enums.ScanStatusProcessing)).
		SetStatus(status).
		SetMetadata(metadata).
		Exec(systemCtx)
	if generated.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	logx.FromContext(ctx).Debug().Str("status", status.String()).Int("completed_parts", completed).Int("parts", len(summary)).Msg("report scan: finalized")

	s.releaseCache(ctx, envelope)

	return nil
}

// releaseCache drops the cached report once the scan is final; failure only leaves the cache to expire on its own
func (s reportScanSaga) releaseCache(ctx context.Context, envelope ReportScanPartEnvelope) {
	if envelope.Cache == "" {
		return
	}

	config, err := json.Marshal(ReportScanReleaseRequest{ScanID: envelope.ScanID, OrganizationID: envelope.OrganizationID, Cache: envelope.Cache})
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("report scan: failed encoding cache release")

		return
	}

	if _, err := s.services.ExecuteRuntimeOperation(ctx, DefinitionID.ID(), ReportScanReleaseOp.Name(), config); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("report scan: failed releasing document cache")
	}
}

// jsonPath builds a postgres text[] literal for a nested jsonb path
func jsonPath(keys ...string) string {
	path := "{"
	for i, key := range keys {
		if i > 0 {
			path += ","
		}

		path += key
	}

	return path + "}"
}

// partBackoff returns the wait before the next attempt, doubling from PartMinInterval up to
// PartMaxInterval with jitter so sibling parts do not retry in lockstep
func partBackoff(attempt int) time.Duration {
	interval := PartMinInterval
	for i := 0; i < attempt && interval < PartMaxInterval; i++ {
		interval *= 2
	}

	if interval > PartMaxInterval {
		interval = PartMaxInterval
	}

	jitter := time.Duration(rand.Int64N(int64(interval) / 4)) //nolint:gosec,mnd

	return interval + jitter
}
