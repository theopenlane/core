package logx_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/shared/logx"
)

func TestMiddleware(t *testing.T) {
	t.Run("should not trigger error handler when HandleError is false", func(t *testing.T) {
		var called bool
		e := echo.New()

		e.HTTPErrorHandler = func(c echo.Context, err error) {
			called = true
			c.JSON(http.StatusInternalServerError, err.Error())
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		m := logx.LoggingMiddleware(logx.Config{})

		next := func(c echo.Context) error {
			return errors.New("error")
		}

		handler := m(next)
		err := handler(c)

		assert.Error(t, err, "should return error")
		assert.False(t, called, "should not call error handler")
	})

	t.Run("should trigger error handler when HandleError is true", func(t *testing.T) {
		var called bool
		e := echo.New()
		e.HTTPErrorHandler = func(c echo.Context, err error) {
			called = true

			c.JSON(http.StatusInternalServerError, err.Error())
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		m := logx.LoggingMiddleware(logx.Config{
			HandleError: true,
		})

		next := func(c echo.Context) error {
			return errors.New("error")
		}

		handler := m(next)
		err := handler(c)

		assert.Error(t, err, "should return error")
		assert.Truef(t, called, "should call error handler")
	})

	t.Run("should use enricher", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}

		l := logx.Configure(logx.LoggerConfig{
			Writer:   b,
			WithEcho: true,
		}).Echo
		m := logx.LoggingMiddleware(logx.Config{
			Logger: l,
			Enricher: func(c echo.Context, logger zerolog.Context) zerolog.Context {
				return logger.Str("test", "test")
			},
		})

		next := func(c echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"test":"test"`)
	})

	t.Run("should escalate log level for slow requests", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := logx.Configure(logx.LoggerConfig{
			Writer:   b,
			WithEcho: true,
		}).Echo
		l.SetLevel(log.INFO)
		m := logx.LoggingMiddleware(logx.Config{
			Logger:              l,
			RequestLatencyLimit: 5 * time.Millisecond,
			RequestLatencyLevel: zerolog.WarnLevel,
		})

		// Slow request should be logged at the escalated level
		next := func(c echo.Context) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		}
		handler := m(next)
		err := handler(c)
		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"level":"warn"`)
		assert.NotContains(t, str, `"level":"info"`)
	})

	t.Run("shouldn't escalate log level for fast requests", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := logx.Configure(logx.LoggerConfig{
			Writer:   b,
			WithEcho: true,
		}).Echo
		l.SetLevel(log.INFO)
		m := logx.LoggingMiddleware(logx.Config{
			Logger:              l,
			RequestLatencyLimit: 5 * time.Millisecond,
			RequestLatencyLevel: zerolog.WarnLevel,
		})

		// Fast request should be logged at the default level
		next := func(c echo.Context) error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"level":"info"`)
		assert.NotContains(t, str, `"level":"warn"`)
	})

	t.Run("should skip middleware before calling next handler when Skipper func returns true", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/skip", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := logx.Configure(logx.LoggerConfig{
			Writer:   b,
			WithEcho: true,
		}).Echo
		l.SetLevel(log.INFO)
		m := logx.LoggingMiddleware(logx.Config{
			Logger: l,
			Skipper: func(c echo.Context) bool {
				return c.Request().URL.Path == "/skip"
			},
		})

		next := func(c echo.Context) error {
			return nil
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Empty(t, str, "should not log anything")
	})

	t.Run("should skip middleware after calling next handler when AfterNextSkipper func returns true", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := logx.Configure(logx.LoggerConfig{
			Writer:   b,
			WithEcho: true,
		}).Echo
		l.SetLevel(log.INFO)
		m := logx.LoggingMiddleware(logx.Config{
			Logger: l,
			AfterNextSkipper: func(c echo.Context) bool {
				return c.Response().Status == http.StatusMovedPermanently
			},
		})

		next := func(c echo.Context) error {
			return c.Redirect(http.StatusMovedPermanently, "/other")
		}

		handler := m(next)
		err := handler(c)

		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Empty(t, str, "should not log anything")
	})

	t.Run("should log client ip headers when provided", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("True-Client-IP", "1.1.1.1")
		req.Header.Set("X-Forwarded-For", "1.1.1.1, 2.2.2.2")
		req.Header.Set("X-Real-IP", "3.3.3.3")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		b := &bytes.Buffer{}
		l := logx.Configure(logx.LoggerConfig{
			Writer:   b,
			WithEcho: true,
		}).Echo
		m := logx.LoggingMiddleware(logx.Config{
			Logger: l,
		})

		handler := m(func(c echo.Context) error {
			return nil
		})
		err := handler(c)
		assert.NoError(t, err, "should not return error")

		str := b.String()
		assert.Contains(t, str, `"true_client_ip":"1.1.1.1"`)
		assert.Contains(t, str, `"x_forwarded_for":"1.1.1.1, 2.2.2.2"`)
		assert.Contains(t, str, `"x_real_ip":"3.3.3.3"`)
	})
}
