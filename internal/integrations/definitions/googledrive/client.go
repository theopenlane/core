package googledrive

import (
	"context"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// clientBuilder returns the Google Drive client builder bound to the operator config
func clientBuilder(cfg Config) func(context.Context, types.ConnectionRequest[googleDriveCred]) (DriveClient, error) {
	return func(ctx context.Context, req types.ConnectionRequest[googleDriveCred]) (DriveClient, error) {
		cred := req.Credential

		if cred.AccessToken == "" {
			return DriveClient{}, ErrOAuthTokenMissing
		}

		ts := providerkit.GoogleTokenSource(context.Background(), cfg.ClientID, cfg.ClientSecret, providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))

		svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("Failed to init drive client with provided credentials")

			return DriveClient{}, ErrDriveServiceBuildFailed
		}

		return DriveClient{Svc: svc}, nil
	}
}
