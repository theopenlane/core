package aikido

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func issuePage(start, count int, issueType string) string {
	issues := make([]string, 0, count)

	for i := range count {
		issues = append(issues, fmt.Sprintf(`{"id":%d,"type":%q,"code_repo_name":"example-repo","ignored_at":null}`, start+i, issueType))
	}

	return `[` + strings.Join(issues, ",") + `]`
}

func payloadSet(payloads []types.IngestPayloadSet, schema string) []types.MappingEnvelope {
	for _, set := range payloads {
		if set.Schema == schema {
			return set.Envelopes
		}
	}

	return nil
}

func envelopeIDs(t *testing.T, envelopes []types.MappingEnvelope) []int64 {
	t.Helper()

	ids := make([]int64, 0, len(envelopes))

	for _, envelope := range envelopes {
		var issue struct {
			ID int64 `json:"id"`
		}

		assert.NilError(t, json.Unmarshal(envelope.Payload, &issue))

		ids = append(ids, issue.ID)
	}

	return ids
}

func TestIssueResource(t *testing.T) {
	tests := []struct {
		name     string
		issue    map[string]any
		expected string
	}{
		{name: "no resource", issue: map[string]any{}, expected: ""},
		{name: "code repository", issue: map[string]any{"code_repo_name": "repo", "cloud_name": "aws"}, expected: "repo"},
		{name: "container", issue: map[string]any{"code_repo_name": "", "container_repo_name": "api-image"}, expected: "api-image"},
		{name: "domain", issue: map[string]any{"domain_name": "example.com"}, expected: "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, is.Equal(tt.expected, issueResource(tt.issue)))
		})
	}
}

func TestCollectIssuesRoutesByType(t *testing.T) {
	client := newTestClient(t, exportOnly(func(w http.ResponseWriter, r *http.Request) {
		assert.Check(t, is.Equal(issueStatusAll, r.URL.Query().Get("filter_status")))
		assert.Check(t, is.Equal(strconv.Itoa(issueExportPageSize), r.URL.Query().Get("per_page")))

		writeJSON(w, `[`+strings.Trim(issuePage(1, 2, issueTypeOpenSource), "[]")+`,`+strings.Trim(issuePage(3, 3, "sast"), "[]")+`]`)
	}))

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	vulnerabilities := payloadSet(payloads, entityops.SchemaVulnerability.Name)
	findings := payloadSet(payloads, entityops.SchemaFinding.Name)

	assert.Check(t, is.DeepEqual([]int64{1, 2}, envelopeIDs(t, vulnerabilities)))
	assert.Check(t, is.DeepEqual([]int64{3, 4, 5}, envelopeIDs(t, findings)))
	assert.Check(t, is.Equal("example-repo", vulnerabilities[0].Resource))
	assert.Check(t, !strings.Contains(string(vulnerabilities[0].Payload), "ignored_at"))
}

func TestCollectIssuesPagesUntilShortPage(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, exportOnly(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		if r.URL.Query().Get("page") == "0" {
			writeJSON(w, issuePage(1, issueExportPageSize, "sast"))

			return
		}

		writeJSON(w, issuePage(issueExportPageSize+1, 2, "sast"))
	}))

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	assert.Check(t, is.Len(payloadSet(payloads, entityops.SchemaFinding.Name), issueExportPageSize+2))
	assert.Check(t, is.Equal(int32(2), requests.Load()))
}

func TestCollectIssuesStopsWhenPageParamIgnored(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, exportOnly(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writeJSON(w, issuePage(1, issueExportPageSize, "sast"))
	}))

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	assert.Check(t, is.Len(payloadSet(payloads, entityops.SchemaFinding.Name), issueExportPageSize))
	assert.Check(t, is.Equal(int32(2), requests.Load()))
}

func TestCollectIssuesPropagatesAPIErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	_, err := CollectIssues(context.Background(), client)
	assert.Check(t, is.ErrorIs(err, ErrIssuesFetchFailed))
	assert.Check(t, is.ErrorIs(err, ErrUnauthorized))
}

func TestCollectIssuesHonorsContextCancellation(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, issuePage(1, 2, "sast"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := CollectIssues(ctx, client)
	assert.Check(t, is.ErrorIs(err, ErrIssuesFetchFailed))
	assert.Check(t, is.ErrorIs(err, ErrAPIRequest))
}

func exportOnly(export http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != pathIssuesExport {
			writeJSON(w, `[]`)

			return
		}

		export(w, r)
	}
}

