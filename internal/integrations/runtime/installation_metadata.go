package runtime

import (
	"context"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
)

// SaveInstallationMetadata persists installation metadata for one installation
func SaveInstallationMetadata(ctx context.Context, installation *ent.Integration, metadata types.IntegrationInstallationMetadata) error {
	if len(metadata.Attributes) == 0 {
		return nil
	}

	if err := ent.FromContext(ctx).Integration.UpdateOneID(installation.ID).SetInstallationMetadata(metadata).Exec(ctx); err != nil {
		return err
	}

	installation.InstallationMetadata = metadata

	return nil
}
