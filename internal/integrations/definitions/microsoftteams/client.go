package microsoftteams

import (
	"context"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/microsoft"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// clientBuilder returns the Microsoft Graph client builder bound to the operator config
func clientBuilder(cfg Config) func(context.Context, types.ConnectionRequest[teamsCred]) (*msgraphsdk.GraphServiceClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[teamsCred]) (*msgraphsdk.GraphServiceClient, error) {
		cred := req.Credential

		if cred.AccessToken == "" {
			return nil, ErrOAuthTokenMissing
		}

		appCfg := microsoft.AppConfig{ApplicationID: cfg.ApplicationID, ClientSecret: cfg.ClientSecret}

		graph, _, err := appCfg.NewDelegatedGraphClient(ctx, microsoft.OrganizationsTenant, providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("error building microsoft teams client")
			return nil, ErrClientBuildFailed
		}

		return graph, nil
	}
}
