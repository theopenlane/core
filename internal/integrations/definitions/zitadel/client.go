package zitadel

import (
	"cmp"
	"context"
	"net"
	"strconv"

	oidcclient "github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/zitadel-go/v3/pkg/client"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

const (
	// zitadelDefaultPageSize is the number of records requested per Zitadel API page
	zitadelDefaultPageSize = 100
)

// patClientBuilder returns the Personal Access Token client constructor, allowPrivate permits non-public instances for local development
func patClientBuilder(allowPrivate bool) func(context.Context, types.ConnectionRequest[CredentialSchema]) (Client, error) {
	return func(ctx context.Context, req types.ConnectionRequest[CredentialSchema]) (Client, error) {
		return patClient(ctx, req, allowPrivate)
	}
}

// oauthClientBuilder returns the OAuth client constructor, allowPrivate permits non-public instances for local development
func oauthClientBuilder(allowPrivate bool) func(context.Context, types.ConnectionRequest[OAuthCredentialSchema]) (Client, error) {
	return func(ctx context.Context, req types.ConnectionRequest[OAuthCredentialSchema]) (Client, error) {
		return oauthClient(ctx, req, allowPrivate)
	}
}

// patClient builds the Zitadel client from a Personal Access Token credential
func patClient(ctx context.Context, req types.ConnectionRequest[CredentialSchema], allowPrivate bool) (Client, error) {
	switch {
	case req.Credential.Domain == "":
		return Client{}, ErrDomainMissing
	case req.Credential.Token == "":
		return Client{}, ErrTokenMissing
	}

	return newClient(ctx, req.Credential.Domain, client.PAT(req.Credential.Token), allowPrivate)
}

// oauthClient builds the Zitadel client from an OAuth client-credentials credential
func oauthClient(ctx context.Context, req types.ConnectionRequest[OAuthCredentialSchema], allowPrivate bool) (Client, error) {
	switch {
	case req.Credential.Domain == "":
		return Client{}, ErrDomainMissing
	case req.Credential.ClientID == "" || req.Credential.ClientSecret == "":
		return Client{}, ErrClientCredentialsMissing
	}

	var auth client.TokenSourceInitializer
	if allowPrivate {
		auth = client.PasswordAuthentication(req.Credential.ClientID, req.Credential.ClientSecret, oidc.ScopeOpenID, client.ScopeZitadelAPI())
	} else {
		auth = publicPasswordAuthentication(req.Credential.ClientID, req.Credential.ClientSecret, oidc.ScopeOpenID, client.ScopeZitadelAPI())
	}

	return newClient(ctx, req.Credential.Domain, auth, allowPrivate)
}

// publicPasswordAuthentication matches client.PasswordAuthentication but sends discovery and token requests only to public hosts
func publicPasswordAuthentication(clientID, clientSecret string, scopes ...string) client.TokenSourceInitializer {
	return func(ctx context.Context, issuer string) (oauth2.TokenSource, error) {
		httpClient, err := urlx.NewCustomerHTTPClient(issuer, false)
		if err != nil {
			return nil, err
		}

		discovery, err := oidcclient.Discover(ctx, issuer, httpClient)
		if err != nil {
			return nil, err
		}

		config := &clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     discovery.TokenEndpoint,
			Scopes:       scopes,
		}

		return config.TokenSource(context.WithValue(ctx, oauth2.HTTPClient, httpClient)), nil
	}
}

// publicGRPCDial dials the Zitadel gRPC endpoint only when it resolves to a public address
func publicGRPCDial(ctx context.Context, address string) (net.Conn, error) {
	return urlx.PublicDialContext(ctx, "tcp", address)
}

// newClient constructs the Zitadel API client for the instance domain with the given token source
func newClient(ctx context.Context, domain string, auth client.TokenSourceInitializer, allowPrivate bool) (Client, error) {
	host, opts := parseHost(domain)

	clientOpts := []client.Option{client.WithAuth(auth)}

	// the instance domain is customer-supplied, so outside local development it must name a public host
	// and every connection is checked at dial time to cover DNS changes after this check
	if !allowPrivate {
		if _, err := urlx.ValidatePublicURL("https://" + host); err != nil {
			return Client{}, err
		}

		clientOpts = append(clientOpts, client.WithGRPCDialOptions(grpc.WithContextDialer(publicGRPCDial), grpc.WithNoProxy()))
	}

	api, err := client.New(
		ctx,
		zitadel.New(host, opts...),
		clientOpts...,
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
