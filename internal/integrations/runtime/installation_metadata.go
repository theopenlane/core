package runtime

import (
	"context"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
)

// saveInstallationMetadata persists installation metadata for one installation
func (r *Runtime) saveInstallationMetadata(ctx context.Context, installation *ent.Integration, metadata types.IntegrationInstallationMetadata) error {
	return r.DB().Integration.UpdateOneID(installation.ID).SetInstallationMetadata(metadata).Exec(ctx)
}
