package googleworkspace

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// Client builds Google Workspace Admin SDK clients for one installation
type Client struct {
	// cfg is the operator-level Google Workspace configuration
	cfg Config
}

// tokenSource builds a refreshing OAuth2 token source for a Google Workspace credential using the
// operator's registered OAuth client id and secret, so a stored access token is refreshed via the
// stored refresh token and expiry rather than reused statically until it expires
func tokenSource(ctx context.Context, cfg Config, cred googleWorkspaceCred) oauth2.TokenSource {
	tok := &oauth2.Token{
		AccessToken:  cred.AccessToken,
		RefreshToken: cred.RefreshToken,
		TokenType:    "Bearer",
	}

	if cred.Expiry != nil {
		tok.Expiry = *cred.Expiry
	}

	return (&oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     google.Endpoint,
	}).TokenSource(ctx, tok)
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

	svc, err := admin.NewService(ctx, option.WithTokenSource(tokenSource(ctx, c.cfg, cred)))
	if err != nil {
		return nil, ErrAdminServiceBuildFailed
	}

	return svc, nil
}
