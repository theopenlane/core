package googleworkspace

import (
	"context"
	"encoding/json"
	"errors"

	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/auth"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	// ClientGoogleWorkspaceAPI identifies the Google Workspace HTTP API client.
	ClientGoogleWorkspaceAPI types.ClientName = "api"
)

var errGoogleWorkspaceAdminServiceClientBuild = errors.New("googleworkspace: failed to build admin service client")

// googleWorkspaceClientDescriptors returns the client descriptors published by Google Workspace.
func googleWorkspaceClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeGoogleWorkspace, ClientGoogleWorkspaceAPI, "Google Workspace REST API client", buildGoogleWorkspaceClient)
}

// buildGoogleWorkspaceClient constructs an Admin SDK client from credential payload.
func buildGoogleWorkspaceClient(ctx context.Context, payload models.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token, err := auth.OAuthTokenFromPayload(payload)
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(svc), nil
}

// resolveGoogleWorkspaceClient returns a pooled Admin SDK client or builds one from the credential payload.
func resolveGoogleWorkspaceClient(ctx context.Context, input types.OperationInput) (*admin.Service, error) {
	if c, ok := types.ClientInstanceAs[*admin.Service](input.Client); ok {
		return c, nil
	}

	instance, err := buildGoogleWorkspaceClient(ctx, input.Credential, json.RawMessage(nil))
	if err != nil {
		return nil, err
	}

	client, ok := types.ClientInstanceAs[*admin.Service](instance)
	if !ok || client == nil {
		return nil, errGoogleWorkspaceAdminServiceClientBuild
	}

	return client, nil
}
