package zitadel

import (
	"context"
	"strings"

	"golang.org/x/oauth2"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	zitadelPkg "github.com/zitadel/zitadel-go/v3/pkg/client"
	zitadelSDK "github.com/zitadel/zitadel-go/v3/pkg/client/zitadel"
	zitadelUser "github.com/zitadel/zitadel-go/v3/pkg/client/user/v2"

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

	domain := strings.TrimPrefix(cred.Domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimRight(domain, "/")

	api := domain + ":443"

	client, err := zitadelUser.NewClient(
		ctx,
		cred.Domain,
		api,
		[]string{oidc.ScopeOpenID, zitadelPkg.ScopeZitadelAPI()},
		zitadelSDK.WithTokenSource(
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: cred.Token,
			}),
		),
	)
	if err != nil {
		return nil, ErrClientBuildFailed
	}

	return client, nil
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