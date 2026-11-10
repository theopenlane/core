package route

import (
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
