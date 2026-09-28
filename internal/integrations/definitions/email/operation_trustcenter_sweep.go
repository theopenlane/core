package email

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
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/note"
	"github.com/theopenlane/core/v2/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersubprocessor"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/trustcenterurl"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/shortlinks"
	"github.com/theopenlane/core/v2/pkg/urlx"
)

// trustCenterNotificationGrace is the debounce window a post or subprocessor change must be stable for before subscribers are notified, giving authors time to make further edits
const trustCenterNotificationGrace = time.Hour

// TrustCenterNotificationSweep configures one trust center notification sweep cycle
type TrustCenterNotificationSweep struct{}

// TrustCenterNotificationOp is the operation ref for the global trust center notification sweep, which runs without a client
var TrustCenterNotificationOp = types.OperationRefOf[TrustCenterNotificationSweep]().HandlesRequest(runTrustCenterNotificationSweep) //nolint:revive

// Handle adapts the trust center notification sweep to the generic operation registration boundary
func (t TrustCenterNotificationSweep) Handle() types.OperationHandler {
	return func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
		processed, err := t.Run(ctx, req)
		if err != nil {
			return nil, err
		}

		return providerkit.EncodeResult(types.ScheduledCycleResult{Processed: processed}, ErrResultEncode)
	}
}

// Run executes one trust center notification sweep, returning the number of notifications dispatched; per-item failures log and continue while the joined error feeds the backoff
func (TrustCenterNotificationSweep) Run(ctx context.Context, req types.OperationRequest) (int, error) {
	now := time.Now()
	cutoff := now.Add(-trustCenterNotificationGrace)
	// the sweep scans trust center settings and subprocessors across every organization, and the
	// per-trust-center sends it dispatches must load branding from the trust center setting. The
	// cross-org bypass rides on the caller
	systemCtx := auth.WithSystemSweepContext(ctx)

	posts, postsErr := dispatchDuePosts(systemCtx, req, cutoff, now)
	subprocessors, subprocessorsErr := dispatchDueSubprocessorChanges(systemCtx, req, cutoff)

	return posts + subprocessors, errors.Join(postsErr, subprocessorsErr)
}

// dispatchDuePosts notifies subscribers about published posts flagged for notification that have been stable for the grace window
func dispatchDuePosts(ctx context.Context, req types.OperationRequest, cutoff, now time.Time) (int, error) {
	posts, err := req.DB.Note.Query().
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
		tc, customDomain, err := loadTrustCenter(ctx, req, post.TrustCenterID)
		if err != nil {
			errs = append(errs, fmt.Errorf("loading trust center for post %s: %w", post.ID, err))

			continue
		}

		title := lo.FromPtr(post.Title)

		content, err := TrustCenterUpdateContent(TrustCenterUpdateRequest{
			PostTitle:      title,
			PostText:       post.Text,
			TrustCenterURL: trackingLink(ctx, req, trustcenterurl.BuildURL(customDomain, tc.Slug), shortlinks.Metadata{Purpose: shortlinks.PurposeTrustCenterUpdate, OrganizationID: tc.OwnerID, TrustCenterID: tc.ID}, post.ID),
			UnsubscribeURL: trustcenterurl.UnsubscribeURL(customDomain, tc.Slug),
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("building notification content for post %s: %w", post.ID, err))

			continue
		}

		content["postID"] = post.ID

		name := lo.CoalesceOrEmpty(title, DefaultTrustCenterUpdateTitle)

		if err := createAndDispatchTrustCenterCampaign(ctx, req, tc.OwnerID, post.TrustCenterID, name, content); err != nil {
			errs = append(errs, fmt.Errorf("dispatching notification for post %s: %w", post.ID, err))

			continue
		}

		if err := req.DB.Note.UpdateOneID(post.ID).SetNotifiedAt(now).Exec(ctx); err != nil {
			errs = append(errs, fmt.Errorf("marking post %s notified: %w", post.ID, err))
		}

		dispatched++
	}

	return dispatched, errors.Join(errs...)
}

