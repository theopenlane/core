package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/emailtemplate"
	"github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessor"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/internal/trustcenterurl"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	// trustCenterNotificationGrace is the debounce window a post or subprocessor change must be stable
	// for before subscribers are notified, giving authors time to make further edits
	trustCenterNotificationGrace = time.Hour
	// trustCenterUpdateTemplateName is the display name of the reusable per-trust-center branded
	// template used to render automated subscriber notifications
	trustCenterUpdateTemplateName = "Trust Center Update"
)

// SeedTrustCenterNotifications starts the durable trust center notification polling loop after runtime
// listeners have been registered
func (r *Runtime) SeedTrustCenterNotifications(ctx context.Context) error {
	active, err := r.Gala().HasActiveJobForTopic(ctx, operations.TrustCenterNotificationTopic)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed checking for active trust center notification poller")

		return err
	}

	if active {
		logx.FromContext(ctx).Debug().Msg("trust center notification poller already active, skipping seed")

		return nil
	}

	receipt := r.Gala().EmitWithHeaders(ctx, operations.TrustCenterNotificationTopic, operations.TrustCenterNotificationEnvelope{}, gala.Headers{
		Tags: []string{"trustcenter-notifications"},
	})

	return receipt.Err
}

// HandleTrustCenterNotifications polls for trust center posts and subprocessor changes that have been
// stable for the grace window and dispatches a subscriber notification campaign for each. Returns the
// number dispatched as the delta for adaptive scheduling
func (r *Runtime) HandleTrustCenterNotifications(ctx context.Context, _ operations.TrustCenterNotificationEnvelope) (int, error) {
	now := time.Now()
	cutoff := now.Add(-trustCenterNotificationGrace)
	systemCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), &auth.Caller{
		Capabilities: auth.CapBypassOrgFilter | auth.CapBypassFGA | auth.CapInternalOperation,
	})

	return r.dispatchDuePosts(systemCtx, cutoff, now) + r.dispatchDueSubprocessorChanges(systemCtx, cutoff), nil
}

// dispatchDuePosts notifies subscribers about published posts flagged for notification that have been
// stable for the grace window
func (r *Runtime) dispatchDuePosts(ctx context.Context, cutoff, now time.Time) int {
	posts, err := r.DB().Note.Query().
		Where(
			note.NotifySubscribers(true),
			note.NotifiedAtIsNil(),
			note.TrustCenterIDNEQ(""),
			note.UpdatedAtLTE(cutoff),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed querying due trust center posts")

		return 0
	}

	dispatched := 0

	for _, post := range posts {
		tc, customDomain, err := r.loadTrustCenter(ctx, post.TrustCenterID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("note_id", post.ID).Msg("failed loading trust center for post notification")

			continue
		}

		title := lo.FromPtr(post.Title)
		if title == "" {
			title = "Trust center update"
		}

		content := trustCenterNotificationContent(title, title, []string{post.Text}, customDomain, tc.Slug)
		content["postID"] = post.ID

		if err := r.createAndDispatchTrustCenterCampaign(ctx, tc.OwnerID, post.TrustCenterID, title, content); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("note_id", post.ID).Msg("failed dispatching post notification")

			continue
		}

		if err := r.DB().Note.UpdateOneID(post.ID).SetNotifiedAt(now).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("note_id", post.ID).Msg("failed marking post notified")
		}

		dispatched++
	}

	return dispatched
}

// dispatchDueSubprocessorChanges notifies subscribers about subprocessor changes for trust centers
// that opted in, coalescing all changes since the last notification into one send per trust center
func (r *Runtime) dispatchDueSubprocessorChanges(ctx context.Context, cutoff time.Time) int {
	settings, err := r.DB().TrustCenterSetting.Query().
		Where(
			trustcentersetting.NotifySubscribersOnSubprocessorChange(true),
			trustcentersetting.TrustCenterIDNEQ(""),
		).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed querying trust center settings for subprocessor notifications")
		return 0
	}

	dispatched := 0

	for _, setting := range settings {
		floor := lo.FromPtr(setting.SubprocessorsNotifiedAt)

		// include soft-deleted rows so subprocessor removals are detected via their bumped updated_at
		changed, err := r.DB().TrustCenterSubprocessor.Query().
			Where(
				trustcentersubprocessor.TrustCenterID(setting.TrustCenterID),
				trustcentersubprocessor.UpdatedAtGT(floor),
				trustcentersubprocessor.UpdatedAtLTE(cutoff),
			).
			All(entx.SkipSoftDelete(ctx))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Msg("failed querying changed subprocessors")

			continue
		}

		if len(changed) == 0 {
			continue
		}

		latest := floor
		ids := make([]string, 0, len(changed))

		for _, sp := range changed {
			ids = append(ids, sp.ID)

			if sp.UpdatedAt.After(latest) {
				latest = sp.UpdatedAt
			}
		}

		tc, customDomain, err := r.loadTrustCenter(ctx, setting.TrustCenterID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Msg("failed loading trust center for subprocessor notification")

			continue
		}

		content := trustCenterNotificationContent(
			"Subprocessor update",
			"We've updated our subprocessors",
			[]string{"Our subprocessor list has changed. Review the latest information in our trust center."},
			customDomain,
			tc.Slug,
		)
		content["subprocessorIDs"] = ids

		if err := r.createAndDispatchTrustCenterCampaign(ctx, tc.OwnerID, setting.TrustCenterID, "Subprocessor update", content); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Msg("failed dispatching subprocessor notification")

			continue
		}

		if err := r.DB().TrustCenterSetting.UpdateOneID(setting.ID).SetSubprocessorsNotifiedAt(latest).Exec(ctx); err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Msg("failed updating subprocessor notified watermark")
		}

		dispatched++
	}

	return dispatched
}

