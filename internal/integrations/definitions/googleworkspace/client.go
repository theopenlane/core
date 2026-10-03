package googleworkspace

import (
	"context"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Client builds Google Workspace Admin SDK clients for one installation
type Client struct {
	// cfg is the operator-level Google Workspace configuration
	cfg Config
}

// Build constructs the Google Workspace Admin SDK client for one installation
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (*admin.Service, error) {
	cred, _, err := workspaceCredential.Resolve(req.Credentials)
	if err != nil {
		return nil, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	svc, err := admin.NewService(ctx, option.WithTokenSource(providerkit.GoogleTokenSource(ctx, c.cfg.ClientID, c.cfg.ClientSecret, providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))))
	if err != nil {
		return nil, ErrAdminServiceBuildFailed
	}

	return svc, nil
}
