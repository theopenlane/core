//go:build cli

package login

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/theopenlane/httpsling"
	"golang.org/x/oauth2"
	"golang.org/x/term"

	cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	modelapi "github.com/theopenlane/core/pkg/openapi"
	"github.com/theopenlane/core/pkg/openlaneclient"
	sso "github.com/theopenlane/core/pkg/ssoutils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"
)

func loginPrimaryHook(_ *speccli.PrimarySpec) speccli.PrimaryPreHook {
	return func(ctx context.Context, _ *cobra.Command) error {
		_, err := runLogin(ctx)
		return err
	}
}

func runLogin(ctx context.Context) (*oauth2.Token, error) {
	client, err := cmdpkg.SetupClient(ctx)
	if err != nil {
		return nil, err
	}

	username := cmdpkg.Config.String("username")
	if strings.TrimSpace(username) == "" {
		return nil, cmdpkg.NewRequiredFieldMissingError("username")
	}

	status, err := fetchSSOStatus(ctx, client, username)
	if err == nil && status.Enforced && status.OrganizationID != "" {
		if status.OrgTFAEnforced && !status.UserTFAEnabled {
			fmt.Println("\nTwo-factor authentication is required by your organization.")
			fmt.Println("Please login to the Openlane console and enable 2FA on your profile page before using the CLI.")
			return nil, fmt.Errorf("2FA required but not enabled on user account")
		}

		tokens, err := ssoAuth(ctx, client, status.OrganizationID)
		if err != nil {
			return nil, err
		}

		if status.UserTFAEnabled {
			if err := handle2FAVerification(ctx, client); err != nil {
				return nil, err
			}
		}

		fmt.Println("\nAuthentication Successful!")
		if err := cmdpkg.StoreToken(tokens); err == nil {
			fmt.Println("auth tokens successfully stored in keychain")
		}

		return tokens, nil
	}

	tokens, err := passwordAuth(ctx, client, username)
	if err != nil {
		return nil, err
	}

	fmt.Println("\nAuthentication Successful!")

	if err := cmdpkg.StoreToken(tokens); err != nil {
		return nil, err
	}

	cmdpkg.StoreSessionCookies(client)
	fmt.Println("auth tokens successfully stored in keychain")

	return tokens, nil
}

func passwordAuth(ctx context.Context, client *openlaneclient.OpenlaneClient, username string) (*oauth2.Token, error) {
	password := cmdpkg.Config.String("password")

	if password == "" {
		fmt.Print("Password: ")

		bytepw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, err
		}

		password = string(bytepw)
	}

	login := modelapi.LoginRequest{
		Username: username,
		Password: password,
	}

	resp, err := client.Login(ctx, &login)
	if err != nil {
		return nil, err
	}

	if resp.TFASetupRequired && !resp.TFAEnabled {
		fmt.Println("\nTwo-factor authentication is required by your organization.")
		fmt.Println("Please login to the Openlane console and enable 2FA on your profile page before using the CLI.")
		return nil, fmt.Errorf("2FA required but not enabled on user account")
	}

	tokens := &oauth2.Token{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		TokenType:    resp.TokenType,
	}

	if resp.TFAEnabled {
		if err := handle2FAVerification(ctx, client); err != nil {
			return nil, err
		}
	}

	return tokens, nil
}

func handle2FAVerification(ctx context.Context, client *openlaneclient.OpenlaneClient) error {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		fmt.Printf("\nEnter your 2FA code (attempt %d of %d): ", attempt, maxAttempts)

		code, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read 2FA code: %w", err)
		}

		codeStr := strings.TrimSpace(string(code))
		fmt.Println()

		if codeStr == "" {
			fmt.Println("2FA code cannot be empty")
			continue
		}

		req := &modelapi.TFARequest{TOTPCode: codeStr}
		resp, err := client.ValidateTOTP(ctx, req)
		if err != nil {
			if attempt == maxAttempts {
				return fmt.Errorf("2FA verification failed after %d attempts: %w", maxAttempts, err)
			}

			fmt.Printf("Invalid 2FA code: %v\n", err)
			continue
		}

		if !resp.Success {
			if attempt == maxAttempts {
				return fmt.Errorf("2FA verification failed after %d attempts", maxAttempts)
			}

			fmt.Println("Invalid 2FA code")
			continue
		}

		fmt.Println("2FA verification successful")
		return nil
	}

	return fmt.Errorf("2FA verification failed after %d attempts", maxAttempts)
}

func fetchSSOStatus(ctx context.Context, client *openlaneclient.OpenlaneClient, email string) (modelapi.SSOStatusReply, error) {
	var out modelapi.SSOStatusReply
	resp, err := client.HTTPSlingRequester().ReceiveWithContext(ctx, &out,
		httpsling.Get("/.well-known/webfinger"),
		httpsling.QueryParam("resource", fmt.Sprintf("acct:%s", email)),
	)
	if err != nil {
		return out, err
	}

	if resp != nil {
		_ = resp.Body.Close()
	}

	if !httpsling.IsSuccess(resp) {
		return out, fmt.Errorf("sso status error: %s", out.Error)
	}

	return out, nil
}

type ssoConfig struct {
	listenAddr string
	timeout    time.Duration
}

type ssoOption func(*ssoConfig)

func withListenAddr(addr string) ssoOption {
	return func(c *ssoConfig) {
		c.listenAddr = addr
	}
}

func withTimeout(d time.Duration) ssoOption {
	return func(c *ssoConfig) {
		c.timeout = d
	}
}

func newSSOConfig(opts ...ssoOption) *ssoConfig {
	cfg := &ssoConfig{
		listenAddr: "127.0.0.1:0",
		timeout:    2 * time.Minute,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func ssoAuth(ctx context.Context, client *openlaneclient.OpenlaneClient, orgID string, opts ...ssoOption) (*oauth2.Token, error) {
	cfg := newSSOConfig(opts...)

	l, err := net.Listen("tcp", cfg.listenAddr)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{}
	tokenCh := make(chan *oauth2.Token, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var access, refresh, sessionID string
		for _, c := range r.Cookies() {
			switch c.Name {
			case auth.AccessTokenCookie:
				access = c.Value
			case auth.RefreshTokenCookie:
				refresh = c.Value
			case sessions.DevCookieName, sessions.DefaultCookieName:
				sessionID = c.Value
			}
		}

		fmt.Fprint(w, "Authentication Successful! You may close this window.")
		tokenCh <- &oauth2.Token{AccessToken: access, RefreshToken: refresh, TokenType: "bearer"}

		if sessionID != "" {
			_ = cmdpkg.StoreSession(sessionID)
		}

		go srv.Shutdown(ctx)
	})

	srv.Handler = mux

	go func() {
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			fmt.Printf("sso callback server error: %v\n", err)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil && err != http.ErrServerClosed {
			fmt.Printf("sso callback server shutdown error: %v\n", err)
		}
	}()

	cbURL := fmt.Sprintf("http://%s", l.Addr().String())

	u := *client.Config().BaseURL
	loginPath := sso.SSOLogin(nil, orgID)
	loginURL, err := url.Parse(loginPath)
	if err != nil {
		return nil, err
	}

	u.Path = loginURL.Path
	q := loginURL.Query()
	q.Set("return", cbURL)
	u.RawQuery = q.Encode()

	if err := browser.OpenURL(u.String()); err != nil {
		fmt.Printf("open the following URL to authenticate:\n%s\n", u.String())
	}

	select {
	case tok := <-tokenCh:
		return tok, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(cfg.timeout):
		return nil, fmt.Errorf("timeout waiting for SSO authentication")
	}
}
