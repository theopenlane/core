package oauth2

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/pkg/testutils"
)

const (
	FailureHandlerCalled = "failure handler called"
	CodePath             = "/?code=any_code&state=d4e5f6"
	SuccessHandlerCalled = "success handler called"
)

func TestLoginHandler(t *testing.T) {
	expectedState := "state_val"
	expectedRedirect := "https://api.example.com/authorize?client_id=client_id&redirect_uri=redirect_url&response_type=code&state=state_val"
	config := &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "redirect_url",
		Endpoint: oauth2.Endpoint{
			AuthURL: "https://api.example.com/authorize",
		},
	}

	failure := testutils.AssertFailureNotCalled(t)

	loginHandler := LoginHandler(config, failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil) // nolint: noctx

	ctx := WithState(context.Background(), expectedState)
	loginHandler.ServeHTTP(w, req.WithContext(ctx))

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, expectedRedirect, w.Result().Header.Get("Location")) // nolint: bodyclose
}

func TestLoginState(t *testing.T) {
	config := &oauth2.Config{}
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := ErrorFromContext(ctx)

		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: context missing state value", err.Error())
		}

		fmt.Fprint(w, FailureHandlerCalled)
	}

	loginHandler := LoginHandler(config, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil) // nolint: noctx
	loginHandler.ServeHTTP(w, req.WithContext(req.Context()))
	assert.Equal(t, FailureHandlerCalled, w.Body.String())
}

func TestCallbackHandler(t *testing.T) {
	jsonData := `{
       "access_token":"2YotnFZFEjr1zCsicMWpAA",
       "token_type":"example",
       "refresh_token":"tGzv3JOkF0XG5Qx2TlKWIA",
       "example_parameter":"example_value"
     }`
	expectedToken := &oauth2.Token{
		AccessToken:  "2YotnFZFEjr1zCsicMWpAA",
		TokenType:    "example",
		RefreshToken: "tGzv3JOkF0XG5Qx2TlKWIA",
	}

	server := NewAccessTokenServer(t, jsonData)
	defer server.Close()

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			TokenURL: server.URL,
		},
	}

	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := TokenFromContext(ctx)
		assert.Equal(t, expectedToken.AccessToken, token.AccessToken)
		assert.Equal(t, expectedToken.TokenType, token.Type())
		assert.Equal(t, expectedToken.RefreshToken, token.RefreshToken)
		// real oauth2.Token populates internal raw field and unmockable Expiry time
		assert.Nil(t, err)
		fmt.Fprint(w, SuccessHandlerCalled)
	}

	failure := testutils.AssertFailureNotCalled(t)
	callbackHandler := CallbackHandler(config, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", CodePath, nil) // nolint: noctx
	ctx := WithState(context.Background(), "d4e5f6")
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, SuccessHandlerCalled, w.Body.String())
}

func TestParseCallback(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := ErrorFromContext(ctx)

		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: request missing code or state", err.Error())
		}

		fmt.Fprint(w, FailureHandlerCalled)
	}

	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/?code=any_code", nil)
	callbackHandler.ServeHTTP(w, req.WithContext(req.Context()))
	assert.Equal(t, FailureHandlerCalled, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(req.Context(), "GET", "/?state=any_state", nil)
	callbackHandler.ServeHTTP(w, req.WithContext(req.Context()))
	assert.Equal(t, FailureHandlerCalled, w.Body.String())
}

func TestStateHandler(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := ErrorFromContext(ctx)

		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: context missing state value", err.Error())
		}

		fmt.Fprint(w, FailureHandlerCalled)
	}

	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", CodePath, nil) // nolint: noctx
	callbackHandler.ServeHTTP(w, req.WithContext(req.Context()))
	assert.Equal(t, FailureHandlerCalled, w.Body.String())
}

func TestStateFromContext(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := ErrorFromContext(ctx)

		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: invalid oauth2 state parameter", err.Error())
		}

		fmt.Fprint(w, FailureHandlerCalled)
	}

	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", CodePath, nil) // nolint: noctx
	ctx := WithState(context.Background(), "differentState")
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, FailureHandlerCalled, w.Body.String())
}

func TestTokenExchange(t *testing.T) {
	_, server := testutils.NewErrorServer("oAuth is no auth'in", http.StatusInternalServerError)
	defer server.Close()

	config := &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			TokenURL: server.URL,
		},
	}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := ErrorFromContext(ctx)

		if assert.NotNil(t, err) {
			// error from golang.org/x/oauth2 config.Exchange as provider is down
			assert.True(t, strings.HasPrefix(err.Error(), "oauth2: cannot fetch token"))
		}

		fmt.Fprint(w, FailureHandlerCalled)
	}

	callbackHandler := CallbackHandler(config, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", CodePath, nil) // nolint: noctx
	ctx := WithState(context.Background(), "d4e5f6")
	callbackHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, FailureHandlerCalled, w.Body.String())
}
