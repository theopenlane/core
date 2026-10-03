package zitadel

import (
	"cmp"
	"context"
	"strconv"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// zitadelDefaultPageSize is the number of records requested per Zitadel API page
	zitadelDefaultPageSize = 100
)

// Client builds Zitadel user service clients for one installation
type Client struct{}

// Build constructs the Zitadel user service client for one installation
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (*client.Client, error) {
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

// parseHost normalizes the instance into a bare host plus connection options
func parseHost(instance string) (string, []zitadel.Option) {
	parsed, err := urlx.Parse(instance)
	if err != nil {
		return instance, nil
	}

	host, err := urlx.NormalizeHostname(instance)
	if err != nil {
		return instance, nil
	}

	if parsed.Scheme == "http" {
		return host, []zitadel.Option{zitadel.WithInsecure(cmp.Or(parsed.Port(), "80"))}
	}

	if port, err := strconv.ParseUint(parsed.Port(), 10, 16); err == nil {
		return host, []zitadel.Option{zitadel.WithPort(uint16(port))}
	}

	return host, nil
}

// resolveAuth selects the auth mode and returns the domain and token source
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

// resolveDomain extracts the instance domain from whichever credential is configured
func resolveDomain(bindings types.CredentialBindings) (string, bool) {
	if pat, ok, err := zitadelPATCredential.Resolve(bindings); err == nil && ok && pat.Domain != "" {
		return pat.Domain, true
	}

	if oauth, ok, err := zitadelOAuthCredential.Resolve(bindings); err == nil && ok && oauth.Domain != "" {
		return oauth.Domain, true
	}

	return "", false
}
