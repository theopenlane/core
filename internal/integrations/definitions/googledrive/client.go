package googledrive

import (
	"context"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// clientBuilder returns the Google Drive client builder refreshing tokens against app, legacy marking the retired drive.readonly connection
func clientBuilder(app AppConfig, legacy bool) func(context.Context, types.ConnectionRequest[googleDriveCred]) (DriveClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[googleDriveCred]) (DriveClient, error) {
		cred := req.Credential

		if cred.AccessToken == "" {
			return DriveClient{}, ErrOAuthTokenMissing
		}

		ts := providerkit.GoogleTokenSource(ctx, app.ClientID, app.ClientSecret, providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))

		svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("Failed to init drive client with provided credentials")

			return DriveClient{}, ErrDriveServiceBuildFailed
		}

		return DriveClient{Svc: svc, TokenSource: ts, Legacy: legacy}, nil
	}
}
