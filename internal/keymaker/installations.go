package keymaker

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// InstallationRecord captures the installation fields required by auth validation.
type InstallationRecord struct {
	ID           string
	OwnerID      string
	DefinitionID types.DefinitionID
}

// InstallationResolver resolves installations used during auth flow validation.
type InstallationResolver interface {
	ResolveInstallation(ctx context.Context, installationID string) (InstallationRecord, error)
}
