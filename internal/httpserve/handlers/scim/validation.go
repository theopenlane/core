package scim

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/pkg/middleware/transaction"
)

// ValidateSSOEnforced checks if SSO is enforced for the organization
// SCIM provisioning requires SSO to be enforced since SCIM users authenticate via SSO
func ValidateSSOEnforced(ctx context.Context, orgID string) error {
	client := transaction.FromContext(ctx)

	orgSetting, err := client.OrganizationSetting.Query().
		Where(organizationsetting.OrganizationID(orgID)).
		Only(ctx)
	if err != nil {
		if generated.IsNotFound(err) {
			return fmt.Errorf("%w: organization %s", ErrOrgNotFound, orgID)
		}

		return fmt.Errorf("failed to query organization: %w", err)
	}

	if !orgSetting.IdentityProviderLoginEnforced {
		return ErrSSONotEnforced
	}

	return nil
}

// ValidateSCIMMode enforces SSO when the provision mode creates real User entities.
// Directory-only mode skips SSO enforcement since no logins are provisioned
func ValidateSCIMMode(ctx context.Context, orgID string, mode enums.SCIMProvisionMode) error {
	if mode == enums.SCIMProvisionModeDirectory {
		return nil
	}

	return ValidateSSOEnforced(ctx, orgID)
}
