package zitadel

import (
	"context"
	"net"
	"strconv"
	"strings"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"

	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// zitadelDefaultPageSize is the number of records requested per Zitadel API page
	zitadelDefaultPageSize = 100
)

// Client builds Zitadel user service clients for one installation
type Client struct{}

// Build constructs the Zitadel user service client for one installation. It supports both
// Personal Access Token and OAuth2 client-credentials auth modes, selecting whichever
// credential the installation was configured with.
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	domain, auth, err := resolveAuth(req.Credentials)
	if err != nil {
		return nil, err
	}

	host, opts := parseHost(domain)

	api, err := client.New(
		ctx,
		zitadel.New(host, opts...),
		client.WithAuth(auth),
	)
	if err != nil {
		return nil, ErrClientBuildFailed
	}

	return api, nil
}

// parseHost normalizes the configured domain into a bare host plus connection options.
// TLS is the default: a bare host ("zitadel.example.com") or an https:// URL always uses
// TLS, honoring an explicit port (e.g. self-hosted "zitadel.example.com:8443") via WithPort.
// Only an explicit http:// scheme opts into a plaintext, non-TLS connection via WithInsecure,
// which is intended for self-hosted or local development instances without TLS.
func parseHost(domain string) (string, []zitadel.Option) {
	raw := strings.TrimSpace(domain)

	insecure := strings.HasPrefix(raw, "http://")

	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimRight(raw, "/")

	// drop any trailing path so only host[:port] remains
	if i := strings.IndexByte(raw, '/'); i >= 0 {
		raw = raw[:i]
	}

	host := raw
	port := ""

	// SplitHostPort errors when no port is present, in which case the default port is used
	if h, p, err := net.SplitHostPort(raw); err == nil {
		host, port = h, p
	}

	switch {
	case insecure:
		// WithInsecure requires a port; fall back to the HTTP default when none is given
		if port == "" {
			port = "80"
		}

		return host, []zitadel.Option{zitadel.WithInsecure(port)}
	case port != "":
		if p, err := strconv.ParseUint(port, 10, 16); err == nil {
			return host, []zitadel.Option{zitadel.WithPort(uint16(p))}
		}
	}

	return host, nil
}

// resolveAuth selects the auth mode from the provided credential bindings, returning the
// instance domain and the matching Zitadel SDK token source initializer. PAT credentials take
// precedence when present, otherwise OAuth2 client-credentials are used.
func resolveAuth(bindings types.CredentialBindings) (string, client.TokenSourceInitializer, error) {
	if pat, ok, err := zitadelPATCredential.Resolve(bindings); err == nil && ok {
		if pat.Domain == "" {
			return "", nil, ErrDomainMissing
		}

		if pat.Token == "" {
			return "", nil, ErrTokenMissing
		}

		return pat.Domain, client.PAT(pat.Token), nil
	}

	if oauth, ok, err := zitadelOAuthCredential.Resolve(bindings); err == nil && ok {
		if oauth.Domain == "" {
			return "", nil, ErrDomainMissing
		}

		if oauth.ClientID == "" || oauth.ClientSecret == "" {
			return "", nil, ErrClientCredentialsMissing
		}

		auth := client.PasswordAuthentication(oauth.ClientID, oauth.ClientSecret, oidc.ScopeOpenID, client.ScopeZitadelAPI())

		return oauth.Domain, auth, nil
	}

	return "", nil, ErrCredentialDecode
}

// resolveDomain extracts the instance domain from whichever credential mode is configured,
// without requiring the full auth material. It is used by call sites that only need the domain.
func resolveDomain(bindings types.CredentialBindings) (string, bool) {
	if pat, ok, err := zitadelPATCredential.Resolve(bindings); err == nil && ok && pat.Domain != "" {
		return pat.Domain, true
	}

	if oauth, ok, err := zitadelOAuthCredential.Resolve(bindings); err == nil && ok && oauth.Domain != "" {
		return oauth.Domain, true
	}

	return "", false
}