package hooks

import (
	"context"
	"fmt"
	"strings"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
)

// requiredMutationString normalizes a mutation string field and validates it is present.
func requiredMutationString(fieldName string, value string, ok bool) (string, error) {
	normalized := strings.TrimSpace(value)
	if !ok || normalized == "" {
		return "", fmt.Errorf("%w: %s", ErrFieldRequired, fieldName)
	}

	return normalized, nil
}

// organizationDisplayNameByID loads and returns an organization's display name.
func organizationDisplayNameByID(ctx context.Context, client *generated.Client, orgID string) (string, error) {
	org, err := client.Organization.Query().
		Where(organization.ID(orgID)).
		Select(organization.FieldDisplayName).
		Only(ctx)
	if err != nil {
		return "", err
	}

	return org.DisplayName, nil
}
