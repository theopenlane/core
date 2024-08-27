package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/pkg/auth"
	"github.com/theopenlane/core/pkg/sessions"
)

var happy = "happy path"

func TestGetAccessToken(t *testing.T) {
	testAccessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguZGF0dW0ubmV0IiwiYXVkIjoiaHR0cHM6Ly9kYXR1bS5uZXQiLCJzdWIiOiJVMVdNNHVGLTNxcGRsLWRtS0lISjQiLCJpYXQiOjE0NTg3ODU3OTYsImV4cCI6MTQ1ODg3MjE5Nn0.oXIjG4PauoHXEmZRDKRE018bkMv9rdZTjn563ujUh6o" // nolint:gosec

	var bear = "Bearer %s"

	tests := []struct {
		name        string
		headerKey   string
		headerValue string
		cookie      *http.Cookie
		wantTks     string
		wantErr     bool
		err         error
	}{
		{
			name:        "happy path from header",
			headerKey:   auth.Authorization,
			headerValue: fmt.Sprintf(bear, testAccessToken),
			wantTks:     testAccessToken,
			wantErr:     false,
			err:         nil,
		},
		{
			name:        "happy path from cookie",
			headerKey:   "",
			headerValue: "",
			cookie: &http.Cookie{
				Name:  auth.AccessTokenCookie,
				Value: testAccessToken,
			},
			wantTks: testAccessToken,
			wantErr: false,
			err:     nil,
		},
		{
			name:        "happy path cookie and header set",
			headerKey:   auth.Authorization,
			headerValue: fmt.Sprintf(bear, testAccessToken),
			cookie: &http.Cookie{
				Name:  auth.AccessTokenCookie,
				Value: "access_token_from_cookie", // to confirm we get the one from the header, this will be a different value
			},
			wantTks: testAccessToken,
			wantErr: false,
			err:     nil,
		},
		{
			name:        "no authorization header",
			headerKey:   "Hackorz",
			headerValue: fmt.Sprintf(bear, testAccessToken),
			wantTks:     "",
			wantErr:     true,
			err:         auth.ErrNoAuthorization,
		},
		{
			name:        "no bearer",
			headerKey:   auth.Authorization,
			headerValue: fmt.Sprintf("Bear %s", testAccessToken),
			wantTks:     "",
			wantErr:     true,
			err:         auth.ErrParseBearer,
		},
		{
			name:        "bearer not spaced",
			headerKey:   auth.Authorization,
			headerValue: fmt.Sprintf("Bearer%s", testAccessToken),
			wantTks:     "",
			wantErr:     true,
			err:         auth.ErrParseBearer,
		},
		{
			name:        "cookie set, but no access token cookie",
			headerKey:   "",
			headerValue: "",
			cookie: &http.Cookie{
				Name:  "not_an_access_token",
				Value: testAccessToken,
			},
			wantTks: "",
			wantErr: true,
			err:     auth.ErrNoAuthorization,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			req := &http.Request{
				Header: http.Header{},
			}

			// Add header/cookies
			req.Header.Add(tc.headerKey, tc.headerValue)

			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			recorder := httptest.NewRecorder()
			res := &echo.Response{
				Writer: recorder,
			}

			ctx := e.NewContext(req, res)

			gotTks, err := auth.GetAccessToken(ctx)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
				assert.Empty(t, gotTks)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantTks, gotTks)
		})
	}
}

