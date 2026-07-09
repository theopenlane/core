package email

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"
	"github.com/theopenlane/newman"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/campaigntarget"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/integrations/templatekit"
	"github.com/theopenlane/core/pkg/logx"
)

// EnsureTrustCenterUpdateTemplate returns the id of the trust center's customizable update template,
// identified by TrustCenterUpdateTemplate, creating it when the trust center does not have one yet.
// Trust center creation and the startup backfill seed the template ahead of time; campaign creation
// calls this as the final guarantee since a campaign cannot send without its template
func EnsureTrustCenterUpdateTemplate(ctx context.Context, db *generated.Client, ownerID, trustCenterID string) (string, error) {
	existingID, err := db.EmailTemplate.Query().
		Where(
			emailtemplate.OwnerID(ownerID),
			emailtemplate.TrustCenterID(trustCenterID),
			emailtemplate.KeyEQ(TrustCenterUpdateTemplate),
		).
		FirstID(ctx)
	if err == nil {
		return existingID, nil
	}

	if !generated.IsNotFound(err) {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Str("trust_center_id", trustCenterID).Msg("failed querying for trust center update template")

		return "", err
	}

	created, err := db.EmailTemplate.Create().
		SetOwnerID(ownerID).
		SetTrustCenterID(trustCenterID).
		SetName(TrustCenterUpdateTemplate).
		SetKey(TrustCenterUpdateTemplate).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Str("trust_center_id", trustCenterID).Msg("failed creating trust center update template")

		return "", err
	}

	return created.ID, nil
}

// EnsureAllTrustCenterUpdateTemplates ensures every trust center has its update template, continuing
// past per-trust-center failures, and returns the number successfully ensured. Used by the startup
// backfill so organizations that pre-date template seeding get one ahead of any campaign send
func EnsureAllTrustCenterUpdateTemplates(ctx context.Context, db *generated.Client) (int, error) {
	trustCenters, err := db.TrustCenter.Query().All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed querying trust centers for update templates")

		return 0, err
	}

	ensured := 0

	// per-trust-center failures are logged inside the ensure and skipped so one broken trust
	// center does not block the rest of the backfill
	for _, tc := range trustCenters {
		if _, err := EnsureTrustCenterUpdateTemplate(ctx, db, tc.OwnerID, tc.ID); err != nil {
			continue
		}

		ensured++
	}

	return ensured, nil
}

// snapshotTrustCenterSubscribers materializes campaign targets from the trust center's active,
// verified, subscribed subscribers. It is idempotent: subscribers already represented by a target on
// the campaign are skipped. Running it inside the dispatch keeps a single source of truth so both the
// manual campaign launch and automated (post-publish, subprocessor change) triggers behave identically
func snapshotTrustCenterSubscribers(ctx context.Context, db *generated.Client, camp *generated.Campaign) error {
	if camp.CampaignType != enums.CampaignTypeTrustCenterUpdate || camp.TrustCenterID == "" {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	subscribers, err := db.Subscriber.Query().
		Where(
			subscriber.TrustCenterID(camp.TrustCenterID),
			subscriber.Active(true),
			subscriber.VerifiedEmail(true),
			subscriber.Unsubscribed(false),
		).
		All(allowCtx)
	if err != nil {
		return err
	}

	if len(subscribers) == 0 {
		return nil
	}

	existing, err := db.CampaignTarget.Query().
		Where(campaigntarget.CampaignIDEQ(camp.ID)).
		All(allowCtx)
	if err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(existing))
	for _, target := range existing {
		if target.SubscriberID != "" {
			seen[target.SubscriberID] = struct{}{}
		}
	}

	builders := make([]*generated.CampaignTargetCreate, 0, len(subscribers))
	for _, sub := range subscribers {
		if _, ok := seen[sub.ID]; ok {
			continue
		}

		builders = append(builders, db.CampaignTarget.Create().
			SetCampaignID(camp.ID).
			SetOwnerID(camp.OwnerID).
			SetEmail(sub.Email).
			SetSubscriberID(sub.ID).
			SetMetadata(map[string]any{MetadataUnsubscribeTokenKey: sub.Token}))
	}

	if len(builders) == 0 {
		return nil
	}

	return db.CampaignTarget.CreateBulk(builders...).Exec(allowCtx)
}

