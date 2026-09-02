//go:build codegen

// this file provides no-op stubs for the helpers in history_helpers.go so the package still
// compiles during code generation, when the generated cleanup functions they wrap may reference
// history packages that are not generated yet; hooks never execute during codegen so the
// stubs are never called
package hooks

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/generated"
)

func purgeOrganizationHistory(_ context.Context, _ string) error {
	return nil
}

func organizationEdgeCleanup(_ context.Context, _ string) error {
	return nil
}

func removeUserOrgScopedMemberships(_ context.Context, _ *generated.OrgMembershipMutation, _, _ string) error {
	return nil
}