func TestCollectIssuesAttachesAttackAnalysis(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathIssuesExport:
			writeJSON(w, `[{"id":7,"type":"ai_pentest","status":"open","assessment_id":"f19e449d"},{"id":8,"type":"sast","status":"open"}]`)
		case fmt.Sprintf(pathAttackAnalysisFormat, "7"):
			writeJSON(w, `{"title":"SQL Injection","summary":"The search endpoint is injectable"}`)
		case fmt.Sprintf(pathAttackAnalysisFormat, "8"):
			t.Error("attack analysis requested for a non pentest issue")
		default:
			writeJSON(w, `[]`)
		}
	})

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	findings := payloadSet(payloads, entityops.SchemaFinding.Name)
	assert.Assert(t, is.Len(findings, 2))

	assert.Check(t, is.Contains(string(findings[0].Payload), `"attack_analysis":{"summary":"The search endpoint is injectable","title":"SQL Injection"}`))
	assert.Check(t, !strings.Contains(string(findings[1].Payload), "attack_analysis"))
}

func TestCollectIssuesAttachesCVEDetailsOncePerCVE(t *testing.T) {
	var cveRequests atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathIssuesExport:
			writeJSON(w, `[{"id":1,"type":"open_source","status":"open","cve_id":"CVE-2026-23864"},{"id":2,"type":"open_source","status":"snoozed","cve_id":"CVE-2026-23864"},{"id":3,"type":"sast","status":"open","cve_id":"CVE-2020-1"},{"id":4,"type":"open_source","status":"closed","cve_id":"CVE-2019-9"}]`)
		case fmt.Sprintf(pathCVEFormat, "CVE-2026-23864"):
			cveRequests.Add(1)
			writeJSON(w, `{"cve":"CVE-2026-23864","tldr":"Denial of service in Next.js"}`)
		case fmt.Sprintf(pathCVEFormat, "CVE-2020-1"):
			t.Error("cve details requested for a non open source issue")
		case fmt.Sprintf(pathCVEFormat, "CVE-2019-9"):
			t.Error("cve details requested for a closed issue")
		default:
			writeJSON(w, `[]`)
		}
	})

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	vulnerabilities := payloadSet(payloads, entityops.SchemaVulnerability.Name)
	assert.Assert(t, is.Len(vulnerabilities, 3))

	assert.Check(t, is.Equal(int32(1), cveRequests.Load()))
	assert.Check(t, is.Contains(string(vulnerabilities[1].Payload), `"tldr":"Denial of service in Next.js"`))
	assert.Check(t, !strings.Contains(string(vulnerabilities[2].Payload), "cve_details"))
}

func TestFetchEachSkipsFailedKeys(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		if r.URL.Path == fmt.Sprintf(pathCVEFormat, "AIKIDO-1") {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		writeJSON(w, `{"tldr":"details"}`)
	})

	records, err := client.fetchEach(context.Background(), []string{"AIKIDO-1", "CVE-2", "CVE-3"}, pathCVEFormat)
	assert.NilError(t, err)

	assert.Check(t, is.Len(records, 2))
	assert.Check(t, is.Equal(int32(3), requests.Load()))
}

func TestFetchEachStopsWhenRateLimited(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	})

	records, err := client.fetchEach(context.Background(), []string{"CVE-1", "CVE-2"}, pathCVEFormat)
	assert.NilError(t, err)

	assert.Check(t, is.Len(records, 0))
	assert.Check(t, is.Equal(int32(maxRateLimitRetries+1), requests.Load()))
}

func TestFetchEachStopsWithoutScope(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusForbidden)
	})

	records, err := client.fetchEach(context.Background(), []string{"CVE-1", "CVE-2"}, pathCVEFormat)
	assert.NilError(t, err)

	assert.Check(t, is.Len(records, 0))
	assert.Check(t, is.Equal(int32(1), requests.Load()))
}

func TestCollectIssuesWithoutAttackAnalysis(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathIssuesExport:
			writeJSON(w, `[{"id":7,"type":"ai_pentest","status":"open"}]`)
		case fmt.Sprintf(pathAttackAnalysisFormat, "7"):
			w.WriteHeader(http.StatusForbidden)
		default:
			writeJSON(w, `[]`)
		}
	})

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	assert.Check(t, is.Len(payloadSet(payloads, entityops.SchemaFinding.Name), 1))
}

func TestCollectIssuesAddsDataleaks(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathDomains:
			writeJSON(w, `[{"id":4,"domain":"app.example.com"},{"id":5,"domain":"down.example.com"}]`)
		case pathDomains + "/4/dataleaks":
			writeJSON(w, `{"dataleaks":[{"id":"a1","email":"jane@example.com","password":"********123","status":"OPEN"}]}`)
		case pathDomains + "/5/dataleaks":
			w.WriteHeader(http.StatusForbidden)
		default:
			writeJSON(w, `[]`)
		}
	})

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	findings := payloadSet(payloads, entityops.SchemaFinding.Name)
	assert.Assert(t, is.Len(findings, 1))

	assert.Check(t, is.Equal(findingVariantDataleak, findings[0].Variant))
	assert.Check(t, is.Equal("app.example.com", findings[0].Resource))
	assert.Check(t, is.Contains(string(findings[0].Payload), `"email":"jane@example.com"`))
	assert.Check(t, !strings.Contains(string(findings[0].Payload), "password"))
}

