package googledrive

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// Client builds Google Drive SDK clients for one installation
type Client struct {
	// cfg is the operator-level Google Drive configuration
	cfg Config
}

// Build constructs the Google Drive SDK client for one installation
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	cred, _, err := driveCredential.Resolve(req.Credentials)
	if err != nil {
		return nil, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	logx.FromContext(ctx).Debug().
		Bool("has_refresh_token", cred.RefreshToken != "").
		Bool("has_expiry", cred.Expiry != nil).
		Msg("googledrive: building client")

	tok := &oauth2.Token{
		AccessToken:  cred.AccessToken,
		RefreshToken: cred.RefreshToken,
		TokenType:    "Bearer",
	}

	if cred.Expiry != nil {
		tok.Expiry = *cred.Expiry
	}

	ts := (&oauth2.Config{
		ClientID:     c.cfg.ClientID,
		ClientSecret: c.cfg.ClientSecret,
		Endpoint:     google.Endpoint,
	}).TokenSource(context.Background(), tok)

	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, ErrDriveServiceBuildFailed
	}

	return svc, nil
}
