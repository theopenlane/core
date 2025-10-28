package scim

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// ValidateSSOEnforced checks if SSO is enforced for the organization
// SCIM provisioning requires SSO to be enforced since SCIM users authenticate via SSO
func ValidateSSOEnforced(ctx context.Context, orgID string) error {
	client := transaction.FromContext(ctx)

	org, err := client.Organization.Query().
		Where(organization.ID(orgID)).
		WithSetting().
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return fmt.Errorf("%w: organization %s", ErrOrgNotFound, orgID)
		}
		return fmt.Errorf("failed to query organization: %w", err)
	}

	if org.Edges.Setting == nil {
		return fmt.Errorf("%w for organization %s", ErrOrgSettingsNotFound, orgID)
	}

	if !org.Edges.Setting.IdentityProviderLoginEnforced {
		return ErrSSONotEnforced
	}

	return nil
}
