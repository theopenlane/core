package googleworkspace

import (
	"context"

	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives Google Workspace installation metadata from the credential
func resolveInstallationMetadata(ctx context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	cred, _, err := workspaceCredential.Resolve(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, ErrCredentialDecode
	}

	if cred.AccessToken == "" {
		return InstallationMetadata{}, false, nil
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cred.AccessToken})

	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return InstallationMetadata{}, false, nil
	}

	customer, err := svc.Customers.Get("my_customer").Context(ctx).Do()
	if err != nil {
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
