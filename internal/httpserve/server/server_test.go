package server_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	echo_log "github.com/labstack/gommon/log"
	"github.com/mcuadros/go-defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"

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

type handlerRoundTripper struct {
	handler http.Handler
}

func (rt *handlerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	rt.handler.ServeHTTP(rec, req)
	return rec.Result(), nil
}

type httpslingTestClient struct {
	requester *httpsling.Requester
	jar       http.CookieJar
	baseURL   *url.URL
}

func newHTTPSlingTestClient(handler http.Handler) (*httpslingTestClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse("https://router.test")
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Jar:       jar,
		Transport: &handlerRoundTripper{handler: handler},
	}

	requester, err := httpsling.New(
		httpsling.WithHTTPClient(httpClient),
		httpsling.URL(baseURL.String()),
	)
	if err != nil {
		return nil, err
	}

	return &httpslingTestClient{
		requester: requester,
		jar:       jar,
		baseURL:   baseURL,
	}, nil
}

func (c *httpslingTestClient) cookieValue(name string) string {
	for _, ck := range c.jar.Cookies(c.baseURL) {
		if ck.Name == name {
			return ck.Value
		}
	}

	return ""
}

func (c *httpslingTestClient) get(ctx context.Context, path string) (*http.Response, error) {
	return c.requester.ReceiveWithContext(ctx, nil, httpsling.Get(path))
}

func (c *httpslingTestClient) post(ctx context.Context, path string, header http.Header) (*http.Response, error) {
	opts := []httpsling.Option{httpsling.Post(path)}
	if header != nil {
		opts = append(opts, httpsling.HeadersFromValues(header))
	}

	return c.requester.ReceiveWithContext(ctx, nil, opts...)
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

	client, err := newHTTPSlingTestClient(srv.Router.Echo)
	require.NoError(t, err)
	ctx := context.Background()

	// first request should set csrf cookie
	resp, err := client.get(ctx, "/livez")
	assert.NoError(t, err)
	resp.Body.Close()
	token := client.cookieValue("ol.csrf-token")
	assert.NotEmpty(t, token)

	// missing header should return 400
	resp, err = client.post(ctx, "/test", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()

	// include token header
	headers := make(http.Header)
	headers.Set("X-CSRF-Token", token)
	resp, err = client.post(ctx, "/test", headers)
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

	client, err := newHTTPSlingTestClient(srv.Router.Echo)
	require.NoError(t, err)
	ctx := context.Background()

	// first request should not set csrf cookie since CSRF is disabled
	resp, err := client.get(ctx, "/livez")
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Empty(t, client.cookieValue("ol.csrf-token"))

	// missing header should return 200 OK since CSRF is disabled
	resp, err = client.post(ctx, "/test", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// include token header, no effect since CSRF is disabled
	headers := make(http.Header)
	headers.Set("X-CSRF-Token", "unused")
	resp, err = client.post(ctx, "/test", headers)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestRecoverMiddleware(t *testing.T) {
	t.Parallel()

	// create echo server with the configured middleware
	e := server.ConfigureEcho(server.LogConfig{
		PrettyLog: false,
		LogLevel:  echo_log.INFO,
	})

	// add a handler that panics
	e.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})

	// add a normal handler to verify server still works
	e.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	client, err := newHTTPSlingTestClient(e)
	require.NoError(t, err)
	ctx := context.Background()

	// verify panic is recovered and returns 500
	resp, err := client.get(ctx, "/panic")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// verify server still works after panic
	resp2, err := client.get(ctx, "/ok")
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}
