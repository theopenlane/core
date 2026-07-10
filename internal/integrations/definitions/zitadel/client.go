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