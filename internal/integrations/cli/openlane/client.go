//go:build examples

package openlane

import (
	"context"
	"fmt"
	"strings"
	"syscall"

	openlaneclient "github.com/theopenlane/go-client"
	"github.com/theopenlane/iam/auth"
	"golang.org/x/term"

	api "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/integrations/cli/config"
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

	switch {
	case resp.TFASetupRequired:
		return nil, fmt.Errorf("%w: enroll in the console or use auth-mode=token", auth.ErrTFAEnrollmentRequired)
	case resp.TFAEnabled:
		resp, err = completeSecondFactor(ctx, client)
		if err != nil {
			return nil, err
		}
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

// completeSecondFactor reads the TOTP code from the terminal and completes the pending login
func completeSecondFactor(ctx context.Context, client *openlaneclient.Client) (*api.LoginResponse, error) {
	if !term.IsTerminal(int(syscall.Stdin)) {
		return nil, fmt.Errorf("%w: no terminal to read the code, use auth-mode=token", auth.ErrSecondFactorRequired)
	}

	fmt.Print("2FA code: ")

	code, err := term.ReadPassword(int(syscall.Stdin))

	fmt.Println()

	if err != nil {
		return nil, err
	}

	resp, err := client.ValidateLoginTOTP(ctx, &api.TFARequest{TOTPCode: string(code)})
	if err != nil || resp.AccessToken == "" {
		return nil, ErrLogin
	}

	return resp, nil
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
