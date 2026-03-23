package runtime

import (
	"context"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
)

// saveInstallationMetadata persists installation metadata for one installation
func (r *Runtime) saveInstallationMetadata(ctx context.Context, installation *ent.Integration, metadata types.IntegrationInstallationMetadata) error {
	db := r.DB()

	if len(metadata.Attributes) == 0 {
		if len(installation.InstallationMetadata.Attributes) == 0 {
			return nil
		}

		if err := db.Integration.UpdateOneID(installation.ID).ClearInstallationMetadata().Exec(ctx); err != nil {
			return err
		}

		installation.InstallationMetadata = types.IntegrationInstallationMetadata{}

		return nil
	}

	if err := db.Integration.UpdateOneID(installation.ID).SetInstallationMetadata(metadata).Exec(ctx); err != nil {
		return err
	}

	installation.InstallationMetadata = metadata

	return nil
}
