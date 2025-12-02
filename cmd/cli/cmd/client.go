//go:build cli

package cmd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/knadh/koanf/v2"
	"golang.org/x/oauth2"

	goclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/keyring"

	models "github.com/theopenlane/shared/openapi"
)

const (
	serviceName     = "openlane"
	accessTokenKey  = "open_lane_token"
	refreshTokenKey = "open_lane_refresh_token" // nolint:gosec
	sessionKey      = "open_lane_session"
)

// TokenAuth uses the token or personal access token to authenticate the client
// if the token is not provided, it will fall back to JWT and session auth
func TokenAuth(ctx context.Context, k *koanf.Koanf) (*goclient.OpenlaneClient, error) {
	var token string
	for _, key := range []string{"token", "pat", "jwt"} {
		token = k.String(key)
		if token != "" {
			break
		}
	}

	if token == "" {
		return nil, fmt.Errorf("no token provided, will fall back to JWT and session auth")
	}

	config, opts, err := configureDefaultOpts()
	if err != nil {
		return nil, err
	}

	opts = append(opts, goclient.WithCredentials(goclient.Authorization{
		BearerToken: token,
	}))

	return goclient.New(config, opts...)
}

// SetupClientWithAuth will setup the openlane client with the the bearer token passed in the Authorization header
// and the session cookie passed in the Cookie header. If the token is expired, it will be refreshed.
// The token and session will be stored in the keyring for future requests
func SetupClientWithAuth(ctx context.Context) (*goclient.OpenlaneClient, error) {
	// setup interceptors
	token, session, err := GetTokenFromKeyring(ctx)
	if err != nil {
		return nil, err
	}

	expired, err := tokens.IsExpired(token.AccessToken)
	if err != nil {
		return nil, err
	}

	// refresh and store the new token pair if the existing access token
	// is expired
	if expired {
		// refresh the token pair using the refresh token
		token, err = refreshToken(ctx, token.RefreshToken)
		if err != nil {
			return nil, err
		}

		// store the new token
		if err := StoreToken(token); err != nil {
			return nil, err
		}
	}

	config, opts, err := configureDefaultOpts()
	if err != nil {
		return nil, err
	}

	opts = append(opts, goclient.WithCredentials(goclient.Authorization{
		BearerToken: token.AccessToken,
		Session:     session,
	}))

	client, err := goclient.New(config, opts...)
	if err != nil {
		return nil, err
	}

	if Config.Bool("disable-csrf") {
		// If CSRF is disabled, return the client without CSRF token
		return client, nil
	}

	return client.ClientWithCSRFToken(ctx, opts...)
}

// SetupClient will setup the client without the Authorization header
// this is used for endpoints that do not require authentication, e.g. `v1/login`
func SetupClient(ctx context.Context) (*goclient.OpenlaneClient, error) {
	config, opts, err := configureDefaultOpts()
	if err != nil {
		return nil, err
	}

	client, err := goclient.New(config, opts...)
	if err != nil {
		return nil, err
	}

	if Config.Bool("disable-csrf") {
		// If CSRF is disabled, return the client without CSRF token
		return client, nil
	}

	return client.ClientWithCSRFToken(ctx, opts...)
}

// configureDefaultOpts will setup the default options for the client
func configureDefaultOpts() (goclient.Config, []goclient.ClientOption, error) {
	config := goclient.NewDefaultConfig()

	// setup the logging interceptor
	if Config.Bool("debug") {
		config.Interceptors = append(config.Interceptors, goclient.WithLoggingInterceptor())
	}

	endpointOpt, err := configureClientEndpoints()
	if err != nil {
		return config, nil, err
	}

	return config, []goclient.ClientOption{endpointOpt}, nil
}

// configureClientEndpoints will setup the base URL for the client
func configureClientEndpoints() (goclient.ClientOption, error) {
	baseURL, err := url.Parse(RootHost)
	if err != nil {
		return nil, err
	}

	return goclient.WithBaseURL(baseURL), nil
}

// StoreSessionCookies gets the session cookie from the cookie jar
// and stores it in the keychain for future requests
func StoreSessionCookies(client *goclient.OpenlaneClient) {
	session, err := client.GetSessionFromCookieJar()
	if err != nil || session == "" {
		fmt.Println("unable to get session from cookie jar")

		return
	}

	if err := StoreSession(session); err != nil {
		fmt.Println("unable to store session in keychain")

		return
	}

	// store the auth cookies if they exist
	StoreAuthCookies(client)
}

// StoreAuthCookies gets the auth cookies from the cookie jar if they exist
// and stores them in the keychain for future requests
func StoreAuthCookies(client *goclient.OpenlaneClient) {
	token := client.GetAuthTokensFromCookieJar()

	if token == nil {
		return // no auth cookies found, nothing to store
	}

	if err := StoreToken(token); err != nil {
		fmt.Println("unable to store auth tokens in keychain")

		return
	}
}

// GetTokenFromKeyring will return the oauth token from the keyring
// if the token is expired, but the refresh token is still valid, the
// token will be refreshed
func GetTokenFromKeyring(ctx context.Context) (*oauth2.Token, string, error) {
	access, err := keyring.QueryKeyring(serviceName, accessTokenKey)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching auth token: %w", err)
	}

	refresh, err := keyring.QueryKeyring(serviceName, refreshTokenKey)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching refresh token: %w", err)
	}

	session, err := keyring.QueryKeyring(serviceName, sessionKey)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching session: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  access,
		RefreshToken: refresh,
	}, session, nil
}

// refreshToken will refresh the oauth token using the refresh token
func refreshToken(ctx context.Context, refresh string) (*oauth2.Token, error) {
	// setup http client
	client, err := SetupClient(ctx)
	if err != nil {
		return nil, err
	}

	req := models.RefreshRequest{
		RefreshToken: refresh,
	}

	resp, err := client.Refresh(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

// StoreToken in local keyring
func StoreToken(token *oauth2.Token) error {
	err := keyring.SetKeying(serviceName, accessTokenKey, []byte(token.AccessToken))
	if err != nil {
		return fmt.Errorf("failed saving access token: %w", err)
	}

	err = keyring.SetKeying(serviceName, refreshTokenKey, []byte(token.RefreshToken))
	if err != nil {
		return fmt.Errorf("failed saving refresh token: %w", err)
	}

	return nil
}

// StoreSession in local keyring
func StoreSession(session string) error {
	err := keyring.SetKeying(serviceName, sessionKey, []byte(session))
	if err != nil {
		return fmt.Errorf("failed saving session: %w", err)
	}

	return nil
}