// loadTrustCenter loads a trust center with its custom domain and returns the resolved custom domain
// (empty when none)
func (r *Runtime) loadTrustCenter(ctx context.Context, trustCenterID string) (*ent.TrustCenter, string, error) {
	tc, err := r.DB().TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenterID)).
		WithCustomDomain().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed loading trust center")

		return nil, "", err
	}

	customDomain := ""
	if tc.Edges.CustomDomain != nil {
		customDomain = tc.Edges.CustomDomain.CnameRecord
	}

	return tc, customDomain, nil
}

// trustCenterNotificationContent builds the campaign metadata content (overlaid on the branded
// template at render time) including the per-recipient unsubscribe link and a trust center button
func trustCenterNotificationContent(subject, title string, intros []string, customDomain, slug string) map[string]any {
	content := map[string]any{
		"subject": subject,
		"title":   title,
		"intros":  intros,
	}

	if url := trustcenterurl.UnsubscribeURL(customDomain, slug); url != "" {
		content["unsubscribeURL"] = url
	}

	if link := trustcenterurl.BuildURL(customDomain, slug); link != "" {
		content["buttonText"] = "View trust center"
		content["buttonLink"] = link
	}

	return content
}

// createAndDispatchTrustCenterCampaign creates a trust center update campaign carrying the supplied
// content and dispatches it through the campaign send, which materializes subscriber targets
func (r *Runtime) createAndDispatchTrustCenterCampaign(ctx context.Context, ownerID, trustCenterID, name string, content map[string]any) error {
	templateID, err := r.ensureTrustCenterUpdateTemplate(ctx, ownerID, trustCenterID)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Str("trust_center_id", trustCenterID).Msg("failed ensuring trust center update template")

		return err
	}

	camp, err := r.DB().Campaign.Create().
		SetOwnerID(ownerID).
		SetName(name).
		SetCampaignType(enums.CampaignTypeTrustCenterUpdate).
		SetTrustCenterID(trustCenterID).
		SetEmailTemplateID(templateID).
		SetMetadata(content).
		Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed creating campaign for trust center notification")

		return err
	}

	config, err := json.Marshal(emaildef.SendBrandedCampaignRequest{
		CampaignDispatchInput: emaildef.CampaignDispatchInput{CampaignID: camp.ID},
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", camp.ID).Msg("failed marshaling campaign dispatch config")

		return err
	}

	integrationID, err := r.ResolveOwnerIntegration(ctx, emaildef.DefinitionID.ID(), ownerID, func(inst *ent.Integration) bool {
		return inst.CampaignEmail
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Msg("failed resolving integration for trust center notification")

		return err
	}

	_, err = r.Dispatch(ctx, operations.DispatchRequest{
		IntegrationID: integrationID,
		DefinitionID:  emaildef.DefinitionID.ID(),
		OwnerID:       ownerID,
		Operation:     emaildef.SendCampaignOp.Name(),
		Config:        config,
		RunType:       enums.IntegrationRunTypeScheduled,
		Runtime:       integrationID == "",
	})

	return err
}

// ensureTrustCenterUpdateTemplate returns the id of the trust center's reusable branded update
// template, creating it on first use. Per-send content is supplied via campaign metadata
func (r *Runtime) ensureTrustCenterUpdateTemplate(ctx context.Context, ownerID, trustCenterID string) (string, error) {
	existing, err := r.DB().EmailTemplate.Query().
		Where(
			emailtemplate.OwnerID(ownerID),
			emailtemplate.TrustCenterID(trustCenterID),
			emailtemplate.Key(emaildef.BrandedMessageOp.Name()),
		).
		First(ctx)
	if err == nil {
		return existing.ID, nil
	}

	if !ent.IsNotFound(err) {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Str("trust_center_id", trustCenterID).Msg("failed querying for existing trust center update template")

		return "", err
	}

	created, err := r.DB().EmailTemplate.Create().
		SetOwnerID(ownerID).
		SetTrustCenterID(trustCenterID).
		SetName(trustCenterUpdateTemplateName).
		SetKey(emaildef.BrandedMessageOp.Name()).
		SetTemplateContext(enums.TemplateContextCampaignRecipient).
		Save(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Str("trust_center_id", trustCenterID).Msg("failed creating trust center update template")

		return "", err
	}

	return created.ID, nil
}
