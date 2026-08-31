package email

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/integrations/templatekit"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// snapshotTrustCenterSubscribers materializes campaign targets from the trust center's active,
// verified, subscribed subscribers. It is idempotent: subscribers already represented by a target on
// the campaign are skipped. Running it inside the dispatch keeps a single source of truth so both the
// manual campaign launch and automated (post-publish, subprocessor change) triggers behave identically
func snapshotTrustCenterSubscribers(ctx context.Context, db *generated.Client, camp *generated.Campaign) error {
	if camp.CampaignType != enums.CampaignTypeTrustCenterUpdate || camp.TrustCenterID == "" {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	return snapshotCampaignRecipients(allowCtx, db, camp, func(handle campaignRecipientHandlerFunc) error {
		return resolveTrustCenterSubscriberRecipients(allowCtx, db, camp, handle)
	})
}

// renderMessagesForCampaign routes campaign rendering: trust center update campaigns render the
// system message from the campaign's per-send metadata branded from the trust center setting, all
// other campaigns render from their email template defaults
func renderMessagesForCampaign(ctx context.Context, client *Client, dispatcher Dispatcher, camp *generated.Campaign, template *generated.EmailTemplate, overlay CampaignContext, targets []*generated.CampaignTarget) ([]*newman.EmailMessage, []string, int) {
	if camp.CampaignType == enums.CampaignTypeTrustCenterUpdate {
		var setting *generated.TrustCenterSetting
		if camp.Edges.TrustCenter != nil {
			setting = camp.Edges.TrustCenter.Edges.Setting
		}

		return renderTrustCenterCampaignMessages(ctx, client, dispatcher, setting, camp.Metadata, overlay, targets)
	}

	return renderCampaignMessages(ctx, client, dispatcher, template.Defaults, camp.Metadata, overlay, targets)
}

// renderTrustCenterCampaignMessages builds an update message per recipient. The post content comes
// from the campaign's per-send metadata (the automated triggers supply it when they create the
// campaign); branding comes from the trust center setting with the system config as fallback at
// render time; the per-recipient unsubscribe token is resolved from each target's metadata
func renderTrustCenterCampaignMessages(ctx context.Context, client *Client, dispatcher Dispatcher, setting *generated.TrustCenterSetting, metadata map[string]any, overlay CampaignContext, targets []*generated.CampaignTarget) ([]*newman.EmailMessage, []string, int) {
	payload, err := templatekit.BuildDispatchPayload(metadata)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed building trust center campaign content")

		return nil, nil, len(targets)
	}

	var base TrustCenterUpdateRequest
	if err := json.Unmarshal(payload, &base); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed decoding trust center campaign content")

		return nil, nil, len(targets)
	}

	base.TrustCenterBranding = TrustCenterBrandingFromSetting(setting)

	messages := make([]*newman.EmailMessage, 0, len(targets))
	targetIDs := make([]string, 0, len(targets))
	failed := 0

	for _, target := range targets {
		first, last := splitFullName(target.FullName)

		req := base
		req.CampaignContext = overlay
		req.RecipientInfo = RecipientInfo{
			Email:            target.Email,
			FirstName:        first,
			LastName:         last,
			UnsubscribeToken: unsubscribeTokenFromMetadata(target.Metadata),
		}

		msgPayload, err := json.Marshal(req)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", target.ID).Msg("failed marshaling trust center message")
			failed++

			continue
		}

		msg, err := dispatcher.RenderMessage(ctx, client, msgPayload, newman.WithTag(newman.Tag{Name: TagCampaignTargetID, Value: target.ID}))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("target_id", target.ID).Msg("failed rendering trust center message")
			failed++

			continue
		}

		messages = append(messages, msg)
		targetIDs = append(targetIDs, target.ID)
	}

	return messages, targetIDs, failed
}
