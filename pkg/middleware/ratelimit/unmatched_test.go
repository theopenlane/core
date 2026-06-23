package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/middleware/ratelimit"
)

// newUnmatchedEcho builds an echo instance with a single registered route and the unmatched limiter applied
func newUnmatchedEcho(config *ratelimit.Config) *echo.Echo {
	e := echo.New()
	e.Use(ratelimit.UnmatchedRouteLimiterWithConfig(config))
	e.GET("/healthz", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	return e
}

func TestUnmatchedRouteLimiterBlocksEnumerationAcrossDistinctPaths(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		Enabled:              true,
		SendRetryAfterHeader: true,
		Options: []ratelimit.RateOption{
			{
				Requests: 2,
				Window:   time.Minute,
			},
		},
	}

	e := newUnmatchedEcho(config)

	// distinct unregistered paths from the same source must share one bucket so a scanner cannot evade the limit
	paths := []string{"/wp-login.php", "/.env", "/admin/config"}

	for i, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = "203.0.113.5:5000"
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		switch {
		case i < 2 && rec.Code != http.StatusNotFound:
			t.Fatalf("expected status %d for unmatched path %q, got %d", http.StatusNotFound, path, rec.Code)
		case i == 2 && rec.Code != http.StatusTooManyRequests:
			t.Fatalf("expected status %d once the enumeration limit is exceeded, got %d", http.StatusTooManyRequests, rec.Code)
		}
	}
}

func TestUnmatchedRouteLimiterIgnoresMatchedRoutes(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		Enabled: true,
		Options: []ratelimit.RateOption{
			{
				Requests: 1,
				Window:   time.Minute,
			},
		},
	}

	e := newUnmatchedEcho(config)

	for i := range 3 {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		req.RemoteAddr = "203.0.113.5:5000"
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected matched route to bypass the unmatched limiter with 200, got %d on request %d", rec.Code, i+1)
		}
	}
}

func TestUnmatchedRouteLimiterIsolatesByIP(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		Enabled: true,
		Options: []ratelimit.RateOption{
			{
				Requests: 1,
				Window:   time.Minute,
			},
		},
	}

	e := newUnmatchedEcho(config)

	request := func(remoteAddr string) int {
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		return rec.Code
	}

	if code := request("198.51.100.1:4000"); code != http.StatusNotFound {
		t.Fatalf("expected first request from IP A to return 404, got %d", code)
	}

	if code := request("198.51.100.1:4000"); code != http.StatusTooManyRequests {
		t.Fatalf("expected second request from IP A to be limited, got %d", code)
	}

	if code := request("198.51.100.2:4000"); code != http.StatusNotFound {
		t.Fatalf("expected request from distinct IP B to be unaffected, got %d", code)
	}
}

func TestUnmatchedRouteLimiterOmitsRetryAfterHeader(t *testing.T) {
	t.Parallel()

	// SendRetryAfterHeader is requested but must be suppressed: the path does not exist, so a Retry-After would
	// only nudge clients into retrying a request that can never succeed
	config := &ratelimit.Config{
		Enabled:              true,
		SendRetryAfterHeader: true,
		Options:              []ratelimit.RateOption{{Requests: 1, Window: time.Minute}},
	}

	e := newUnmatchedEcho(config)

	var (
		lastCode   int
		retryAfter string
	)

	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
		req.RemoteAddr = "203.0.113.9:5000"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		lastCode = rec.Code
		retryAfter = rec.Header().Get(echo.HeaderRetryAfter)
	}

	if lastCode != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be limited with %d, got %d", http.StatusTooManyRequests, lastCode)
	}

	if retryAfter != "" {
		t.Fatalf("expected no Retry-After header on an unmatched limit, got %q", retryAfter)
	}
}

func TestUnmatchedRouteLimiterKeysOnCFConnectingIPWithRemoteAddrFallback(t *testing.T) {
	t.Parallel()

	config := &ratelimit.Config{
		Enabled: true,
		Options: []ratelimit.RateOption{{Requests: 1, Window: time.Minute}},
	}

	e := newUnmatchedEcho(config)

	request := func(cfIP, remoteAddr string) int {
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		req.RemoteAddr = remoteAddr

		if cfIP != "" {
			req.Header.Set("CF-Connecting-IP", cfIP)
		}

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		return rec.Code
	}

	// CF-Connecting-IP is the key: the same client IP behind two different proxy sockets shares one bucket
	if code := request("198.51.100.7", "203.0.113.1:1000"); code != http.StatusNotFound {
		t.Fatalf("expected first CF-keyed request to return 404, got %d", code)
	}

	if code := request("198.51.100.7", "203.0.113.2:2000"); code != http.StatusTooManyRequests {
		t.Fatalf("expected same CF-Connecting-IP from a different socket to be limited, got %d", code)
	}

	// fallback: with no CF-Connecting-IP header present the limiter keys on RemoteAddr, an independent bucket
	if code := request("", "192.0.2.50:7000"); code != http.StatusNotFound {
		t.Fatalf("expected fallback request keyed on RemoteAddr to return 404, got %d", code)
	}

	if code := request("", "192.0.2.50:7000"); code != http.StatusTooManyRequests {
		t.Fatalf("expected second fallback request from the same RemoteAddr to be limited, got %d", code)
	}

	if code := request("", "192.0.2.51:7000"); code != http.StatusNotFound {
		t.Fatalf("expected fallback request from a different RemoteAddr to be unaffected, got %d", code)
	}
}
