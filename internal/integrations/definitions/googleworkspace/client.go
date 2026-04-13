package googleworkspace

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Client builds Google Workspace Admin SDK clients for one installation
type Client struct {
	// cfg is the operator-level Google Workspace configuration
	cfg Config
}

// Build constructs the Google Workspace Admin SDK client for one installation
func (c Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	cred, _, err := workspaceCredential.Resolve(req.Credentials)
	if err != nil {
		return nil, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

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
	}).TokenSource(ctx, tok)

	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, ErrAdminServiceBuildFailed
	}

	return svc, nil
}