func TestCollectIssuesAttachesBulkDetails(t *testing.T) {
	var detailIDs string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathIssuesExport:
			writeJSON(w, `[{"id":1,"type":"open_source","status":"open"},{"id":2,"type":"sast"}]`)
		case pathIssueDetailsBulk:
			detailIDs = r.URL.Query().Get("issue_ids")
			writeJSON(w, `[{"id":1,"status":"ignored","cve_id_alias":"CVE-2020-7598","ignore_reasons":[{"kind":"manual_ignore","reason":"false positive"}]}]`)
		default:
			writeJSON(w, `[]`)
		}
	})

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	vulnerability := string(payloadSet(payloads, entityops.SchemaVulnerability.Name)[0].Payload)

	assert.Check(t, is.Equal("1,2", detailIDs))
	assert.Check(t, is.Contains(vulnerability, `"cve_id_alias":"CVE-2020-7598"`))
	assert.Check(t, is.Contains(vulnerability, `"ignore_reasons":[`))
	assert.Check(t, is.Contains(vulnerability, `"status":"open"`))
}

func TestCollectIssuesWithoutBulkDetails(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case pathIssuesExport:
			writeJSON(w, `[{"id":1,"type":"sast"}]`)
		case pathIssueDetailsBulk:
			w.WriteHeader(http.StatusForbidden)
		default:
			writeJSON(w, `[]`)
		}
	})

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual([]int64{1}, envelopeIDs(t, payloadSet(payloads, entityops.SchemaFinding.Name))))
}

func TestIssueDetailsBatchesIDs(t *testing.T) {
	var batches []int

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		batches = append(batches, len(strings.Split(r.URL.Query().Get("issue_ids"), ",")))
		writeJSON(w, `[]`)
	})

	ids := make([]string, 0, issueDetailBatchSize+1)
	for i := range issueDetailBatchSize + 1 {
		ids = append(ids, strconv.Itoa(i+1))
	}

	_, err := client.issueDetails(context.Background(), ids)
	assert.NilError(t, err)

	assert.Check(t, is.DeepEqual([]int{issueDetailBatchSize, 1}, batches))
}

func TestCollectIssuesAttachesGroups(t *testing.T) {
	statuses := map[string]int32{}

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathIssuesExport {
			writeJSON(w, `[{"id":1,"group_id":11,"type":"open_source"},{"id":2,"group_id":99,"type":"sast"}]`)

			return
		}

		if r.URL.Path != pathIssueGroups {
			writeJSON(w, `[]`)

			return
		}

		status := r.URL.Query().Get("filter_status")
		statuses[status]++

		if status == "closed" {
			writeJSON(w, `[{"id":11,"title":"lodash","description":"Prototype pollution","how_to_fix":"Upgrade lodash"}]`)

			return
		}

		writeJSON(w, `[]`)
	})

	payloads, err := CollectIssues(context.Background(), client)
	assert.NilError(t, err)

	vulnerabilities := payloadSet(payloads, entityops.SchemaVulnerability.Name)
	findings := payloadSet(payloads, entityops.SchemaFinding.Name)

	assert.Check(t, is.Contains(string(vulnerabilities[0].Payload), `"how_to_fix":"Upgrade lodash"`))
	assert.Check(t, !strings.Contains(string(findings[0].Payload), `"group":`))
	assert.Check(t, is.DeepEqual(map[string]int32{"open": 1, "closed": 1, "snoozed": 1, "ignored": 1}, statuses))
}

func TestIssueGroupsFollowsHasNextPage(t *testing.T) {
	var requests atomic.Int32

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("filter_status") != "open" {
			writeJSON(w, `[]`)

			return
		}

		requests.Add(1)

		if r.URL.Query().Get("page") == "0" {
			w.Header().Set(headerHasNextPage, "true")
			writeJSON(w, `[{"id":1}]`)

			return
		}

		w.Header().Set(headerHasNextPage, "false")
		writeJSON(w, `[{"id":2}]`)
	})

	groups, err := client.issueGroups(context.Background())
	assert.NilError(t, err)

	assert.Check(t, is.Len(groups, 2))
	assert.Check(t, is.Equal(int32(2), requests.Load()))
}

func TestCollectIssuesPropagatesGroupErrors(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pathIssueGroups {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		writeJSON(w, `[]`)
	})

	_, err := CollectIssues(context.Background(), client)
	assert.Check(t, is.ErrorIs(err, ErrIssueGroupsFetchFailed))
}
