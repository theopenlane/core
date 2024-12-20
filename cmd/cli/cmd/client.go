package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/99designs/keyring"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

const (
	serviceName     = "openlane"
	accessTokenKey  = "open_lane_token"
	refreshTokenKey = "open_lane_refresh_token" // nolint:gosec
	sessionKey      = "open_lane_session"
)

func TokenAuth(ctx context.Context, k *koanf.Koanf) (*openlaneclient.OpenlaneClient, error) {
	token := k.String("token")
	if token == "" {
		token = k.String("pat")
	}

	if token == "" {
		return nil, fmt.Errorf("no token provided, will fall back to JWT and session auth")
	}

	log.Debug().Msg("setting up client with token")
	config, opts, err := configureDefaultOpts()
	if err != nil {
		return nil, err
	}

	opts = append(opts, openlaneclient.WithCredentials(openlaneclient.Authorization{
		BearerToken: token,
	}))

	return openlaneclient.New(config, opts...)
}

// SetupClientWithAuth will setup the openlane client with the the bearer token passed in the Authorization header
// and the session cookie passed in the Cookie header. If the token is expired, it will be refreshed.
// The token and session will be stored in the keyring for future requests
func SetupClientWithAuth(ctx context.Context) (*openlaneclient.OpenlaneClient, error) {
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

	opts = append(opts, openlaneclient.WithCredentials(openlaneclient.Authorization{
		BearerToken: token.AccessToken,
		Session:     session,
	}))

	return openlaneclient.New(config, opts...)
}

// SetupClient will setup the client without the Authorization header
// this is used for endpoints that do not require authentication, e.g. `v1/login`
func SetupClient(ctx context.Context) (*openlaneclient.OpenlaneClient, error) {
	config, opts, err := configureDefaultOpts()
	if err != nil {
		return nil, err
	}

	return openlaneclient.New(config, opts...)
}

// configureDefaultOpts will setup the default options for the client
func configureDefaultOpts() (openlaneclient.Config, []openlaneclient.ClientOption, error) {
	config := openlaneclient.NewDefaultConfig()

	// setup the logging interceptor
	if Config.Bool("debug") {
		config.Interceptors = append(config.Interceptors, openlaneclient.WithLoggingInterceptor())
	}

	endpointOpt, err := configureClientEndpoints()
	if err != nil {
		return config, nil, err
	}

	return config, []openlaneclient.ClientOption{endpointOpt}, nil
}

// configureClientEndpoints will setup the base URL for the client
func configureClientEndpoints() (openlaneclient.ClientOption, error) {
	baseURL, err := url.Parse(RootHost)
	if err != nil {
		return nil, err
	}

	return openlaneclient.WithBaseURL(baseURL), nil
}

// StoreSessionCookies gets the session cookie from the cookie jar
// and stores it in the keychain for future requests
func StoreSessionCookies(client *openlaneclient.OpenlaneClient) {
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
func StoreAuthCookies(client *openlaneclient.OpenlaneClient) {
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
	ring, err := GetKeyring()
	if err != nil {
		return nil, "", fmt.Errorf("error opening keyring: %w", err)
	}

	access, err := ring.Get(accessTokenKey)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching auth token: %w", err)
	}

	refresh, err := ring.Get(refreshTokenKey)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching refresh token: %w", err)
	}

	session, err := ring.Get(sessionKey)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching refresh token: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  string(access.Data),
		RefreshToken: string(refresh.Data),
	}, string(session.Data), nil
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

// GetKeyring will return the already loaded keyring so that we don't prompt users for passwords multiple times
func GetKeyring() (keyring.Keyring, error) {
	var err error

	if userKeyringLoaded {
		return userKeyring, nil
	}

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	userKeyring, err = keyring.Open(keyring.Config{
		ServiceName: serviceName,

		// MacOS keychain
		KeychainTrustApplication: true,

		// KDE Wallet
		KWalletAppID:  serviceName,
		KWalletFolder: serviceName,

		// Windows
		WinCredPrefix: serviceName,

		// Fallback encrypted file
		FileDir:          path.Join(cfgDir, serviceName, "keyring"),
		FilePasswordFunc: keyring.TerminalPrompt,
	})
	if err == nil {
		userKeyringLoaded = true
	}

	return userKeyring, err
}

// StoreToken in local keyring
func StoreToken(token *oauth2.Token) error {
	ring, err := GetKeyring()
	if err != nil {
		return fmt.Errorf("error opening keyring: %w", err)
	}

	err = ring.Set(keyring.Item{
		Key:  accessTokenKey,
		Data: []byte(token.AccessToken),
	})
	if err != nil {
		return fmt.Errorf("failed saving access token: %w", err)
	}

	err = ring.Set(keyring.Item{
		Key:  refreshTokenKey,
		Data: []byte(token.RefreshToken),
	})
	if err != nil {
		return fmt.Errorf("failed saving refresh token: %w", err)
	}

	return nil
}

// StoreSession in local keyring
func StoreSession(session string) error {
	ring, err := GetKeyring()
	if err != nil {
		return fmt.Errorf("error opening keyring: %w", err)
	}

	err = ring.Set(keyring.Item{
		Key:  sessionKey,
		Data: []byte(session),
	})
	if err != nil {
		return fmt.Errorf("failed saving session: %w", err)
	}

	return nil
}
