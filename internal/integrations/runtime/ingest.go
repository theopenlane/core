package runtime

import (
	"context"

	"github.com/samber/do/v2"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

// IngestOptions aliases the shared ingest options for runtime callers
type IngestOptions = operations.IngestOptions

// IngestPayloadSets processes mapped payload sets for one installation and operation through the shared ingest runtime
func (r *Runtime) IngestPayloadSets(ctx context.Context, db *ent.Client, installation *ent.Integration, operationName string, payloadSets []types.IngestPayloadSet, options IngestOptions) error {
	if installation == nil {
		return ErrInstallationRequired
	}

	operation, err := r.Registry().Operation(installation.DefinitionID, operationName)
	if err != nil {
		switch err {
		case registry.ErrDefinitionNotFound:
			return ErrDefinitionNotFound
		case registry.ErrOperationNotFound:
			return ErrOperationNotFound
		default:
			return err
		}
	}

	if db == nil {
		db = do.MustInvoke[*ent.Client](r.injector)
	}

	return operations.ProcessPayloadSetsWithOptions(ctx, r.Registry(), db, installation, operation.Ingest, payloadSets, options)
}
