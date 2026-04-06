package scim

import (
	"context"
	"errors"
	"fmt"

	scimerrors "github.com/elimity-com/scim/errors"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	definitionscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	integrationops "github.com/theopenlane/core/internal/integrations/operations"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

const (
	scimStatusBadRequest = 400
	scimStatusConflict   = 409
)

// ingestPayloadSets routes SCIM directory payloads through the standard ingest path
func ingestPayloadSets(ctx context.Context, client *generated.Client, rt *integrationsruntime.Runtime, integration *generated.Integration, payloadSets []integrationtypes.IngestPayloadSet) error {
	contracts := make([]integrationtypes.IngestContract, 0, len(payloadSets))
	for _, ps := range payloadSets {
		contracts = append(contracts, integrationtypes.IngestContract{Schema: ps.Schema})
	}

	return integrationops.ProcessPayloadSets(
		ctx,
		integrationops.IngestContext{
			Registry:    rt.Registry(),
			DB:          client,
			Runtime:     rt.Gala(),
			Integration: integration,
		},
		contracts,
		payloadSets,
		integrationops.IngestOptions{
			Source: integrationgenerated.IntegrationIngestSourceDirect,
		},
	)
}

// handleIngestError maps shared ingest failures to SCIM-compatible errors
func handleIngestError(err error, detail string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, definitionscim.ErrInvalidAttributes),
		errors.Is(err, integrationops.ErrIngestMappedDocumentInvalid),
		errors.Is(err, integrationops.ErrIngestUpsertKeyMissing):
		return scimerrors.ScimError{
			ScimType: scimerrors.ScimTypeInvalidValue,
			Detail:   detail,
			Status:   scimStatusBadRequest,
		}
	case errors.Is(err, integrationops.ErrIngestUpsertConflict):
		return scimerrors.ScimError{
			ScimType: scimerrors.ScimTypeUniqueness,
			Detail:   detail,
			Status:   scimStatusConflict,
		}
	default:
		return fmt.Errorf("directory ingest failed: %w", err)
	}
}
