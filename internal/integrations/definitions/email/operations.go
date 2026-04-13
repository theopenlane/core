package email

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// handleSendCampaign dispatches a full campaign by loading campaign data and recipients
// from the database and sending individual emails through the configured sender
var handleSendCampaign = providerkit.WithClientRequestConfig(emailClientRef, sendCampaignOp, ErrTemplateRenderFailed, func(ctx context.Context, req types.OperationRequest, client *EmailClient, campaignReq SendCampaignRequest) (json.RawMessage, error) {
	if err := SendCampaignEmails(ctx, req.DB, client, campaignReq.CampaignID); err != nil {
		return nil, err
	}

	return nil, nil
})

