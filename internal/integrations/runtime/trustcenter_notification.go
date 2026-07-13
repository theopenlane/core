package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	ent "github.com/theopenlane/core/internal/ent/generated"
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

// trustCenterNotificationGrace is the debounce window a post or subprocessor change must be stable
// for before subscribers are notified, giving authors time to make further edits
const trustCenterNotificationGrace = time.Hour

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
// number dispatched as the delta for adaptive scheduling; a returned error feeds the scheduler's
// error-streak backoff while per-item failures inside each sweep only log and continue
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

	posts, postsErr := r.dispatchDuePosts(systemCtx, cutoff, now)
	subprocessors, subprocessorsErr := r.dispatchDueSubprocessorChanges(systemCtx, cutoff)

	return posts + subprocessors, errors.Join(postsErr, subprocessorsErr)
}

// dispatchDuePosts notifies subscribers about published posts flagged for notification that have been
// stable for the grace window. Per-post failures continue the sweep so one broken trust center does
// not block the rest; every failure is joined into the returned error for the scheduler's backoff
func (r *Runtime) dispatchDuePosts(ctx context.Context, cutoff, now time.Time) (int, error) {
	posts, err := r.DB().Note.Query().
		Where(
			note.NotifySubscribers(true),
			note.NotifiedAtIsNil(),
			note.TrustCenterIDNotNil(),
			note.UpdatedAtLTE(cutoff),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying due trust center posts: %w", err)
	}

	dispatched := 0

	var errs []error

	for _, post := range posts {
		tc, customDomain, err := r.loadTrustCenter(ctx, post.TrustCenterID)
		if err != nil {
			errs = append(errs, fmt.Errorf("loading trust center for post %s: %w", post.ID, err))

			continue
		}

		// the campaign metadata carries only the post data; the trust center update operation composes
		// the subject and body copy and the branding is resolved from the trust center setting at render
		title := lo.FromPtr(post.Title)

		content, err := emaildef.TrustCenterUpdateContent(emaildef.TrustCenterUpdateRequest{
			PostTitle:      title,
			PostText:       post.Text,
			TrustCenterURL: trustcenterurl.BuildURL(customDomain, tc.Slug),
			UnsubscribeURL: trustcenterurl.UnsubscribeURL(customDomain, tc.Slug),
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("building notification content for post %s: %w", post.ID, err))

			continue
		}

		content["postID"] = post.ID

		name := lo.CoalesceOrEmpty(title, emaildef.DefaultTrustCenterUpdateTitle)

		if err := r.createAndDispatchTrustCenterCampaign(ctx, tc.OwnerID, post.TrustCenterID, name, content); err != nil {
			errs = append(errs, fmt.Errorf("dispatching notification for post %s: %w", post.ID, err))

			continue
		}

		if err := r.DB().Note.UpdateOneID(post.ID).SetNotifiedAt(now).Exec(ctx); err != nil {
			errs = append(errs, fmt.Errorf("marking post %s notified: %w", post.ID, err))
		}

		dispatched++
	}

	return dispatched, errors.Join(errs...)
}

// dispatchDueSubprocessorChanges notifies subscribers about subprocessor changes for trust centers
// that opted in, coalescing all changes since the last notification into one send per trust center.
// Per-trust-center failures continue the sweep so one broken trust center does not block the rest;
// every failure is joined into the returned error for the scheduler's backoff
func (r *Runtime) dispatchDueSubprocessorChanges(ctx context.Context, cutoff time.Time) (int, error) {
	settings, err := r.DB().TrustCenterSetting.Query().
		Where(
			trustcentersetting.NotifySubscribersOnSubprocessorChange(true),
			trustcentersetting.TrustCenterIDNotNil(),
			trustcentersetting.EnvironmentEQ(enums.TrustCenterEnvironmentLive),
		).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("querying trust center settings for subprocessor notifications: %w", err)
	}

	dispatched := 0

	var errs []error

	for _, setting := range settings {
		// a zero floor would treat the trust center's entire subprocessor history as changes, so establish the
		// baseline at the cutoff and notify from the next change instead
		if setting.SubprocessorsNotifiedAt == nil {
			if err := r.DB().TrustCenterSetting.UpdateOneID(setting.ID).SetSubprocessorsNotifiedAt(cutoff).Exec(ctx); err != nil {
				errs = append(errs, fmt.Errorf("initializing subprocessor notified baseline for trust center %s: %w", setting.TrustCenterID, err))
			}

			continue
		}

		floor := *setting.SubprocessorsNotifiedAt

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
			errs = append(errs, fmt.Errorf("querying changed subprocessors for trust center %s: %w", setting.TrustCenterID, err))

			continue
		}

		if len(changed) == 0 {
			continue
		}

		latest := lo.MaxBy(changed, func(a, b *ent.TrustCenterSubprocessor) bool {
			return a.UpdatedAt.After(b.UpdatedAt)
		}).UpdatedAt

		tc, customDomain, err := r.loadTrustCenter(ctx, setting.TrustCenterID)
		if err != nil {
			errs = append(errs, fmt.Errorf("loading trust center %s for subprocessor notification: %w", setting.TrustCenterID, err))

			continue
		}

		entries := subprocessorEntries(changed, floor)

		// churn that nets to no change (e.g. a vendor added and removed within the window) leaves nothing
		// to report: advance the baseline past the processed rows without emailing
		if len(entries) == 0 {
			if err := r.DB().TrustCenterSetting.UpdateOneID(setting.ID).SetSubprocessorsNotifiedAt(latest).Exec(ctx); err != nil {
				errs = append(errs, fmt.Errorf("advancing subprocessor notified baseline for trust center %s: %w", setting.TrustCenterID, err))
			}

			continue
		}

		// the subprocessor notification is a direct system email: send it to each active subscriber
		// with their unsubscribe token, like every other system email
		subscribers, err := r.activeTrustCenterSubscribers(ctx, setting.TrustCenterID)
		if err != nil {
			errs = append(errs, fmt.Errorf("loading subscribers for trust center %s subprocessor notification: %w", setting.TrustCenterID, err))

			continue
		}

		// the request carries only data; the subprocessor operation composes the subject and body copy
		// from the branding's company name with defined fallbacks
		base := emaildef.SubprocessorNotificationRequest{
			TrustCenterBranding: emaildef.TrustCenterBrandingFromSetting(setting),
			Subprocessors:       entries,
			TrustCenterURL:      trustcenterurl.BuildURL(customDomain, tc.Slug),
		}

		// a dead logo URL renders a broken image, so fall back to the default logo when it does not load
		if base.LogoURL != "" && !logoURLReachable(ctx, base.LogoURL) {
			logx.FromContext(ctx).Warn().Str("trust_center_id", setting.TrustCenterID).Str("logo_url", base.LogoURL).Msg("trust center logo URL unreachable, using default logo")

			base.LogoURL = ""
		}

		failedSends := 0

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
				errs = append(errs, fmt.Errorf("sending subprocessor notification to subscriber %s for trust center %s: %w", sub.ID, setting.TrustCenterID, err))

				failedSends++
			}
		}

		// a total send failure keeps the baseline so the window retries next poll; partial failures
		// advance it to avoid re-sending to recipients that succeeded
		if len(subscribers) > 0 && failedSends == len(subscribers) {
			continue
		}

		if err := r.DB().TrustCenterSetting.UpdateOneID(setting.ID).SetSubprocessorsNotifiedAt(latest).Exec(ctx); err != nil {
			errs = append(errs, fmt.Errorf("advancing subprocessor notified baseline for trust center %s: %w", setting.TrustCenterID, err))
		}

		dispatched++
	}

	return dispatched, errors.Join(errs...)
}

