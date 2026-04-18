// Package openlane wraps the Openlane API client for use inside the
// integrations CLI. Its sole responsibility is turning a resolved
// OpenlaneConfig into an authenticated *openlaneclient.Client; anything past
// that belongs in integration-specific subcommands.
package openlane

import (
	"context"
	"strings"

	openlaneclient "github.com/theopenlane/go-client"

	api "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/integrations/cli/config"
)

// Auth modes supported by Connect
const (
	AuthModeAuto        = "auto"
	AuthModeToken       = "token"
	AuthModeCredentials = "credentials"
)

// Connect returns an authenticated Openlane client for the given config.
// auto mode prefers token/PAT when present and falls back to credential login;
// token mode requires a token; credentials mode requires email/password.
func Connect(ctx context.Context, cfg config.OpenlaneConfig) (*openlaneclient.Client, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return nil, ErrHostRequired
	}

	token := firstNonEmpty(cfg.Auth.Token, cfg.Auth.PAT)
	mode := strings.ToLower(strings.TrimSpace(cfg.Auth.Mode))

	if mode == "" || mode == AuthModeAuto {
		if token != "" {
			mode = AuthModeToken
		} else {
			mode = AuthModeCredentials
		}
	}

	switch mode {
	case AuthModeToken:
		if token == "" {
			return nil, ErrTokenRequired
		}

		return openlaneclient.New(
			openlaneclient.WithBaseURL(host),
			openlaneclient.WithCredentials(openlaneclient.Authorization{BearerToken: token}),
		)
	case AuthModeCredentials:
		if cfg.Auth.Email == "" || cfg.Auth.Password == "" {
			return nil, ErrCredentialsRequired
		}

		return loginWithCredentials(ctx, host, cfg.Auth.Email, cfg.Auth.Password)
	default:
		return nil, ErrUnsupportedAuthMode
	}
}

// loginWithCredentials exchanges email/password for an authenticated client
func loginWithCredentials(ctx context.Context, host, email, password string) (*openlaneclient.Client, error) {
	client, err := openlaneclient.New(openlaneclient.WithBaseURL(host))
	if err != nil {
		return nil, err
	}

	resp, err := client.Login(ctx, &api.LoginRequest{Username: email, Password: password})
	if err != nil {
		return nil, ErrLogin
	}

	session := strings.TrimSpace(resp.Session)
	if session == "" {
		if jarSession, jarErr := client.GetSessionFromCookieJar(); jarErr == nil {
			session = jarSession
		}
	}

	return openlaneclient.New(
		openlaneclient.WithBaseURL(host),
		openlaneclient.WithCredentials(openlaneclient.Authorization{
			BearerToken: resp.AccessToken,
			Session:     session,
		}),
	)
}

// firstNonEmpty returns the first non-empty string from values
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}

	return ""
}
