package hooks

import (
	"context"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
)

// organizationDisplayNameByID loads and returns an organization's display name.
func organizationDisplayNameByID(ctx context.Context, client *generated.Client, orgID string) (string, error) {
	orgDisplayName, err := client.Organization.Query().
		Where(organization.ID(orgID)).
		Select(organization.FieldDisplayName).
		String(ctx)
	if err != nil {
		return "", err
	}
	return orgDisplayName, nil
}
