package googleworkspace

import (
	"context"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// clientBuilder returns the Google Workspace Admin SDK client builder bound to the operator config
func clientBuilder(cfg Config) func(context.Context, types.ConnectionRequest[googleWorkspaceCred]) (*admin.Service, error) {
	return func(ctx context.Context, req types.ConnectionRequest[googleWorkspaceCred]) (*admin.Service, error) {
		cred := req.Credential

		if cred.AccessToken == "" {
			return nil, ErrOAuthTokenMissing
		}

		svc, err := admin.NewService(ctx, option.WithTokenSource(providerkit.GoogleTokenSource(ctx, cfg.ClientID, cfg.ClientSecret, providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))))
		if err != nil {
			return nil, ErrAdminServiceBuildFailed
		}

		return svc, nil
	}
}
