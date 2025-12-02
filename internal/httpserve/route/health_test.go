package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/httpserve/handlers"
	"github.com/theopenlane/ent/entdb"
)

// TestHealthEndpointsShutdown verifies liveness and readiness handlers return ServiceUnavailable during shutdown
func TestHealthEndpointsShutdown(t *testing.T) {
	entdb.ResetShutdown()
	r := newTestRouter()
	r.Handler = &handlers.Handler{}

	require.NoError(t, registerLivenessHandler(r))
	require.NoError(t, registerReadinessHandler(r))

	ts := httptest.NewServer(r.Echo)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/livez")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/ready")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	entdb.BeginShutdown()

	resp, err = http.Get(ts.URL + "/livez")
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/ready")
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	resp.Body.Close()

	entdb.ResetShutdown()
}
