package github

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"

	oauth2Login "github.com/theopenlane/core/pkg/providers/oauth2"
	"github.com/theopenlane/core/pkg/sessions"
)

const (
	ProviderName = "GITHUB"
)

// StateHandler checks for a state cookie, if found, adds to context; if missing, a
// random generated value is added to the context and to a (short-lived) state cookie
// issued to the requester - this implements OAuth 2 RFC 6749 10.12 CSRF Protection
func StateHandler(config sessions.CookieConfig, success http.Handler) http.Handler {
	return oauth2Login.StateHandler(config, success)
}

// LoginHandler handles Github login requests by reading the state value from
// the ctx and redirecting requests to the AuthURL with that state value
func LoginHandler(config *oauth2.Config, failure http.Handler) http.Handler {
	return oauth2Login.LoginHandler(config, failure)
}

// CallbackHandler adds the GitHub access token and User to the ctx
func CallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = githubHandler(config, &ClientConfig{IsEnterprise: false, IsMock: false}, success, failure)

	return oauth2Login.CallbackHandler(config, success, failure)
}

// EnterpriseCallbackHandler handles GitHub Enterprise redirection URI requests
// and adds the GitHub access token and User to the ctx
func EnterpriseCallbackHandler(config *oauth2.Config, success, failure http.Handler) http.Handler {
	success = githubHandler(config, &ClientConfig{IsEnterprise: true, IsMock: false}, success, failure)

	return oauth2Login.CallbackHandler(config, success, failure)
}

// githubHandler gets the OAuth2 Token from the ctx to fetch the corresponding GitHub
// User and add them to the context
func githubHandler(config *oauth2.Config, clientConfig *ClientConfig, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = DefaultFailureHandler
	}

	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := oauth2Login.TokenFromContext(ctx)

		if err != nil {
			ctx = WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		githubClient, err := getGithubClient(ctx, clientConfig, config, token)
		if err != nil {
			ctx = WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		user, resp, err := githubClient.Users.Get(ctx, "")
		err = validateResponse(user, resp, err)
		if err != nil {
			ctx = WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))

			return
		}

		// Make a request to `user/emails` if the email was not returned (due to privacy)
		if user.Email == nil {
			user.Email, err = getUserEmails(ctx, githubClient)
			if err != nil {
				ctx = WithError(ctx, err)
				failure.ServeHTTP(w, req.WithContext(ctx))

				return
			}
		}

		ctx = WithUser(ctx, user)
		success.ServeHTTP(w, req.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

// validateResponse returns an error if we get something unexpected
func validateResponse(user *github.User, resp *github.Response, err error) error {
	if err != nil || resp.StatusCode != http.StatusOK {
		return ErrUnableToGetGithubUser
	}

	if user == nil || user.ID == nil {
		return ErrUnableToGetGithubUser
	}

	return nil
}

func getGithubClient(ctx context.Context, clientConfig *ClientConfig, config *oauth2.Config, token *oauth2.Token) (githubClient GitHubClient, err error) {
	// create httClient with provided token
	httpClient := config.Client(ctx, token)

	// setup client
	var g GitHubInterface

	g = &GitHubCreator{}
	if clientConfig.IsMock {
		g = &GitHubMock{}
	}

	// Set params for enterprise Github client
	if clientConfig.IsEnterprise {
		clientConfig, err = enterpriseGithubClientFromAuthURL(config.Endpoint.AuthURL, clientConfig)
		if err != nil {
			return
		}
	}

	// Set config on client
	g.SetConfig(clientConfig)

	// Create new Github Client
	githubClient = g.NewClient(httpClient)

	return
}

// enterpriseGithubClientFromAuthURL returns a client config with required settings for GHE instance
func enterpriseGithubClientFromAuthURL(authURL string, config *ClientConfig) (*ClientConfig, error) {
	baseURL, err := url.Parse(authURL)
	if err != nil {
		return config, ErrFailedConstructingEndpointURL
	}

	baseURL.Path = "/api/v3/"
	config.BaseURL = baseURL
	config.UploadURL = baseURL

	return config, nil
}

// getUserEmails from `user/emails` and return the user's primary email address
func getUserEmails(ctx context.Context, githubClient GitHubClient) (*string, error) {
	emails, _, err := githubClient.Users.ListEmails(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Get the primary email
	for _, em := range emails {
		if em.GetPrimary() {
			return em.Email, nil
		}
	}

	return nil, ErrPrimaryEmailNotFound
}

// VerifyClientToken checks the client token and returns an error if it is invalid
func VerifyClientToken(ctx context.Context, token *oauth2.Token, config *oauth2.Config, email string, clientConfig *ClientConfig) (err error) {
	githubClient, err := getGithubClient(ctx, clientConfig, config, token)
	if err != nil {
		return err
	}

	user, resp, err := githubClient.Users.Get(ctx, "")

	// ensure the emails match
	githubEmail, err := getUserEmails(ctx, githubClient)
	if err != nil {
		return err
	}

	if *githubEmail != email {
		return ErrUnableToGetGithubUser
	}

	return validateResponse(user, resp, err)
}
