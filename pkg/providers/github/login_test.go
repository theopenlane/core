package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	oauth2Login "github.com/theopenlane/core/pkg/providers/oauth2"
	"github.com/theopenlane/core/pkg/testutils"

	"github.com/google/go-github/v63/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

const (
	SuccessHandlerCalled = "success handler called"
	FailureHandlerCalled = "failure handler called"
	anytoken             = "any-token"
)

func TestGithubHandler(t *testing.T) {
	jsonData := `{"id": 917408, "name": "Sarah Funkytown"}`
	emailJSONData := `[{"primary": true, "email": "sfunk@meow.net"}, {"primary": false, "email": "sfunk@woof.net"}]`

	expectedUser := &github.User{
		ID:    github.Int64(917408),
		Name:  github.String("Sarah Funkytown"),
		Email: github.String("sfunk@meow.net"),
	}

	proxyClient, server := newGithubTestServer("", jsonData, emailJSONData)

	defer server.Close()

	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: anytoken}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		githubUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, githubUser)
		fmt.Fprint(w, SuccessHandlerCalled)
	}
	failure := testutils.AssertFailureNotCalled(t)

	// GithubHandler assert that:
	// - Token is read from the ctx and passed to the GitHub API
	// - github User is obtained from the GitHub API
	// - success handler is called
	// - github User is added to the ctx of the success handler
	githubHandler := githubHandler(config, &ClientConfig{IsEnterprise: false, IsMock: false}, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil) // nolint: noctx
	githubHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, SuccessHandlerCalled, w.Body.String())
}

func TestMissingCtxToken(t *testing.T) {
	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := ErrorFromContext(ctx)

		if assert.NotNil(t, err) {
			assert.Equal(t, "oauth2: context missing token", err.Error())
		}

		fmt.Fprint(w, FailureHandlerCalled)
	}

	// GithubHandler called without Token in ctx, assert that:
	// - failure handler is called
	// - error about ctx missing token is added to the failure handler ctx
	githubHandler := githubHandler(config, &ClientConfig{IsEnterprise: false, IsMock: false}, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil) // nolint: noctx
	githubHandler.ServeHTTP(w, req)
	assert.Equal(t, FailureHandlerCalled, w.Body.String())
}

func TestErrorGettingUser(t *testing.T) {
	proxyClient, server := testutils.NewErrorServer("GitHub Service Down", http.StatusInternalServerError)
	defer server.Close()
	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: anytoken}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	success := testutils.AssertSuccessNotCalled(t)
	failure := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := ErrorFromContext(ctx)

		if assert.NotNil(t, err) {
			assert.Equal(t, ErrUnableToGetGithubUser, err)
		}

		fmt.Fprint(w, FailureHandlerCalled)
	}

	// GithubHandler cannot get GitHub User, assert that:
	// - failure handler is called
	// - error cannot get GitHub User added to the failure handler ctx
	githubHandler := githubHandler(config, &ClientConfig{IsEnterprise: false, IsMock: false}, success, http.HandlerFunc(failure))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil) // nolint: noctx
	githubHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, FailureHandlerCalled, w.Body.String())
}

func TestGithubEnterprise(t *testing.T) {
	jsonData := `{"id": 917408, "name": "Sarah Funkytown"}`
	emailJSONData := `[{"primary": true, "email": "sfunk@meow.net"}, {"primary": false, "email": "sfunk@woof.net"}]`
	expectedUser := &github.User{
		ID:    github.Int64(917408),
		Name:  github.String("Sarah Funkytown"),
		Email: github.String("sfunk@meow.net"),
	}

	proxyClient, server := newGithubTestServer("/api/v3", jsonData, emailJSONData)

	defer server.Close()

	// oauth2 Client will use the proxy client's base Transport
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, proxyClient)
	anyToken := &oauth2.Token{AccessToken: anytoken}
	ctx = oauth2Login.WithToken(ctx, anyToken)

	config := &oauth2.Config{}
	config.Endpoint.AuthURL = "https://github.mattisthebest.com/login/oauth/authorize"
	success := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		githubUser, err := UserFromContext(ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedUser, githubUser)
		fmt.Fprint(w, SuccessHandlerCalled)
	}
	failure := testutils.AssertFailureNotCalled(t)

	// GithubHandler assert that:
	// - Token is read from the ctx and passed to the GitHub API
	// - github User is obtained from the GitHub API
	// - success handler is called
	// - github User is added to the ctx of the success handler
	githubHandler := githubHandler(config, &ClientConfig{IsEnterprise: true, IsMock: false}, http.HandlerFunc(success), failure)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil) // nolint: noctx
	githubHandler.ServeHTTP(w, req.WithContext(ctx))
	assert.Equal(t, SuccessHandlerCalled, w.Body.String())
}

func TestValidateResponse(t *testing.T) {
	validUser := &github.User{ID: github.Int64(123)}
	validResponse := &github.Response{Response: &http.Response{StatusCode: 200}}
	invalidResponse := &github.Response{Response: &http.Response{StatusCode: 500}}

	assert.Equal(t, nil, validateResponse(validUser, validResponse, nil))
	assert.Equal(t, ErrUnableToGetGithubUser, validateResponse(validUser, validResponse, ErrServerError))
	assert.Equal(t, ErrUnableToGetGithubUser, validateResponse(validUser, invalidResponse, nil))
	assert.Equal(t, ErrUnableToGetGithubUser, validateResponse(&github.User{}, validResponse, nil))
}

func TestEnterpriseGithubClientFromAuthURL(t *testing.T) {
	cases := []struct {
		authURL          string
		expClientBaseURL string
	}{
		{"https://github.mattisthebest.com/login/oauth/authorize", "https://github.mattisthebest.com/api/v3/"},
		{"http://github.mattisthebest.com/login/oauth/authorize", "http://github.mattisthebest.com/api/v3/"},
	}
	for _, c := range cases {
		client, err := enterpriseGithubClientFromAuthURL(c.authURL, &ClientConfig{IsEnterprise: true, IsMock: false})
		assert.Nil(t, err)
		assert.Equal(t, client.BaseURL.String(), c.expClientBaseURL)
	}
}