// dispatchDueSubprocessorChanges notifies subscribers about subprocessor changes for trust centers that opted in, coalescing all changes since the last notification into one send per trust center
func dispatchDueSubprocessorChanges(ctx context.Context, req types.OperationRequest, cutoff time.Time) (int, error) {
	settings, err := req.DB.TrustCenterSetting.Query().
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
		if setting.SubprocessorsNotifiedAt == nil {
			if err := req.DB.TrustCenterSetting.UpdateOneID(setting.ID).SetSubprocessorsNotifiedAt(cutoff).Exec(ctx); err != nil {
				errs = append(errs, fmt.Errorf("initializing subprocessor notified baseline for trust center %s: %w", setting.TrustCenterID, err))
			}

			continue
		}

		floor := *setting.SubprocessorsNotifiedAt

		changed, err := req.DB.TrustCenterSubprocessor.Query().
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

		tc, customDomain, err := loadTrustCenter(ctx, req, setting.TrustCenterID)
		if err != nil {
			errs = append(errs, fmt.Errorf("loading trust center %s for subprocessor notification: %w", setting.TrustCenterID, err))

			continue
		}

		entries := subprocessorEntries(changed, floor)

		if len(entries) == 0 {
			if err := req.DB.TrustCenterSetting.UpdateOneID(setting.ID).SetSubprocessorsNotifiedAt(latest).Exec(ctx); err != nil {
				errs = append(errs, fmt.Errorf("advancing subprocessor notified baseline for trust center %s: %w", setting.TrustCenterID, err))
			}

			continue
		}

		subscribers, err := activeTrustCenterSubscribers(ctx, req, setting.TrustCenterID)
		if err != nil {
			errs = append(errs, fmt.Errorf("loading subscribers for trust center %s subprocessor notification: %w", setting.TrustCenterID, err))

			continue
		}

		// the request carries only data; the subprocessor operation composes the subject and body copy
		// from the branding's company name with defined fallbacks
		// the tracking link is shared by every subscriber of this notification, so it is created once
		// per baseline window and only when there is someone to send it to
		trustCenterURL := trustcenterurl.BuildURL(customDomain, tc.Slug)
		if len(subscribers) > 0 {
			trustCenterURL = trackingLink(ctx, req, trustCenterURL, shortlinks.Metadata{Purpose: shortlinks.PurposeSubprocessorNotification, OrganizationID: tc.OwnerID, TrustCenterID: tc.ID}, latest.UTC().Format(time.RFC3339))
		}

		base := SubprocessorNotificationRequest{
			TrustCenterBranding: TrustCenterBrandingFromSetting(setting),
			Subprocessors:       entries,
			TrustCenterURL:      trustCenterURL,
		}

		if base.LogoURL != "" && !logoURLReachable(ctx, base.LogoURL) {
			logx.FromContext(ctx).Warn().Str("trust_center_id", setting.TrustCenterID).Str("logo_url", base.LogoURL).Msg("trust center logo URL unreachable, using default logo")

			base.LogoURL = ""
		}

		failedSends := 0

		for _, sub := range subscribers {
			sendReq := base
			sendReq.RecipientInfo = RecipientInfo{
				Email:            sub.Email,
				UnsubscribeToken: sub.Token,
			}
			sendReq.UnsubscribeURL = trustcenterurl.UnsubscribeURLWithToken(customDomain, tc.Slug, sub.Token)

			if err := sendSubprocessorNotification(ctx, req, sendReq); err != nil {
				errs = append(errs, fmt.Errorf("sending subprocessor notification to subscriber %s for trust center %s: %w", sub.ID, setting.TrustCenterID, err))

				failedSends++
			}
		}

		if len(subscribers) > 0 && failedSends == len(subscribers) {
			continue
		}

		if err := req.DB.TrustCenterSetting.UpdateOneID(setting.ID).SetSubprocessorsNotifiedAt(latest).Exec(ctx); err != nil {
			errs = append(errs, fmt.Errorf("advancing subprocessor notified baseline for trust center %s: %w", setting.TrustCenterID, err))
		}

		dispatched++
	}

	return dispatched, errors.Join(errs...)
}

