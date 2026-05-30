package graphapi

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/integrations/definitions/googledrive"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// LiveExternalContents is the resolver for the liveExternalContents field
func (r *internalPolicyResolver) LiveExternalContents(ctx context.Context, obj *generated.InternalPolicy) (*string, error) {
	if obj.ManagementMode != enums.DocumentManagementModeIntegration {
		return nil, nil
	}

	if obj.ExternalFileID == nil || *obj.ExternalFileID == "" {
		return nil, nil
	}

	if r.integrationsRuntime == nil {
		return nil, nil
	}

	ctx = privacy.DecisionContext(ctx, privacy.Allow)

	integ, err := r.findPrimaryDriveIntegration(ctx, obj.OwnerID)
	if err != nil || integ == nil {
		return nil, nil
	}

	op, err := r.integrationsRuntime.Registry().Operation(googledrive.DefinitionID(), googledrive.ExportOperationName())
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("google drive export operation not found in registry")
		return nil, nil
	}

	cfg, err := json.Marshal(googledrive.DocumentExport{FileID: *obj.ExternalFileID})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to marshal export config")
		return nil, nil
	}

	result, err := r.integrationsRuntime.ExecuteOperation(ctx, integ, op, nil, cfg)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("integration_id", integ.ID).Msg("drive export operation failed")
		return nil, nil
	}

	var export googledrive.DocumentExport
	if err := json.Unmarshal(result, &export); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to unmarshal export result")
		return nil, nil
	}

	obj.ExternalContents = &export.HTML

	return &export.HTML, nil
}

// findPrimaryDriveIntegration finds the Drive integration marked as primary for the given org
func (r *internalPolicyResolver) findPrimaryDriveIntegration(ctx context.Context, ownerID string) (*generated.Integration, error) {
	integrations, err := withTransactionalMutation(ctx).Integration.Query().
		Where(
			integration.OwnerID(ownerID),
			integration.DefinitionID(googledrive.DefinitionID()),
			integration.StatusEQ(enums.IntegrationStatusConnected),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query drive integrations")
		return nil, err
	}

	for _, integ := range integrations {
		var input googledrive.UserInput
		if err := jsonx.UnmarshalIfPresent(integ.Config.ClientConfig, &input); err != nil {
			continue
		}

		if input.Primary {
			return integ, nil
		}
	}

	if len(integrations) > 0 {
		return integrations[0], nil
	}

	return nil, nil
}
