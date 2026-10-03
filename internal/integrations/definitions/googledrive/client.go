package googledrive

import (
	"context"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// Client builds Google Drive SDK clients for one installation
type Client struct {
	// cfg is the operator-level Google Drive configuration
	cfg Config
}

// Build constructs the Google Drive SDK client for one installation
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (DriveClient, error) {
	cred, _, err := driveCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("Failed to resolve drive credentials")

		return DriveClient{}, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return DriveClient{}, ErrOAuthTokenMissing
	}

	ts := providerkit.GoogleTokenSource(context.Background(), c.cfg.ClientID, c.cfg.ClientSecret, providerkit.OAuthToken(cred.AccessToken, cred.RefreshToken, cred.Expiry))

	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("Failed to init drive client with provided credentials")

		return DriveClient{}, ErrDriveServiceBuildFailed
	}

	return DriveClient{Svc: svc}, nil
}