// loadTrustCenter loads a trust center with its custom domain, returning the resolved custom domain
func loadTrustCenter(ctx context.Context, req types.OperationRequest, trustCenterID string) (*ent.TrustCenter, string, error) {
	tc, err := req.DB.TrustCenter.Query().
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

// createAndDispatchTrustCenterCampaign creates a trust center update campaign carrying the supplied content and dispatches it through the campaign send, which materializes subscriber targets
func createAndDispatchTrustCenterCampaign(ctx context.Context, req types.OperationRequest, ownerID, trustCenterID, name string, content map[string]any) error {
	camp, err := req.DB.Campaign.Create().
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

	config, err := json.Marshal(SendBrandedCampaignRequest{
		CampaignDispatchInput: CampaignDispatchInput{CampaignID: camp.ID},
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("campaign_id", camp.ID).Msg("failed marshaling campaign dispatch config")

		return err
	}

	integrationID, err := operations.ResolveOwnerIntegration(ctx, req.DB, DefinitionID.ID(), ownerID, func(inst *ent.Integration) bool {
		return inst.CampaignEmail
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", ownerID).Msg("failed resolving integration for trust center notification")

		return err
	}

	_, err = req.Dispatch(ctx, types.DispatchRequest{
		IntegrationID: integrationID,
		DefinitionID:  DefinitionID.ID(),
		OwnerID:       ownerID,
		Operation:     SendCampaignOp.Name(),
		Config:        config,
		RunType:       enums.IntegrationRunTypeScheduled,
		Runtime:       integrationID == "",
	})

	return err
}

// subprocessorEntries coalesces the changed trust center subprocessor join rows per subprocessor and maps each vendor's net change relative to the last notification floor into a structured change entry
func subprocessorEntries(changed []*ent.TrustCenterSubprocessor, floor time.Time) []SubprocessorEntry {
	withVendor := lo.Filter(changed, func(sp *ent.TrustCenterSubprocessor, _ int) bool {
		return sp.Edges.Subprocessor != nil
	})

	groups := lo.GroupBy(withVendor, func(sp *ent.TrustCenterSubprocessor) string {
		return sp.SubprocessorID
	})

	return lo.FilterMap(lo.UniqBy(withVendor, func(sp *ent.TrustCenterSubprocessor) string {
		return sp.SubprocessorID
	}), func(sp *ent.TrustCenterSubprocessor, _ int) (SubprocessorEntry, bool) {
		change, ok := subprocessorNetChange(groups[sp.SubprocessorID], floor)
		if !ok {
			return SubprocessorEntry{}, false
		}

		return SubprocessorEntry{
			Name:   sp.Edges.Subprocessor.Name,
			Change: change,
		}, true
	})
}

// subprocessorNetChange classifies a vendor's changed join rows relative to the last notification floor
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

// activeTrustCenterSubscribers returns the trust center's active, verified, subscribed recipients
func activeTrustCenterSubscribers(ctx context.Context, req types.OperationRequest, trustCenterID string) ([]*ent.Subscriber, error) {
	return req.DB.Subscriber.Query().
		Where(
			subscriber.TrustCenterID(trustCenterID),
			subscriber.Active(true),
			subscriber.VerifiedEmail(true),
			subscriber.Unsubscribed(false),
		).
		All(ctx)
}

// sendSubprocessorNotification dispatches the subprocessor notification to a single recipient as a runtime system email
func sendSubprocessorNotification(ctx context.Context, req types.OperationRequest, sendReq SubprocessorNotificationRequest) error {
	config, err := json.Marshal(sendReq)
	if err != nil {
		return err
	}

	_, err = req.Dispatch(ctx, types.DispatchRequest{
		DefinitionID: DefinitionID.ID(),
		Operation:    SubprocessorNotificationOp.Name(),
		Config:       config,
		RunType:      enums.IntegrationRunTypeScheduled,
		Runtime:      true,
	})

	return err
}

const logoCheckTimeout = 5 * time.Second

// logoURLReachable reports whether a logo URL is an absolute http(s) URL that currently serves a successful response
func logoURLReachable(ctx context.Context, rawURL string) bool {
	logoURL, err := urlx.ParseAbsolute(rawURL)
	if err != nil {
		return false
	}

	requester, err := urlx.NewRequester(httpsling.Client(httpclient.Timeout(logoCheckTimeout)))
	if err != nil {
		return false
	}

	resp, err := requester.SendWithContext(ctx,
		httpsling.Get(),
		httpsling.URL(logoURL.String()),
	)
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	return httpsling.IsSuccess(resp)
}
