package onedrive

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// clientBuilder returns the OneDrive Graph client builder refreshing tokens against app, bound to the operator config
func clientBuilder(cfg Config, app microsoft.AppConfig) func(context.Context, types.ConnectionRequest[oneDriveCred]) (*DriveClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[oneDriveCred]) (*DriveClient, error) {
		cred := req.Credential

		if cred.AccessToken == "" {
			return nil, ErrOAuthTokenMissing
		}

		graph, ts, err := app.NewDelegatedGraphClient(ctx, microsoft.CommonTenant, providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error building onedrive client")
			return nil, ErrClientBuildFailed
		}

		return &DriveClient{Graph: graph, TS: ts, Cfg: cfg}, nil
	}
}
