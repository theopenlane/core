//go:build cli

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/knadh/koanf/v2"
	"golang.org/x/oauth2"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/utils/keyring"

	models "github.com/theopenlane/core/pkg/openapi"
	openlane "github.com/theopenlane/go-client"
)

const (
	serviceName     = "openlane"
	accessTokenKey  = "open_lane_token"
	refreshTokenKey = "open_lane_refresh_token" //nolint:gosec
	sessionKey      = "open_lane_session"

	csrfHeader = "X-CSRF-Token"
	csrfCookie = "ol.csrf-token" // this should match the cookie name in the server config
	csrfPath   = "/livez"
)

var (
	ErrEmptyCSRFToken = errors.New("empty csrf token received from server, cannot continue")
	ErrNoCookieJarSet = errors.New("client does not have a cookie jar, cannot set cookies")
)

// TokenAuth uses the token or personal access token to authenticate the client
// if the token is not provided, it will fall back to JWT and session auth
func TokenAuth(ctx context.Context, k *koanf.Koanf) (*openlane.Client, error) {
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

	var opts = []openlane.ClientOption{
		openlane.WithAPIToken(token),
		openlane.WithBaseURL(RootHost),
	}

	// setup the logging interceptor
	if Config.Bool("debug") {
		opts = append(opts, openlane.WithInterceptors(openlane.WithLoggingInterceptor()))
	}

	return openlane.New(opts...)
}

// SetupClientWithAuth will setup the openlane client with the the bearer token passed in the Authorization header
// and the session cookie passed in the Cookie header. If the token is expired, it will be refreshed.
// The token and session will be stored in the keyring for future requests
func SetupClientWithAuth(ctx context.Context) (*openlane.Client, error) {
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

	opts := []openlane.ClientOption{
		openlane.WithBaseURL(RootHost),
	}

	opts = append(opts, openlane.WithCredentials(openlane.Authorization{
		BearerToken: token.AccessToken,
		Session:     session,
	}))

	client, err := openlane.New(opts...)
	if err != nil {
		return nil, err
	}

	if Config.Bool("disable-csrf") {
		// If CSRF is disabled, return the client without CSRF token
		return client, nil
	}

	return clientWithCSRFToken(ctx, client, opts...)
}

// SetupClient will setup the client without the Authorization header
// this is used for endpoints that do not require authentication, e.g. `v1/login`
func SetupClient(ctx context.Context) (*openlane.Client, error) {
	opts := []openlane.ClientOption{
		openlane.WithBaseURL(RootHost),
	}

	client, err := openlane.New(opts...)
	if err != nil {
		return nil, err
	}

	if Config.Bool("disable-csrf") {
		// If CSRF is disabled, return the client without CSRF token
		return client, nil
	}

	return clientWithCSRFToken(ctx, client, opts...)
}

func initCSRF(ctx context.Context, client *openlane.Client) (string, error) {
	token, err := fetchCSRFToken(ctx, client)
	if err != nil {
		return "", err
	}

	return token, nil
}

// fetchCSRFToken performs a safe request to retrieve the CSRF cookie value.
func fetchCSRFToken(ctx context.Context, client *openlane.Client) (string, error) {
	if client.HTTPSlingRequester().CookieJar() == nil {
		return "", ErrNoCookieJarSet
	}

	// make a GET request to acquire the CSRF cookie
	resp, err := client.HTTPSlingRequester().ReceiveWithContext(ctx, nil, httpsling.Get(csrfPath))
	if err != nil {
		return "", err
	}

	if resp != nil {
		resp.Body.Close()
	}

	return getCSRFToken(client)
}

// getCSRFToken retrieves the CSRF token from the cookie jar
// and returns it. If the token is not found or is empty, it returns an error.
// if it doesn't exist, it returns an empty string without an error.
// this is used for cases where CSRF protection is not enabled.
func getCSRFToken(client *openlane.Client) (string, error) {
	cookies, err := client.Cookies()
	if err != nil {
		return "", err
	}

	for _, ck := range cookies {
		if ck.Name == csrfCookie {
			if ck.Value == "" {
				return "", ErrEmptyCSRFToken
			}

			return ck.Value, nil
		}
	}

	// do not return an error, if CSRF protection is not enabled
	// there may not be a CSRF cookie set
	return "", nil
}

func clientWithCSRFToken(ctx context.Context, client *openlane.Client, opts ...openlane.ClientOption) (*openlane.Client, error) {
	// initialize csrf token for subsequent requests
	csrfToken, err := initCSRF(ctx, client)
	if err != nil {
		return nil, err
	}

	opts = append(opts, openlane.WithCSRFToken(csrfToken))

	return cloneClientWithCookies(client, opts...)
}

func cloneClientWithCookies(client *openlane.Client, opts ...openlane.ClientOption) (*openlane.Client, error) {
	newClient, err := openlane.New(opts...)
	if err != nil {
		return nil, err
	}

	// Copy cookies from the original client to the new one
	cookies, err := client.Cookies()
	if err != nil {
		return nil, err
	}

	u := newClient.Config().BaseURL.ResolveReference(&url.URL{Path: "/"})
	newClient.HTTPSlingRequester().CookieJar().SetCookies(u, cookies)

	return newClient, nil
}

// StoreSessionCookies gets the session cookie from the cookie jar
// and stores it in the keychain for future requests
func StoreSessionCookies(client *openlane.Client) {
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
func StoreAuthCookies(client *openlane.Client) {
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
