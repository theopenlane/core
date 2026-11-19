package route

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/entdb"
	"github.com/theopenlane/core/v2/internal/httpserve/handlers"
)

// TestHealthEndpointsShutdown verifies liveness and readiness handlers return ServiceUnavailable during shutdown
func TestHealthEndpointsShutdown(t *testing.T) {
	entdb.ResetShutdown()
	r := newTestRouter()
	r.Handler = &handlers.Handler{}

	assert.NilError(t, registerLivenessHandler(r))
	assert.NilError(t, registerReadinessHandler(r))

	ts := httptest.NewServer(r.Echo)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/livez")
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(http.StatusOK, resp.StatusCode))
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/ready")
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(http.StatusOK, resp.StatusCode))
	resp.Body.Close()

	entdb.BeginShutdown()

	resp, err = http.Get(ts.URL + "/livez")
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(http.StatusServiceUnavailable, resp.StatusCode))
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/ready")
	assert.NilError(t, err)
	assert.Assert(t, is.Equal(http.StatusServiceUnavailable, resp.StatusCode))
	resp.Body.Close()

	entdb.ResetShutdown()
}

// errCheckFailed is a stand-in dependency error
var errCheckFailed = errors.New("dependency unavailable")

// TestReadinessChecks verifies the readiness handler runs its checks concurrently and reports failures
func TestReadinessChecks(t *testing.T) {
	entdb.ResetShutdown()

	pass := func(context.Context) error { return nil }

	firstStarted := make(chan struct{})
	secondStarted := make(chan struct{})

	// each check passes only once the other has started
	waitFor := func(started chan<- struct{}, other <-chan struct{}) handlers.CheckFunc {
		return func(ctx context.Context) error {
			if _, ok := ctx.Deadline(); !ok {
				return errCheckFailed
			}

			close(started)

			select {
			case <-other:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	tests := []struct {
		name   string
		checks map[string]handlers.CheckFunc
		status int
	}{
		{
			name:   "every check passes",
			checks: map[string]handlers.CheckFunc{"first": pass, "second": pass},
			status: http.StatusOK,
		},
		{
			name:   "a check fails",
			checks: map[string]handlers.CheckFunc{"healthy": pass, "broken": func(context.Context) error { return errCheckFailed }},
			status: http.StatusServiceUnavailable,
		},
		{
			name: "checks run concurrently under a deadline",
			checks: map[string]handlers.CheckFunc{
				"first":  waitFor(firstStarted, secondStarted),
				"second": waitFor(secondStarted, firstStarted),
			},
			status: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestRouter()
			r.Handler = &handlers.Handler{}

			for name, check := range tc.checks {
				r.Handler.AddReadinessCheck(name, check)
			}

			assert.NilError(t, registerReadinessHandler(r))

			ts := httptest.NewServer(r.Echo)
			defer ts.Close()

			resp, err := http.Get(ts.URL + "/ready")
			assert.NilError(t, err)
			assert.Check(t, is.Equal(tc.status, resp.StatusCode))
			resp.Body.Close()
		})
	}
}
