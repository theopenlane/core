package googleworkspace

import (
	"context"

	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Client builds Google Workspace Admin SDK clients for one installation
type Client struct{}

// Build constructs the Google Workspace Admin SDK client for one installation
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	cred, _, err := workspaceCredential.Resolve(req.Credentials)
	if err != nil {
		return nil, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return nil, ErrOAuthTokenMissing
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cred.AccessToken})

	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, ErrAdminServiceBuildFailed
	}

	return svc, nil
}
