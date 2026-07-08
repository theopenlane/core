package zitadel

import (
	"context"
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

	host := strings.TrimPrefix(cred.Domain, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimRight(host, "/")

	api, err := client.New(
		ctx,
		zitadel.New(host),
		client.WithAuth(client.PAT(cred.Token)),
	)
	if err != nil {
		return nil, ErrClientBuildFailed
	}

	return api, nil
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