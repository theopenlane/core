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

// patClient builds the Zitadel client from a Personal Access Token credential
func patClient(ctx context.Context, req types.ConnectionRequest[CredentialSchema]) (Client, error) {
	switch {
	case req.Credential.Domain == "":
		return Client{}, ErrDomainMissing
	case req.Credential.Token == "":
		return Client{}, ErrTokenMissing
	}

	return newClient(ctx, req.Credential.Domain, client.PAT(req.Credential.Token))
}

// oauthClient builds the Zitadel client from an OAuth client-credentials credential
func oauthClient(ctx context.Context, req types.ConnectionRequest[OAuthCredentialSchema]) (Client, error) {
	switch {
	case req.Credential.Domain == "":
		return Client{}, ErrDomainMissing
	case req.Credential.ClientID == "" || req.Credential.ClientSecret == "":
		return Client{}, ErrClientCredentialsMissing
	}

	auth := client.PasswordAuthentication(req.Credential.ClientID, req.Credential.ClientSecret, oidc.ScopeOpenID, client.ScopeZitadelAPI())

	return newClient(ctx, req.Credential.Domain, auth)
}

// newClient constructs the Zitadel API client for the instance domain with the given token source
func newClient(ctx context.Context, domain string, auth client.TokenSourceInitializer) (Client, error) {
	host, opts := parseHost(domain)

	api, err := client.New(
		ctx,
		zitadel.New(host, opts...),
		client.WithAuth(auth),
	)
	if err != nil {
		return Client{}, ErrClientBuildFailed
	}

	return Client{Client: api, Domain: domain}, nil
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
