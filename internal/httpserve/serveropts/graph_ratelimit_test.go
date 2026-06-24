package serveropts

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/theopenlane/echox"

	coreconfig "github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/pkg/middleware/ratelimit"
)

func TestGraphRateLimitConfig(t *testing.T) {
	t.Parallel()

	in := ratelimit.Config{
		Enabled:              true,
		SendRetryAfterHeader: true,
		DryRun:               true,
	}

	cfg := graphRateLimitConfig(in)

	if !cfg.Enabled {
		t.Fatalf("expected graph rate limit config to be enabled")
	}

	if !cfg.SendRetryAfterHeader {
		t.Fatalf("expected graph rate limit config to send Retry-After")
	}

	if !cfg.DryRun {
		t.Fatalf("expected graph rate limit config to inherit DryRun from the base config")
	}

	if len(cfg.Options) != 1 || cfg.Options[0].Requests != graphRateLimitRequests || cfg.Options[0].Window != graphRateLimitWindow {
		t.Fatalf("unexpected graph rate limit window: %+v", cfg.Options)
	}

	if len(cfg.Headers) != 3 || cfg.Headers[0] != "CF-Connecting-IP" || cfg.Headers[1] != "True-Client-IP" || cfg.Headers[2] != "RemoteAddr" {
		t.Fatalf("expected CF-Connecting-IP and True-Client-IP preferred with RemoteAddr fallback, got %v", cfg.Headers)
	}
}

func TestWithGraphRateLimiterEnforcesAheadOfGraphMiddleware(t *testing.T) {
	t.Parallel()

	// a downstream marker stands in for the real graph middleware (auth, etc.) so we can prove the limiter runs first
	var downstreamCalls int

	marker := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			downstreamCalls++

			return next(c)
		}
	}

	so := &ServerOptions{Config: serverconfig.Config{Settings: coreconfig.Config{
		Ratelimit: ratelimit.Config{
			Enabled:              true,
			SendRetryAfterHeader: true,
		},
	}}}
	so.Config.GraphMiddleware = append(so.Config.GraphMiddleware, marker)

	WithGraphRateLimiter().apply(so)

	if len(so.Config.GraphMiddleware) != 2 {
		t.Fatalf("expected the limiter to be added ahead of existing graph middleware, got %d entries", len(so.Config.GraphMiddleware))
	}

	e := echo.New()
	grp := e.Group("", so.Config.GraphMiddleware...)
	grp.POST("/query", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// vary RemoteAddr while holding CF-Connecting-IP constant: if the limiter keyed on RemoteAddr it would never trip,
	// so tripping here proves it keys on CF-Connecting-IP
	request := func(cfIP string, i int) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/query", nil)
		req.RemoteAddr = fmt.Sprintf("203.0.113.%d:5000", i%250)

		if cfIP != "" {
			req.Header.Set("CF-Connecting-IP", cfIP)
		}

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		return rec
	}

	limit := int(graphRateLimitRequests)

	for i := range limit {
		if code := request("198.51.100.7", i).Code; code != http.StatusOK {
			t.Fatalf("expected request %d under the limit to succeed, got %d", i+1, code)
		}
	}

	blocked := request("198.51.100.7", limit)
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("expected request over the limit to be rate limited, got %d", blocked.Code)
	}

	if blocked.Header().Get(echo.HeaderRetryAfter) == "" {
		t.Fatalf("expected Retry-After header on the rate limited graph response")
	}

	// the limiter must short-circuit ahead of the downstream graph middleware: it ran for every allowed request but not the blocked one
	if downstreamCalls != limit {
		t.Fatalf("expected downstream middleware to run %d times, got %d", limit, downstreamCalls)
	}

	// a distinct client (different CF-Connecting-IP) is its own bucket and unaffected
	if code := request("198.51.100.8", 0).Code; code != http.StatusOK {
		t.Fatalf("expected a distinct CF-Connecting-IP to be unaffected, got %d", code)
	}

	// with no CF-Connecting-IP header the limiter falls back to RemoteAddr, an independent bucket
	noCF := httptest.NewRequest(http.MethodPost, "/query", nil)
	noCF.RemoteAddr = "192.0.2.77:9000"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, noCF)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected RemoteAddr-keyed fallback request to succeed, got %d", rec.Code)
	}
}
