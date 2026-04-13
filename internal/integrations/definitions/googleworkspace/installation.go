package googleworkspace

import (
	"context"

	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// resolveInstallationMetadata derives Google Workspace installation metadata from the credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := workspaceCredential.Resolve(req.Credentials)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googleworkspace: failed to resolve workspace credential")

		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return InstallationMetadata{}, false, nil
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cred.AccessToken})

	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googleworkspace: failed to create admin service, attempting to resolve installation metadata with token info")

		return InstallationMetadata{}, false, nil
	}

	if req.Integration.InstallationMetadata.Display.ExternalID == "" {
		logx.FromContext(ctx).Info().Msg("googleworkspace: no installation metadata external ID, attempting to resolve with token info")

		return InstallationMetadata{}, false, ErrCustomerIDMissing
	}

	customer, err := svc.Customers.Get(req.Integration.InstallationMetadata.Display.ExternalID).Context(ctx).Do()
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("googleworkspace: failed to fetch customer information")
		return InstallationMetadata{}, false, nil
	}

	if customer.Id == "" && customer.CustomerDomain == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		CustomerID: customer.Id,
		Domain:     customer.CustomerDomain,
	}, true, nil
}
