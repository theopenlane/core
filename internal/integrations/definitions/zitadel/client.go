package zitadel

import (
	"context"
	"net"
	"strconv"
	"strings"

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

// Build constructs the Zitadel user service client for one installation
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	cred, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if cred.Domain == "" {
		return nil, ErrDomainMissing
	}

	if cred.Token == "" {
		return nil, ErrTokenMissing
	}

	host, opts := parseHost(cred.Domain)

	api, err := client.New(
		ctx,
		zitadel.New(host, opts...),
		client.WithAuth(client.PAT(cred.Token)),
	)
	if err != nil {
		return nil, ErrClientBuildFailed
	}

	return api, nil
}

// parseHost normalizes the configured domain into a bare host and, when an explicit
// port is present (e.g. self-hosted "zitadel.example.com:8443"), a WithPort option.
// TLS is always required; plaintext HTTP instances are not supported.
func parseHost(domain string) (string, []zitadel.Option) {
	host := strings.TrimSpace(domain)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimRight(host, "/")

	// drop any trailing path so only host[:port] remains
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}

	// SplitHostPort errors when no port is present, in which case the default port is used
	if h, p, err := net.SplitHostPort(host); err == nil {
		if port, perr := strconv.ParseUint(p, 10, 16); perr == nil {
			return h, []zitadel.Option{zitadel.WithPort(uint16(port))}
		}
	}

	return host, nil
}

// resolveCredential extracts the CredentialSchema from the provided credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := zitadelCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrCredentialDecode
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialDecode
	}

	return cred, nil
}