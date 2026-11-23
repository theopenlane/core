package vulnintel

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// enricher fills the empty advisory fields of an installation's vulnerabilities from OSV
type enricher struct {
	// services exposes the database and the objectstore runtime operations backing the cache
	services types.RuntimeServices
	// osv looks advisories up
	osv *osvClient
	// maxPerRun caps the OSV lookups one enrichment makes
	maxPerRun int
	// now returns the current time, overridden in tests
	now func() time.Time
}

// vulnerabilityPageSize is how many vulnerabilities one enrichment loads per query
const vulnerabilityPageSize = 500

// enrichResult counts the work one enrichment pass did
type enrichResult struct {
	// Considered counts the vulnerabilities with an advisory id
	Considered int
	// Updated counts the vulnerabilities that gained a field
	Updated int
	// Lookups counts the OSV lookups made
	Lookups int
	// Deferred counts the advisories left for a later run once the lookup cap was reached
	Deferred int
}

// enrichRun is the state of one enrichment pass
type enrichRun struct {
	*enricher
	// cache holds the advisories already looked up across runs
	cache *advisoryCache
	// advisories memoizes this run's advisories by id, nil when there is none to apply
	advisories map[string]*osvRecord
	// result counts the work done
	result enrichResult
	// failures counts the OSV lookups that failed in a row
	failures int
}

// maxConsecutiveFailures stops a run's OSV lookups once this many fail in a row, leaving the rest for the next run
const maxConsecutiveFailures = 5

// enrich fills the empty advisory fields of the open vulnerabilities the installation wrote, leaving every field the
// provider set untouched, and writes rows directly so the provider's provenance on them is kept
func (e *enricher) enrich(ctx context.Context, envelope operations.IngestCompletedEnvelope) (enrichResult, error) {
	ctx = auth.EnsureIntegrationCaller(ctx, envelope.OwnerID)

	run := &enrichRun{enricher: e, cache: &advisoryCache{services: e.services}, advisories: map[string]*osvRecord{}}
	afterID := ""

	for {
		page, err := e.services.DB().Vulnerability.Query().
			Where(
				vulnerability.HasIntegrationsWith(integration.ID(envelope.IntegrationID)),
				vulnerability.OpenEQ(true),
				vulnerability.CveIDNotNil(),
				vulnerability.CveIDNEQ(""),
				vulnerability.IDGT(afterID),
			).
			Order(vulnerability.ByID()).
			Limit(vulnerabilityPageSize).
			All(ctx)
		if err != nil {
			return run.result, err
		}

		for _, vuln := range page {
			run.apply(ctx, vuln)
		}

		if len(page) < vulnerabilityPageSize {
			return run.result, nil
		}

		afterID = page[len(page)-1].ID
	}
}

// apply fills one vulnerability from its advisory when one is available
func (r *enrichRun) apply(ctx context.Context, vuln *generated.Vulnerability) {
	r.result.Considered++

	if enriched(vuln) {
		return
	}

	record := r.advisory(ctx, vuln.CveID)
	if record == nil {
		return
	}

	update, changed := advisoryUpdate(r.services.DB().Vulnerability.UpdateOneID(vuln.ID), vuln, *record)
	if !changed {
		return
	}

	if err := update.Exec(ctx); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("vulnerability_id", vuln.ID).Msg("vulnintel: failed to enrich vulnerability")

		return
	}

	r.result.Updated++
}

// advisory returns the OSV advisory for the id from this run, the cache or OSV, nil when OSV has none, the lookup
// failed, or the run already made its maximum lookups
func (r *enrichRun) advisory(ctx context.Context, rawID string) *osvRecord {
	id, ok := advisoryID(rawID)
	if !ok {
		return nil
	}

	if record, ok := r.advisories[id]; ok {
		return record
	}

	now := r.now()

	if entry, ok := r.cache.get(ctx, id, now); ok {
		r.advisories[id] = cachedRecord(entry)

		return r.advisories[id]
	}

	if r.result.Lookups >= r.maxPerRun || r.failures >= maxConsecutiveFailures || ctx.Err() != nil {
		r.result.Deferred++
		r.advisories[id] = nil

		return nil
	}

	r.result.Lookups++

	record, found, err := r.osv.lookup(ctx, id)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("id", id).Msg("vulnintel: osv lookup failed, skipping advisory")

		r.failures++
		r.advisories[id] = nil

		return nil
	}

	r.failures = 0

	entry := cacheEntry{FetchedAt: now, Found: found, Record: record}
	r.cache.put(ctx, id, entry)
	r.advisories[id] = cachedRecord(entry)

	return r.advisories[id]
}

// enriched reports whether the vulnerability already has every field an advisory can fill, so no lookup is needed
func enriched(vuln *generated.Vulnerability) bool {
	return vuln.Summary != "" && vuln.Description != "" && vuln.Vector != "" && vuln.PublishedAt != nil && len(vuln.References) > 0
}

// advisoryIDPattern matches OSV identifiers such as CVE-2026-1, GHSA-xxxx-xxxx-xxxx or PYSEC-2026-1
var advisoryIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// advisoryID trims the provider's advisory id, false when it is not a safe OSV identifier to use as a cache key
func advisoryID(rawID string) (string, bool) {
	id := strings.TrimSpace(rawID)

	return id, len(id) <= maxAdvisoryIDLength && advisoryIDPattern.MatchString(id) && !strings.Contains(id, "..")
}

// maxAdvisoryIDLength bounds the advisory ids looked up
const maxAdvisoryIDLength = 128

// cachedRecord returns the advisory of a lookup, nil when OSV had none
func cachedRecord(entry cacheEntry) *osvRecord {
	if !entry.Found {
		return nil
	}

	return &entry.Record
}

// advisoryUpdate sets each advisory field the vulnerability does not have yet and merges the advisory references,
// reporting whether anything changed
func advisoryUpdate(update *generated.VulnerabilityUpdateOne, vuln *generated.Vulnerability, record osvRecord) (*generated.VulnerabilityUpdateOne, bool) {
	changed := false

	if vuln.Summary == "" && record.Summary != "" {
		update.SetSummary(record.Summary)
		changed = true
	}

	if vuln.Description == "" && record.Details != "" {
		update.SetDescription(record.Details)
		changed = true
	}

	if vuln.Vector == "" && record.vector() != "" {
		update.SetVector(record.vector())
		changed = true
	}

	if published := record.publishedAt(); vuln.PublishedAt == nil && !published.IsZero() {
		update.SetPublishedAt(models.DateTime(published))
		changed = true
	}

	if len(vuln.References) == 0 {
		update.SetReferences(record.referenceURLs())
		changed = true
	}

	return update, changed
}