func TestGetRefreshToken(t *testing.T) {
	testRefreshToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguZGF0dW0ubmV0IiwiYXVkIjoiaHR0cHM6Ly9kYXR1bS5uZXQiLCJzdWIiOiJVMVdNNHVGLTNxcGRsLWRtS0lISjQiLCJpYXQiOjE0NTg3ODU3OTYsImV4cCI6MTQ1ODg3MjE5Nn0.oXIjG4PauoHXEmZRDKRE018bkMv9rdZTjn563ujUh6o" //nolint:gosec
	tests := []struct {
		name    string
		cookie  *http.Cookie
		wantTks string
		wantErr bool
		err     error
	}{

		{
			name: "happy path from cookie",
			cookie: &http.Cookie{
				Name:  auth.RefreshTokenCookie,
				Value: testRefreshToken,
			},
			wantTks: testRefreshToken,
			wantErr: false,
			err:     nil,
		},
		{
			name:    "no cookie",
			wantTks: "",
			wantErr: true,
			err:     auth.ErrNoRefreshToken,
		},
		{
			name: "cookie present, but no refresh cookie",
			cookie: &http.Cookie{
				Name:  "another_cookie",
				Value: testRefreshToken,
			},
			wantTks: "",
			wantErr: true,
			err:     auth.ErrNoRefreshToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			req := &http.Request{
				Header: http.Header{},
			}

			// Add cookies
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			recorder := httptest.NewRecorder()
			res := &echo.Response{
				Writer: recorder,
			}

			ctx := e.NewContext(req, res)

			gotTks, err := auth.GetRefreshToken(ctx)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
				assert.Empty(t, gotTks)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantTks, gotTks)
		})
	}
}

func TestSetAuthCookies(t *testing.T) {
	validCtx, err := auth.NewTestEchoContextWithValidUser("funks")
	if err != nil {
		t.Fatal()
	}

	testAccessToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguZGF0dW0ubmV0IiwiYXVkIjoiaHR0cHM6Ly9kYXR1bS5uZXQiLCJzdWIiOiJVMVdNNHVGLTNxcGRsLWRtS0lISjQiLCJpYXQiOjE3MDE5ODc2NDYsImV4cCI6MzMyNTg4OTY0NDZ9.y51S2D9qMHLRixj230YZbvQZyhWzDOQ2RPbyJmnEYXA"  //nolint:gosec
	testRefreshToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJodHRwczovL2F1dGguZGF0dW0ubmV0IiwiYXVkIjoiaHR0cHM6Ly9kYXR1bS5uZXQiLCJzdWIiOiJVMVdNNHVGLTNxcGRsLWRtS0lISjQiLCJpYXQiOjE3MDE5ODc2NDYsImV4cCI6MzMyNTg4OTY0NDZ9.y51S2D9qMHLRixj230YZbvQZyhWzDOQ2RPbyJmnEYXA" //nolint:gosec

	tests := []struct {
		name         string
		ctx          echo.Context
		accessToken  string
		refreshToken string
	}{

		{
			name:         happy,
			ctx:          validCtx,
			accessToken:  testAccessToken,
			refreshToken: testRefreshToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth.SetAuthCookies(tc.ctx.Response().Writer, tc.accessToken, tc.refreshToken, *sessions.DebugCookieConfig)
		})
	}
}

func TestClearAuthCookies(t *testing.T) {
	validCtx, err := auth.NewTestEchoContextWithValidUser("funks")
	if err != nil {
		t.Fatal()
	}

	tests := []struct {
		name         string
		ctx          echo.Context
		accessToken  string
		refreshToken string
		wantErr      bool
		err          error
	}{

		{
			name: happy,
			ctx:  validCtx,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auth.ClearAuthCookies(tc.ctx.Response().Writer)
		})
	}
}

func TestCookieExpired(t *testing.T) {
	tests := []struct {
		name   string
		cookie *http.Cookie
		want   bool
	}{
		{
			name: "expired cookie based on expires",
			cookie: &http.Cookie{
				Name:    auth.AccessTokenCookie,
				Value:   "access_token_from_cookie",
				Expires: time.Now().Add(-1 * time.Minute),
			},
			want: true,
		},
		{
			name: "expired cookie based on max age",
			cookie: &http.Cookie{
				Name:   auth.AccessTokenCookie,
				Value:  "access_token_from_cookie",
				MaxAge: -1,
			},
			want: true,
		},
		{
			name: "not expired",
			cookie: &http.Cookie{
				Name:   auth.AccessTokenCookie,
				Value:  "access_token_from_cookie",
				MaxAge: 3600,
			},
			want: false,
		},
		{
			name: "not expired",
			cookie: &http.Cookie{
				Name:    auth.AccessTokenCookie,
				Value:   "access_token_from_cookie",
				Expires: time.Now().Add(10 * time.Minute),
			},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := auth.CookieExpired(tc.cookie)
			assert.Equal(t, tc.want, got)
		})
	}
}