// renderMessagesForCampaign routes campaign rendering: trust center update campaigns brand the
// message from the trust center setting, all other campaigns render from the email template defaults
func renderMessagesForCampaign(ctx context.Context, client *Client, dispatcher Dispatcher, camp *generated.Campaign, template *generated.EmailTemplate, overlay CampaignContext, targets []*generated.CampaignTarget) ([]*newman.EmailMessage, []string, int) {
	if camp.CampaignType == enums.CampaignTypeTrustCenterUpdate {
		var setting *generated.TrustCenterSetting
		if camp.Edges.TrustCenter != nil {
			setting = camp.Edges.TrustCenter.Edges.Setting
		}

		return renderTrustCenterCampaignMessages(ctx, client, dispatcher, template, setting, camp.Metadata, overlay, targets)
	}

	return renderCampaignMessages(ctx, client, dispatcher, template.Defaults, camp.Metadata, overlay, targets)
}

// renderTrustCenterCampaignMessages builds a branded message per recipient. Content is the email
// template defaults overlaid with the campaign's per-send metadata (so automated triggers supply the
// post content via metadata over a shared template); branding comes from the trust center setting; the
// per-recipient unsubscribe token is resolved from each target's metadata
func renderTrustCenterCampaignMessages(ctx context.Context, client *Client, dispatcher Dispatcher, template *generated.EmailTemplate, setting *generated.TrustCenterSetting, metadata map[string]any, overlay CampaignContext, targets []*generated.CampaignTarget) ([]*newman.EmailMessage, []string, int) {
	overlays := make([]any, 0, 1)
	if len(metadata) > 0 {
		overlays = append(overlays, metadata)
	}

	payload, err := templatekit.BuildDispatchPayload(template.Defaults, overlays...)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed building trust center campaign content")

		return nil, nil, len(targets)
	}

	var base BrandedMessageRequest
	if err := json.Unmarshal(payload, &base); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed decoding trust center campaign content")

		return nil, nil, len(targets)
	}

	applyTrustCenterBranding(&base, setting, client.Config)

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

// applyTrustCenterBranding sets the branding fields on a branded message request from the trust
// center setting, falling back to the runtime email config for any value the trust center does not
// define. Content fields authored on the email template are left untouched
func applyTrustCenterBranding(req *BrandedMessageRequest, setting *generated.TrustCenterSetting, fallback RuntimeEmailConfig) {
	companyName := fallback.CompanyName
	logo := fallback.LogoURL
	primaryColor := fallback.HeadingColor
	buttonColor := fallback.ButtonColor
	bodyBackground := fallback.BodyBackgroundColor
	cardBackground := fallback.CardBackgroundColor
	textColor := fallback.TextColor

	if setting != nil {
		companyName = lo.CoalesceOrEmpty(setting.CompanyName, companyName)
		logo = lo.CoalesceOrEmpty(lo.FromPtr(setting.LogoRemoteURL), logo)
		primaryColor = lo.CoalesceOrEmpty(setting.PrimaryColor, primaryColor)
		buttonColor = lo.CoalesceOrEmpty(setting.AccentColor, buttonColor)
		bodyBackground = lo.CoalesceOrEmpty(setting.BackgroundColor, bodyBackground)
		cardBackground = lo.CoalesceOrEmpty(setting.SecondaryBackgroundColor, cardBackground)
		textColor = lo.CoalesceOrEmpty(setting.ForegroundColor, textColor)
	}

	req.CompanyName = companyName
	req.Corporation = lo.CoalesceOrEmpty(req.Corporation, fallback.Corporation)
	req.LogoURL = logo
	req.HeaderLogoURL = logo
	req.PrimaryColor = primaryColor
	req.ButtonColor = buttonColor
	req.BodyBackgroundColor = bodyBackground
	req.CardBackgroundColor = cardBackground
	req.TextColor = textColor
}