// loadTrustCenter loads a trust center with its custom domain and returns the resolved custom
// domain (empty when none)
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

// createAndDispatchTrustCenterCampaign creates a trust center update campaign carrying the supplied
// content and dispatches it through the campaign send, which materializes subscriber targets. The
// message is fully composed in the campaign metadata and rendered through the system trust center
// update operation, so no email template is linked
func (r *Runtime) createAndDispatchTrustCenterCampaign(ctx context.Context, ownerID, trustCenterID, name string, content map[string]any) error {
	camp, err := r.DB().Campaign.Create().
		SetOwnerID(ownerID).
		SetName(name).
		SetCampaignType(enums.CampaignTypeTrustCenterUpdate).
		SetTrustCenterID(trustCenterID).
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

// subprocessorEntries coalesces the changed trust center subprocessor join rows per subprocessor and
// maps each vendor's net change relative to the last notification floor into a structured change entry,
// reading the vendor name from the eager-loaded subprocessor edge. Coalescing keeps a vendor whose row
// was deleted and recreated within one window (e.g. a full list re-import) from rendering as both a
// removal and an addition
func subprocessorEntries(changed []*ent.TrustCenterSubprocessor, floor time.Time) []emaildef.SubprocessorEntry {
	withVendor := lo.Filter(changed, func(sp *ent.TrustCenterSubprocessor, _ int) bool {
		return sp.Edges.Subprocessor != nil
	})

	groups := lo.GroupBy(withVendor, func(sp *ent.TrustCenterSubprocessor) string {
		return sp.SubprocessorID
	})

	return lo.FilterMap(lo.UniqBy(withVendor, func(sp *ent.TrustCenterSubprocessor) string {
		return sp.SubprocessorID
	}), func(sp *ent.TrustCenterSubprocessor, _ int) (emaildef.SubprocessorEntry, bool) {
		change, ok := subprocessorNetChange(groups[sp.SubprocessorID], floor)
		if !ok {
			return emaildef.SubprocessorEntry{}, false
		}

		return emaildef.SubprocessorEntry{
			Name:   sp.Edges.Subprocessor.Name,
			Change: change,
		}, true
	})
}

// subprocessorNetChange classifies a vendor's changed join rows relative to the last notification floor:
// present at the floor but not anymore is a removal, newly present is an addition, present at both is an
// update, and a vendor both added and removed within the window nets to no change and is dropped
func subprocessorNetChange(rows []*ent.TrustCenterSubprocessor, floor time.Time) (string, bool) {
	wasPresent := lo.SomeBy(rows, func(sp *ent.TrustCenterSubprocessor) bool {
		return !sp.CreatedAt.After(floor)
	})
	isPresent := lo.SomeBy(rows, func(sp *ent.TrustCenterSubprocessor) bool {
		return sp.DeletedAt.IsZero()
	})

	switch {
	case wasPresent && !isPresent:
		return "Removed", true
	case !wasPresent && isPresent:
		return "Added", true
	case wasPresent && isPresent:
		return "Updated", true
	default:
		return "", false
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

const logoCheckTimeout = 5 * time.Second

// logoURLReachable reports whether a logo URL currently serves a successful response. Only the status is
// inspected (the body is discarded), so a dead URL can fall back to the default logo instead of rendering
// a broken image
func logoURLReachable(ctx context.Context, rawURL string) bool {
	resp, err := httpsling.SendWithContext(ctx,
		httpsling.Get(),
		httpsling.URL(rawURL),
		httpsling.Client(httpclient.Timeout(logoCheckTimeout)),
	)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return httpsling.IsSuccess(resp)
}
