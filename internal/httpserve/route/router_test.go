package route

import (
	"net/http"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/middleware/ratelimit"
)

func newTestRouter() *Router {
	return &Router{
		Echo: echo.New(),
		OAS:  &openapi3.T{Paths: openapi3.NewPaths()},
	}
}

func TestRateLimitedMiddlewares(t *testing.T) {
	base := []echo.MiddlewareFunc{func(next echo.HandlerFunc) echo.HandlerFunc { return next }}

	enabled := &ratelimit.Config{
		Enabled: true,
		Options: []ratelimit.RateOption{{Requests: 1, Window: time.Minute}},
	}
	dryRun := &ratelimit.Config{
		DryRun:  true,
		Options: []ratelimit.RateOption{{Requests: 1, Window: time.Minute}},
	}
	inactive := &ratelimit.Config{
		Options: []ratelimit.RateOption{{Requests: 1, Window: time.Minute}},
	}

	cases := []struct {
		name     string
		limit    *ratelimit.Config
		expected int
	}{
		{name: "nil leaves middleware untouched", limit: nil, expected: len(base)},
		{name: "inactive leaves middleware untouched", limit: inactive, expected: len(base)},
		{name: "enabled prepends a limiter", limit: enabled, expected: len(base) + 1},
		{name: "dry-run prepends a limiter", limit: dryRun, expected: len(base) + 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rateLimitedMiddlewares(Config{Middlewares: base, RateLimit: tc.limit})
			if len(got) != tc.expected {
				t.Fatalf("expected %d middleware, got %d", tc.expected, len(got))
			}
		})
	}
}

func TestAddEchoOnlyRoute(t *testing.T) {
	r := newTestRouter()
	rt := echo.Route{Path: "/e", Method: http.MethodGet, Handler: func(echo.Context) error { return nil }}
	if err := r.AddEchoOnlyRoute(rt); err != nil {
		t.Fatalf("add echo route failed: %v", err)
	}
	if r.OAS.Paths.Find("/e") != nil {
		t.Fatalf("should not add to spec")
	}
}
