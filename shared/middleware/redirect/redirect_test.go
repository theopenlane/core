package redirect_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/shared/middleware/redirect"
)

func TesetRedirectWithConfig(t *testing.T) {
	var testCases = []struct {
		name           string
		givenCode      int
		givenSkipFunc  func(c echo.Context) bool
		givenLocation  string
		expectLocation string
		expectedCode   int
	}{
		{
			name:           "usual redirect",
			givenLocation:  "/.well-known/change-password",
			expectLocation: "/v1/forgot-password",
			expectedCode:   http.StatusMovedPermanently,
		},
		{
			name: "redirect is skipped",
			givenSkipFunc: func(c echo.Context) bool {
				return true
			},
			givenLocation:  "https://api.theopenlane.io",
			expectLocation: "",
			expectedCode:   http.StatusOK,
		},
		{
			name:           "redirect with custom status code",
			givenCode:      http.StatusSeeOther,
			givenLocation:  "/meowmeow",
			expectLocation: "/v1/meowmeow",
			expectedCode:   http.StatusSeeOther,
		},
		{
			name:           "redirect with int code",
			givenCode:      302,
			givenLocation:  "/hello",
			expectLocation: "/v1/goodbye",
			expectedCode:   http.StatusFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			e.Use(redirect.NewWithConfig(redirect.Config{
				Redirects: map[string]string{
					tc.givenLocation: tc.expectLocation,
				},
				Code:    tc.givenCode,
				Skipper: tc.givenSkipFunc,
			}))

			e.GET("/new", func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tc.givenLocation, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectLocation, req.URL.Path)
			assert.Equal(t, tc.expectedCode, rec.Code)
			assert.Equal(t, tc.expectLocation, rec.Header().Get(echo.HeaderLocation))
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	var testCases = []struct {
		name           string
		givenConfig    redirect.Config
		givenSkipFunc  func(c echo.Context) bool
		givenPath      string
		expectRedirect bool
		expectCode     int
		expectTarget   string
	}{
		{
			name: "redirect exists for given path",
			givenConfig: redirect.Config{
				Redirects: map[string]string{
					"/old": "/new",
				},
				Code: http.StatusMovedPermanently,
			},
			givenPath:      "/old",
			expectRedirect: true,
			expectCode:     http.StatusMovedPermanently,
			expectTarget:   "/new",
		},
		{
			name: "no redirect exists for given path",
			givenConfig: redirect.Config{
				Redirects: map[string]string{
					"/old": "/new",
				},
				Code: http.StatusMovedPermanently,
			},
			givenPath:      "/other",
			expectRedirect: false,
			expectCode:     http.StatusNotFound,
			expectTarget:   "",
		},
		{
			name: "skip function returns true",
			givenConfig: redirect.Config{
				Redirects: map[string]string{
					"/old": "/new",
				},
				Code:    http.StatusMovedPermanently,
				Skipper: func(c echo.Context) bool { return true },
			},
			givenPath:      "/old",
			expectRedirect: false,
			expectCode:     http.StatusNotFound,
			expectTarget:   "",
		},
		{
			name: "skip function returns false",
			givenConfig: redirect.Config{
				Redirects: map[string]string{
					"/old": "/new",
				},
				Code:    http.StatusMovedPermanently,
				Skipper: func(c echo.Context) bool { return false },
			},
			givenPath:      "/old",
			expectRedirect: true,
			expectCode:     http.StatusMovedPermanently,
			expectTarget:   "/new",
		},
		{
			name: "default code is used",
			givenConfig: redirect.Config{
				Redirects: map[string]string{
					"/old": "/new",
				},
			},
			givenPath:      "/old",
			expectRedirect: true,
			expectCode:     http.StatusMovedPermanently,
			expectTarget:   "/new",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			middleware := redirect.NewWithConfig(tc.givenConfig)

			e.Use(middleware)
			e.GET("/new", func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tc.givenPath, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if tc.expectRedirect {
				assert.Equal(t, tc.expectCode, rec.Code)
				assert.Equal(t, tc.expectTarget, rec.Header().Get(echo.HeaderLocation))
			} else {
				assert.Equal(t, tc.expectCode, rec.Code)
				assert.Empty(t, rec.Header().Get(echo.HeaderLocation))
			}
		})
	}
}
