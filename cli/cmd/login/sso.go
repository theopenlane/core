//go:build cli

package login

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/browser"
	"github.com/theopenlane/httpsling"
	"golang.org/x/oauth2"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/sessions"

	corecmd "github.com/theopenlane/core/cli/cmd"
	models "github.com/theopenlane/core/common/openapi"
	sso "github.com/theopenlane/core/pkg/ssoutils"
	openlane "github.com/theopenlane/go-client"
)

// ssoConfig configures the local SSO login server
type ssoConfig struct {
	listenAddr string
	timeout    time.Duration
}

// SSOOption configures ssoConfig
type SSOOption func(*ssoConfig)

// WithListenAddr sets the listening address for the local callback server
func WithListenAddr(addr string) SSOOption {
	return func(c *ssoConfig) {
		c.listenAddr = addr
	}
}

// WithTimeout sets the timeout for the SSO flow
func WithTimeout(d time.Duration) SSOOption {
	return func(c *ssoConfig) {
		c.timeout = d
	}
}

// newSSOConfig returns a new ssoConfig with the provided options
func newSSOConfig(opts ...SSOOption) *ssoConfig {
	cfg := &ssoConfig{
		listenAddr: "127.0.0.1:0",
		timeout:    2 * time.Minute,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// fetchSSOStatus queries the webfinger endpoint for SSO enforcement status
func fetchSSOStatus(ctx context.Context, client *openlane.Client, email string) (models.SSOStatusReply, error) {
	var out models.SSOStatusReply
	resp, err := client.HTTPSlingRequester().ReceiveWithContext(ctx, &out,
		httpsling.Get("/.well-known/webfinger"),
		httpsling.QueryParam("resource", fmt.Sprintf("acct:%s", email)),
	)
	if err != nil {
		return out, err
	}

	if resp != nil {
		resp.Body.Close()
	}

	if !httpsling.IsSuccess(resp) {
		return out, fmt.Errorf("sso status error: %s", out.Error)
	}

	return out, nil
}

// ssoAuth handles SSO login by opening the browser and capturing the callback
func ssoAuth(ctx context.Context, client *openlane.Client, orgID string, opts ...SSOOption) (*oauth2.Token, error) {
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
			_ = corecmd.StoreSession(sessionID)
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
