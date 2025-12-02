package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/shared/middleware/ratelimit"
)

func TestRateLimiterWithConfigBlocksAfterLimit(t *testing.T) {
	t.Parallel()

	e := echo.New()

	config := &ratelimit.Config{
		Enabled:              true,
		Headers:              []string{"True-Client-IP"},
		SendRetryAfterHeader: true,
		Options: []ratelimit.RateOption{
			{
				Requests: 2,
				Window:   time.Minute,
			},
		},
	}

	e.Use(ratelimit.RateLimiterWithConfig(config))
	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("True-Client-IP", "10.0.0.1")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		switch {
		case i < 2 && rec.Code != http.StatusOK:
			t.Fatalf("expected status %d but got %d on request %d", http.StatusOK, rec.Code, i+1)
		case i == 2 && rec.Code != http.StatusTooManyRequests:
			t.Fatalf("expected status %d but got %d on request %d", http.StatusTooManyRequests, rec.Code, i+1)
		}

		if i == 2 && rec.Header().Get(echo.HeaderRetryAfter) == "" {
			t.Fatalf("expected Retry-After header to be present on rate limited response")
		}
	}
}

func TestRateLimiterWithConfigIncludesPathWhenConfigured(t *testing.T) {
	t.Parallel()

	e := echo.New()

	config := &ratelimit.Config{
		Enabled:     true,
		Headers:     []string{"True-Client-IP"},
		IncludePath: true,
		DenyStatus:  http.StatusTooManyRequests,
		DenyMessage: "Too many requests",
		Options: []ratelimit.RateOption{
			{
				Requests: 1,
				Window:   time.Minute,
			},
		},
	}

	e.Use(ratelimit.RateLimiterWithConfig(config))
	e.GET("/alpha", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/beta", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	request := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("True-Client-IP", "10.0.0.2")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	if status := request("/alpha").Code; status != http.StatusOK {
		t.Fatalf("expected request to /alpha to succeed, got status %d", status)
	}

	if status := request("/beta").Code; status != http.StatusOK {
		t.Fatalf("expected request to /beta to succeed, got status %d", status)
	}

	if status := request("/alpha").Code; status != http.StatusTooManyRequests {
		t.Fatalf("expected second request to /alpha to be rate limited, got status %d", status)
	}
}

func TestRateLimiterWithDryRunAllowsRequests(t *testing.T) {
	t.Parallel()

	e := echo.New()

	config := &ratelimit.Config{
		Enabled:              true,
		DryRun:               true,
		Headers:              []string{"True-Client-IP"},
		SendRetryAfterHeader: true,
		Options: []ratelimit.RateOption{
			{
				Requests: 1,
				Window:   time.Minute,
			},
		},
	}

	e.Use(ratelimit.RateLimiterWithConfig(config))
	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("True-Client-IP", "192.0.2.1")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected dry-run request to succeed with 200, got status %d on request %d", rec.Code, i+1)
		}

		if rec.Header().Get(echo.HeaderRetryAfter) != "" {
			t.Fatalf("expected Retry-After header to be omitted during dry-run")
		}
	}
}
