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
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
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
	// subprocessorNotificationSubject and the related copy back the subprocessor change system email
	subprocessorNotificationSubject    = "Subprocessor update"
	subprocessorNotificationTitle      = "We've updated our subprocessors"
	subprocessorNotificationIntro      = "The subprocessors we use have changed. The updates are listed below - review the full list anytime in our trust center."
	subprocessorNotificationButtonText = "View subprocessors"
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
	// the poller scans trust center settings and subprocessors across every organization, and the
	// per-trust-center sends it dispatches must load branding from the trust center setting. The
	// cross-org bypass rides on the caller (which gala persists across the durable dispatch boundary,
	// unlike the privacy decision), matching the other scheduled pollers' system caller
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

		// include soft-deleted rows so subprocessor removals are detected via their bumped updated_at, and
		// eager-load each row's subprocessor for the vendor name and logo
		changed, err := r.DB().TrustCenterSubprocessor.Query().
			Where(
				trustcentersubprocessor.TrustCenterID(setting.TrustCenterID),
				trustcentersubprocessor.UpdatedAtGT(floor),
				trustcentersubprocessor.UpdatedAtLTE(cutoff),
			).
			WithSubprocessor().
			All(entx.SkipSoftDelete(ctx))
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Msg("failed querying changed subprocessors")

			continue
		}

		if len(changed) == 0 {
			continue
		}

		latest := floor
		for _, sp := range changed {
			if sp.UpdatedAt.After(latest) {
				latest = sp.UpdatedAt
			}
		}

		tc, customDomain, err := r.loadTrustCenter(ctx, setting.TrustCenterID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Msg("failed loading trust center for subprocessor notification")

			continue
		}

		entries := subprocessorEntries(changed, floor)

		// the subprocessor notification is a system email, not a customizable campaign template: send it
		// directly to each active subscriber with their unsubscribe token, like every other system email
		subscribers, err := r.activeTrustCenterSubscribers(ctx, setting.TrustCenterID)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Msg("failed loading subscribers for subprocessor notification")

			continue
		}

		base := subprocessorNotificationRequest(setting, customDomain, tc.Slug, entries)

		for _, sub := range subscribers {
			req := base
			req.RecipientInfo = emaildef.RecipientInfo{
				Email:            sub.Email,
				UnsubscribeToken: sub.Token,
			}
			// the direct system dispatch does not run template interpolation, so resolve the per-recipient
			// unsubscribe link here with the subscriber's actual token
			req.UnsubscribeURL = trustcenterurl.UnsubscribeURLWithToken(customDomain, tc.Slug, sub.Token)

			if err := r.sendSubprocessorNotification(ctx, req); err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", setting.TrustCenterID).Str("subscriber_id", sub.ID).Msg("failed sending subprocessor notification")
			}
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
// template (the customizable message-updates template), creating it on first use. Per-send content is
// supplied via campaign metadata
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

// subprocessorEntries maps the changed trust center subprocessor join rows into the structured change
// entries rendered by the notification, reading the vendor name from each row's eager-loaded subprocessor
// edge and labeling the kind of change
func subprocessorEntries(changed []*ent.TrustCenterSubprocessor, floor time.Time) []emaildef.SubprocessorEntry {
	entries := make([]emaildef.SubprocessorEntry, 0, len(changed))

	for _, sp := range changed {
		vendor := sp.Edges.Subprocessor
		if vendor == nil {
			continue
		}

		entries = append(entries, emaildef.SubprocessorEntry{
			Name:   vendor.Name,
			Change: subprocessorChange(sp, floor),
		})
	}

	return entries
}

// subprocessorChange classifies a changed subprocessor join row relative to the last notification floor:
// a soft-deleted row is a removal, a row created after the floor is an addition, otherwise an update
func subprocessorChange(sp *ent.TrustCenterSubprocessor, floor time.Time) string {
	switch {
	case !sp.DeletedAt.IsZero():
		return "Removed"
	case sp.CreatedAt.After(floor):
		return "Added"
	default:
		return "Updated"
	}
}

// activeTrustCenterSubscribers returns the trust center's subscribers eligible for notifications:
// active, email-verified, and not unsubscribed
func (r *Runtime) activeTrustCenterSubscribers(ctx context.Context, trustCenterID string) ([]*ent.Subscriber, error) {
	return r.DB().Subscriber.Query().
		Where(
			subscriber.TrustCenterID(trustCenterID),
			subscriber.Active(true),
			subscriber.VerifiedEmail(true),
			subscriber.Unsubscribed(false),
		).
		All(ctx)
}

// subprocessorNotificationRequest builds the shared subprocessor notification body for a trust center:
// the controlled content plus the trust center branding overlay (logo, colors, company name) sourced
// from the setting. Empty branding values fall through to the Openlane system defaults at render time.
// RecipientInfo is set per subscriber by the caller
func subprocessorNotificationRequest(setting *ent.TrustCenterSetting, customDomain, slug string, entries []emaildef.SubprocessorEntry) emaildef.SubprocessorNotificationRequest {
	return emaildef.SubprocessorNotificationRequest{
		Subject:             subprocessorNotificationSubject,
		Title:               subprocessorNotificationTitle,
		Intros:              []string{subprocessorNotificationIntro},
		Subprocessors:       entries,
		ButtonText:          subprocessorNotificationButtonText,
		ButtonLink:          trustcenterurl.BuildURL(customDomain, slug),
		CompanyName:         setting.CompanyName,
		LogoURL:             lo.FromPtr(setting.LogoRemoteURL),
		PrimaryColor:        setting.PrimaryColor,
		ButtonColor:         setting.AccentColor,
		BodyBackgroundColor: setting.BackgroundColor,
		CardBackgroundColor: setting.SecondaryBackgroundColor,
		TextColor:           setting.ForegroundColor,
	}
}

// sendSubprocessorNotification dispatches the subprocessor notification to a single recipient as a
// runtime system email, mirroring how subscribe and other system emails are sent
func (r *Runtime) sendSubprocessorNotification(ctx context.Context, req emaildef.SubprocessorNotificationRequest) error {
	config, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = r.Dispatch(ctx, operations.DispatchRequest{
		DefinitionID: emaildef.DefinitionID.ID(),
		Operation:    emaildef.SubprocessorNotificationOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeScheduled,
		Runtime:      true,
	})

	return err
}
