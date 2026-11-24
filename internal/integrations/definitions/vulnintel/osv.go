package vulnintel

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ossf/osv-schema/bindings/go/osvschema"
	gocvss30 "github.com/pandatix/go-cvss/30"
	gocvss31 "github.com/pandatix/go-cvss/31"
	gocvss40 "github.com/pandatix/go-cvss/40"
	"osv.dev/bindings/go/osvdev"
)

const (
	// osvPageURL is the public OSV page of one advisory, added to the vulnerability references
	osvPageURL = "https://osv.dev/vulnerability/"
	// osvRequestTimeout bounds one OSV lookup
	osvRequestTimeout = 15 * time.Second
	// osvMaxAttempts is how many times the client tries one lookup
	osvMaxAttempts = 2
	// osvNotFoundStatus is how the OSV client reports an unknown advisory, it returns no typed not found error
	osvNotFoundStatus = `status="404 Not Found"`
	// severityCVSSV3 is the OSV severity type carrying a CVSS 3.x vector
	severityCVSSV3 = "CVSS_V3"
	// severityCVSSV4 is the OSV severity type carrying a CVSS 4.0 vector
	severityCVSSV4 = "CVSS_V4"
)

// osvRecord is the subset of an OSV advisory used to enrich vulnerabilities, and the shape cached in the bucket
type osvRecord struct {
	// ID is the OSV advisory identifier
	ID string `json:"id"`
	// Summary is the one line advisory summary
	Summary string `json:"summary,omitempty"`
	// Details is the advisory description
	Details string `json:"details,omitempty"`
	// Published is when the advisory was published
	Published string `json:"published,omitempty"`
	// Severity lists the CVSS vectors scored for the advisory
	Severity []osvSeverity `json:"severity,omitempty"`
	// References are links about the advisory
	References []string `json:"references,omitempty"`
}

// osvSeverity is one scored severity of an advisory
type osvSeverity struct {
	// Type is the scoring system, such as CVSS_V3
	Type string `json:"type"`
	// Score is the vector string for CVSS types
	Score string `json:"score"`
}

// recordFromOSV keeps the fields enrichment uses from an OSV advisory
func recordFromOSV(vuln *osvschema.Vulnerability) osvRecord {
	record := osvRecord{
		ID:      vuln.GetId(),
		Summary: vuln.GetSummary(),
		Details: vuln.GetDetails(),
	}

	if published := vuln.GetPublished(); published != nil {
		record.Published = published.AsTime().UTC().Format(time.RFC3339)
	}

	for _, severity := range vuln.GetSeverity() {
		record.Severity = append(record.Severity, osvSeverity{Type: severity.GetType().String(), Score: severity.GetScore()})
	}

	for _, reference := range vuln.GetReferences() {
		if reference.GetUrl() != "" {
			record.References = append(record.References, reference.GetUrl())
		}
	}

	return record
}

// vector returns the CVSS vector whose base score equals the stored score, else the CVSS 4.0 vector, else the
// CVSS 3.x vector
func (r osvRecord) vector(score float64) string {
	if score > 0 {
		for _, severity := range r.Severity {
			if base, ok := baseScore(severity.Score); ok && base == roundScore(score) {
				return severity.Score
			}
		}
	}

	if vector := r.vectorOfType(severityCVSSV4); vector != "" {
		return vector
	}

	return r.vectorOfType(severityCVSSV3)
}

// vectorOfType returns the advisory's first vector of the severity type
func (r osvRecord) vectorOfType(severityType string) string {
	for _, severity := range r.Severity {
		if severity.Type == severityType && severity.Score != "" {
			return severity.Score
		}
	}

	return ""
}

// baseScore computes the base score of a CVSS 3.0, 3.1 or 4.0 vector, false when it does not parse
func baseScore(vector string) (float64, bool) {
	switch {
	case strings.HasPrefix(vector, "CVSS:3.1/"):
		parsed, err := gocvss31.ParseVector(vector)
		if err != nil {
			return 0, false
		}

		return roundScore(parsed.BaseScore()), true
	case strings.HasPrefix(vector, "CVSS:3.0/"):
		parsed, err := gocvss30.ParseVector(vector)
		if err != nil {
			return 0, false
		}

		return roundScore(parsed.BaseScore()), true
	case strings.HasPrefix(vector, "CVSS:4.0/"):
		parsed, err := gocvss40.ParseVector(vector)
		if err != nil {
			return 0, false
		}

		return roundScore(parsed.Score()), true
	default:
		return 0, false
	}
}

// roundScore rounds a score to one decimal, the precision CVSS scores are published at
func roundScore(score float64) float64 {
	return math.Round(score*scorePrecision) / scorePrecision
}

// scorePrecision is the scale rounding scores to one decimal
const scorePrecision = 10

// referenceURLs returns the OSV page of the advisory followed by its reference links
func (r osvRecord) referenceURLs() []string {
	return append([]string{osvPageURL + r.ID}, r.References...)
}

// publishedAt parses the advisory publish time, zero when absent or unparsable
func (r osvRecord) publishedAt() time.Time {
	published, err := time.Parse(time.RFC3339, r.Published)
	if err != nil {
		return time.Time{}
	}

	return published
}

// osvClient looks up advisories on the public OSV API, which needs no credentials
type osvClient struct {
	// client is the OSV API client
	client *osvdev.OSVClient
}

// newOSVClient returns a client for the public OSV API with a bounded request timeout
func newOSVClient() *osvClient {
	client := osvdev.DefaultClient()
	client.HTTPClient = &http.Client{Timeout: osvRequestTimeout}
	// the client retries with sleeps that ignore context, so allow one retry and leave the rest to the next run
	client.Config.MaxRetryAttempts = osvMaxAttempts

	return &osvClient{client: client}
}

// lookup fetches one advisory by any OSV identifier or alias, found is false when OSV has no such advisory
func (c *osvClient) lookup(ctx context.Context, id string) (record osvRecord, found bool, err error) {
	vuln, err := c.client.GetVulnByID(ctx, url.PathEscape(strings.TrimSpace(id)))
	if err != nil {
		if strings.Contains(err.Error(), osvNotFoundStatus) {
			return osvRecord{}, false, nil
		}

		return osvRecord{}, false, fmt.Errorf("%w: %w", ErrOSVRequest, err)
	}

	return recordFromOSV(vuln), true, nil
}
