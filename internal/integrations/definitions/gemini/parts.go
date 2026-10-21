package gemini

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/precis/docextract"
	"github.com/theopenlane/precis/docextract/kinds/soc2"

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
	// PartWaveSize is how many part jobs start together, so a scan does not exhaust the per-minute token quota at once
	PartWaveSize = 3
	// PartWaveInterval is the delay between waves of part jobs
	PartWaveInterval = 30 * time.Second
	// PartQuotaMinInterval is the wait before re-requesting a part the model rejected for quota, doubled on each further rejection
	PartQuotaMinInterval = time.Minute
	// PartQuotaMaxWaits bounds how many quota rejections a part waits out before it counts against its attempts
	PartQuotaMaxWaits = 6
)

// RequestedParts reads the sections a scan asked for from its metadata, validating every name;
// an absent or empty list means every configured section
func RequestedParts(metadata map[string]any) ([]string, error) {
	raw, _ := metadata[requestedPartsMetadataKey].([]any)

	names := make([]string, 0, len(raw))

	for _, value := range raw {
		name, _ := value.(string)
		if !soc2.IsPart(name) {
			return nil, fmt.Errorf("%w: %q, expected one of %s", ErrUnknownRequestedPart, name, strings.Join(soc2.Parts, ", "))
		}

		names = append(names, name)
	}

	return names, nil
}

// selectParts narrows the configured sections to the ones the scan asked for, or to the profile's
// sections when it asked for none. Selecting no configured section is an error, as is a profile that
// requires all of its sections missing any
func (r ReportScanRequest) selectParts(metadata map[string]any, profile soc2.Profile) ([]string, error) {
	requested, err := RequestedParts(metadata)
	if err != nil {
		return nil, err
	}

	if len(requested) == 0 {
		requested = profile.Parts
	}

	selected := lo.Filter(r.parts, func(name string, _ int) bool { return slices.Contains(requested, name) })
	if len(selected) == 0 || (profile.RequireAll && len(selected) < len(lo.Uniq(requested))) {
		return nil, fmt.Errorf("%w: requested %s, configured %s", ErrRequestedPartsUnavailable, strings.Join(requested, ", "), strings.Join(r.parts, ", "))
	}

	return selected, nil
}

// scheduleParts emits one part job per section, holding back the control-scoped sections until the
// controls part lands so they can be batched by ref code
func (s reportScanSaga) scheduleParts(ctx context.Context, organizationID, scanID, reportFileID string, partNames []string, cache string) error {
	deferDependents := slices.Contains(partNames, soc2.ControlsPart)
	now := time.Now()
	scheduled := 0

	for _, name := range orderParts(partNames) {
		if deferDependents && slices.Contains(soc2.ControlScopedParts, name) {
			continue
		}

		if _, err := s.services.Gala().EmitWithHeaders(ctx, reportScanPartTopic.Name, ReportScanPartEnvelope{
			OrganizationID: organizationID,
			ScanID:         scanID,
			Part:           name,
			Cache:          cache,
			ReportFileID:   reportFileID,
		}, gala.Headers{ScheduledAt: partStartAt(now, scheduled), UniqueOnce: true}); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("part", name).Msg("report scan: failed scheduling part job")

			return err
		}

		scheduled++
	}

	logx.FromContext(ctx).Debug().Int("parts", len(partNames)).Str("cache", docextract.CacheID(cache)).Msg("report scan: part jobs scheduled")

	return nil
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

	if err := s.setMetadataPaths(ctx, envelope, metadataWrite{path: jsonPath(summaryMetadataKey, part, summaryBatchesKey), value: progress}); err != nil {
		return err
	}

	now := time.Now()

	for index, batch := range batches {
		if _, err := s.services.Gala().EmitWithHeaders(ctx, reportScanPartTopic.Name, ReportScanPartEnvelope{
			OrganizationID: envelope.OrganizationID,
			ScanID:         envelope.ScanID,
			Part:           part,
			RefCodes:       batch,
			BatchIndex:     index,
			BatchCount:     len(batches),
			Cache:          envelope.Cache,
			ReportFileID:   envelope.ReportFileID,
		}, gala.Headers{ScheduledAt: partStartAt(now, index), UniqueOnce: true}); err != nil {
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

// orderParts sorts the requested parts into the SOC 2 request order, so the parts others depend on start in the first wave
func orderParts(partNames []string) []string {
	order := soc2.Parts
	ordered := slices.Clone(partNames)

	slices.SortStableFunc(ordered, func(a, b string) int {
		return cmp.Compare(slices.Index(order, a), slices.Index(order, b))
	})

	return ordered
}

// partStartAt returns when the job at index should start, releasing jobs in waves of PartWaveSize
func partStartAt(now time.Time, index int) *time.Time {
	at := now.Add(time.Duration(index/PartWaveSize) * PartWaveInterval)

	return &at
}

// waitOutQuota reschedules the part after the model rejected it for quota, freeing the worker instead of waiting in
// place; once the quota waits are spent the rejection counts against the part's attempts like any other failure
func (s reportScanSaga) waitOutQuota(ctx context.Context, envelope ReportScanPartEnvelope, quotaErr *docextract.QuotaError) error {
	if envelope.QuotaWaits >= PartQuotaMaxWaits {
		return s.retryOrFailPart(ctx, envelope, quotaErr)
	}

	scheduledAt := time.Now().Add(max(backoffFrom(PartQuotaMinInterval, envelope.QuotaWaits), quotaErr.After))

	logx.FromContext(ctx).Info().Int("quota_waits", envelope.QuotaWaits).Time("scheduled_at", scheduledAt).Msg("report scan: model quota exhausted, rescheduling part")

	retry := envelope
	retry.QuotaWaits = envelope.QuotaWaits + 1

	_, err := s.services.Gala().EmitWithHeaders(ctx, reportScanPartTopic.Name, retry, gala.Headers{ScheduledAt: &scheduledAt, UniqueOnce: true})

	return err
}

// partBackoff returns the wait before the next attempt, doubling from PartMinInterval up to
// PartMaxInterval with jitter so sibling parts do not retry in lockstep
func partBackoff(attempt int) time.Duration {
	return backoffFrom(PartMinInterval, attempt)
}

// backoffFrom doubles base once per attempt up to PartMaxInterval, adding jitter so sibling parts do not come back in lockstep
func backoffFrom(base time.Duration, attempt int) time.Duration {
	interval := base
	for i := 0; i < attempt && interval < PartMaxInterval; i++ {
		interval *= 2
	}

	if interval > PartMaxInterval {
		interval = PartMaxInterval
	}

	jitter := time.Duration(rand.Int64N(int64(interval) / 4)) //nolint:gosec,mnd

	return interval + jitter
}
