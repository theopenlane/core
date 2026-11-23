package vulnintel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"osv.dev/bindings/go/osvdev"
)

const osvAdvisory = `{
	"id": "GHSA-253c-mchw-3w2r",
	"summary": "Regular expression denial of service in zod",
	"details": "The emoji validator backtracks exponentially.",
	"published": "2026-10-21T17:59:27Z",
	"severity": [
		{"type": "CVSS_V4", "score": "CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:N/VI:N/VA:L/SC:N/SI:N/SA:N"},
		{"type": "CVSS_V3", "score": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:L"}
	],
	"references": [{"type": "ADVISORY", "url": "https://github.com/advisories/GHSA-253c-mchw-3w2r"}],
	"affected": [{"ranges": [{"type": "SEMVER", "events": [{"introduced": "0"}, {"fixed": "3.25.2"}]}]}]
}`

func newTestOSVClient(t *testing.T, handler http.HandlerFunc) *osvClient {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := newOSVClient()
	client.client.BaseHostURL = server.URL
	client.client.HTTPClient = server.Client()
	client.client.Config.MaxRetryAttempts = 1

	return client
}

func TestOSVLookup(t *testing.T) {
	client := newTestOSVClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != osvdev.GetEndpoint+"/GHSA-253c-mchw-3w2r" {
			w.WriteHeader(http.StatusNotFound)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(osvAdvisory))
	})

	record, found, err := client.lookup(context.Background(), "GHSA-253c-mchw-3w2r")
	assert.NilError(t, err)
	assert.Assert(t, found)

	assert.Check(t, is.Equal("Regular expression denial of service in zod", record.Summary))
	assert.Check(t, is.Equal("CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:L", record.vector()))
	assert.Check(t, is.Equal(time.Date(2026, 10, 21, 17, 59, 27, 0, time.UTC), record.publishedAt()))
	assert.Check(t, is.DeepEqual([]string{"https://osv.dev/vulnerability/GHSA-253c-mchw-3w2r", "https://github.com/advisories/GHSA-253c-mchw-3w2r"}, record.referenceURLs()))

	_, found, err = client.lookup(context.Background(), "AIKIDO-2026-161437")
	assert.NilError(t, err)
	assert.Check(t, !found)
}

func TestOSVLookupFailure(t *testing.T) {
	client := newTestOSVClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, _, err := client.lookup(context.Background(), "CVE-2026-1")
	assert.Check(t, is.ErrorIs(err, ErrOSVRequest))
}

func TestOSVLookupSendsID(t *testing.T) {
	client := newTestOSVClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Check(t, is.Equal(osvdev.GetEndpoint+"/CVE-2026-23864", r.URL.Path))
		w.WriteHeader(http.StatusNotFound)
	})

	_, found, err := client.lookup(context.Background(), " CVE-2026-23864 ")
	assert.NilError(t, err)
	assert.Check(t, !found)
}

func TestOSVVectorFallsBackToV4(t *testing.T) {
	record := osvRecord{Severity: []osvSeverity{{Type: severityCVSSV4, Score: "CVSS:4.0/AV:N"}}}

	assert.Check(t, is.Equal("CVSS:4.0/AV:N", record.vector()))
}

func TestCacheEntryFresh(t *testing.T) {
	now := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)

	assert.Check(t, cacheEntry{FetchedAt: now.Add(-6 * 24 * time.Hour), Found: true}.fresh(now))
	assert.Check(t, !cacheEntry{FetchedAt: now.Add(-8 * 24 * time.Hour), Found: true}.fresh(now))
	assert.Check(t, !cacheEntry{FetchedAt: now.Add(-25 * time.Hour)}.fresh(now))
}
