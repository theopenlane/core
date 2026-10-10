package googleworkspace

import (
	"context"
	"fmt"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// verify resolves the Google Workspace customer the credential is scoped to
func verify(ctx context.Context, _ types.ConnectionRequest[googleWorkspaceCred], svc *admin.Service) (InstallationMetadata, error) {
	customer, err := svc.Customers.Get(defaultCustomerID).Context(ctx).Do()
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("googleworkspace: customer lookup failed")

		return InstallationMetadata{}, fmt.Errorf("%w: %v", ErrCustomerFetchFailed, err)
	}

	if customer.Id == "" && customer.CustomerDomain == "" {
		return InstallationMetadata{}, ErrCustomerUnresolved
	}

	return InstallationMetadata{
		CustomerID: customer.Id,
		Domain:     customer.CustomerDomain,
	}, nil
}
