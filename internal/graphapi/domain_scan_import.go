package graphapi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/graphapi/model"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/cloudflare"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

var (
	errDomainScanImportURLInvalid     = errors.New("domain scan import: invalid URL")
	errDomainScanImportDispatchFailed = errors.New("domain scan import: failed to queue domain scan")
)

func (r *mutationResolver) requestDomainImport(ctx context.Context, input model.RequestDomainImportInput) (*model.RequestDomainImportPayload, error) {
	if r.integrationsRuntime == nil {
		return nil, nil
	}

	target, err := urlx.Parse(input.URL)
	if err != nil {
		return nil, errDomainScanImportURLInvalid
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return nil, rout.NewMissingRequiredFieldError("owner_id")
	}

	config, err := json.Marshal(cloudflare.DomainScanRequest{
		OrganizationID: orgID,
		Domain:         target.String(),
		ForceRefresh:   true,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan import: failed to encode domain scan request")

		return nil, errDomainScanImportDispatchFailed
	}

	_, err = r.integrationsRuntime.Dispatch(ctx, types.DispatchRequest{
		DefinitionID: cloudflare.DefinitionID.ID(),
		OwnerID:      orgID,
		Operation:    cloudflare.DomainScanRequestOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeEvent,
		Runtime:      true,
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("domain scan import: failed to queue domain scan")

		return nil, errDomainScanImportDispatchFailed
	}

	return &model.RequestDomainImportPayload{
		Accepted: true,
	}, nil
}
