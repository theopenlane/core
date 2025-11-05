package server_test

import (
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mcuadros/go-defaults"

	echo "github.com/theopenlane/echox"

	"net/http/httptest"

	"github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
	server "github.com/theopenlane/core/internal/httpserve/server"
	"github.com/theopenlane/core/internal/httpserve/serveropts"
)

// testHandler implements server.handler and exposes a simple POST endpoint
// used to validate CSRF behavior
type testHandler struct{}

func (testHandler) Routes(g *echo.Group) {
	g.POST("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
}

func TestServerCSRF(t *testing.T) {
	// create base configuration with defaults
	cfg := config.Config{}
	defaults.SetDefaults(&cfg)
	cfg.Server.Listen = "localhost:0"
	cfg.Server.MetricsPort = ":0"
	cfg.Server.CSRFProtection.Enabled = true
	cfg.Server.CSRFProtection.Secure = false
	cfg.ObjectStorage.Enabled = false

	so := &serveropts.ServerOptions{
		Config: serverconfig.Config{Settings: cfg},
	}

	// apply middleware options
	so.AddServerOptions(serveropts.WithCSRF())

	srv, err := server.NewServer(so.Config)
	assert.NoError(t, err)

	// apply middleware to echo instance
	for _, m := range so.Config.Handler.AdditionalMiddleware {
		if m != nil {
			srv.Router.Echo.Use(m)
		}
	}

	// manually register /livez and test route
	srv.Router.Echo.GET("/livez", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	testHandler{}.Routes(srv.Router.Echo.Group(""))

	ts := httptest.NewServer(srv.Router.Echo)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)

	client := &http.Client{Jar: jar}

	// first request should set csrf cookie
	resp, err := client.Get(ts.URL + "/livez")
	assert.NoError(t, err)
	resp.Body.Close()
	var token string
	for _, ck := range jar.Cookies(resp.Request.URL) {
		if ck.Name == "ol.csrf-token" {
			token = ck.Value
		}
	}
	assert.NotEmpty(t, token)

	// missing header should return 400
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/test", nil)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// include token header
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/test", nil)
	assert.NoError(t, err)
	req.Header.Set("X-CSRF-Token", token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestServerDisabledCSRF(t *testing.T) {
	t.Parallel()

	// create base configuration with defaults
	cfg := config.Config{}
	defaults.SetDefaults(&cfg)
	cfg.Server.Listen = "localhost:0"
	cfg.Server.MetricsPort = ":0"
	cfg.Server.CSRFProtection.Enabled = false
	cfg.ObjectStorage.Enabled = false

	so := &serveropts.ServerOptions{
		Config: serverconfig.Config{Settings: cfg},
	}

	// apply middleware options
	so.AddServerOptions(serveropts.WithCSRF())

	srv, err := server.NewServer(so.Config)
	assert.NoError(t, err)

	// apply middleware to echo instance
	for _, m := range so.Config.Handler.AdditionalMiddleware {
		if m != nil {
			srv.Router.Echo.Use(m)
		}
	}

	// manually register /livez and test route
	srv.Router.Echo.GET("/livez", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	testHandler{}.Routes(srv.Router.Echo.Group(""))

	ts := httptest.NewServer(srv.Router.Echo)
	defer ts.Close()

	jar, err := cookiejar.New(nil)
	assert.NoError(t, err)

	client := &http.Client{Jar: jar}

	// first request should not set csrf cookie since CSRF is disabled
	resp, err := client.Get(ts.URL + "/livez")
	assert.NoError(t, err)
	resp.Body.Close()
	var token string
	for _, ck := range jar.Cookies(resp.Request.URL) {
		if ck.Name == "ol.csrf-token" {
			token = ck.Value
		}
	}
	assert.Empty(t, token)

	// missing header should return 200 OK since CSRF is disabled
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/test", nil)
	assert.NoError(t, err)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// include token header, no effect since CSRF is disabled
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/test", nil)
	assert.NoError(t, err)
	req.Header.Set("X-CSRF-Token", token)
	resp, err = client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
