package hooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/cenkalti/backoff/v5"
	"github.com/samber/lo"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	notegen "github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentercompliance"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterentity"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessor"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaTrustCenterCacheListeners registers trust center cache listeners on Gala.
func RegisterGalaTrustCenterCacheListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeTrustCenterDoc),
			Name:   "trustcenter.cache.doc",
			Handle: handleTrustCenterDocMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeNote),
			Name:   "trustcenter.cache.note",
			Handle: handleNoteMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeTrustCenterEntity),
			Name:   "trustcenter.cache.entity",
			Handle: handleTrustCenterEntityMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeTrustCenterSubprocessor),
			Name:   "trustcenter.cache.trustcenter_subprocessor",
			Handle: handleTrustCenterSubprocessorMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeTrustCenterCompliance),
			Name:   "trustcenter.cache.compliance",
			Handle: handleTrustCenterComplianceMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeSubprocessor),
			Name:   "trustcenter.cache.subprocessor",
			Handle: handleSubprocessorMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeStandard),
			Name:   "trustcenter.cache.standard",
			Handle: handleStandardMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeTrustCenterSetting),
			Name:   "trustcenter.cache.setting",
			Handle: handleTrustCenterSettingMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:  eventqueue.MutationTopic(eventqueue.MutationConcernDirect, entgen.TypeTrustCenter),
			Name:   "trustcenter.cache.trust_center",
			Handle: handleTrustCenterMutationGala,
		},
	)
}

// handleTrustCenterDocMutationGala processes TrustCenterDoc mutations and invalidates cache when needed.
func handleTrustCenterDocMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	shouldClearCache := false

	switch strings.TrimSpace(payload.Operation) {
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), eventqueue.SoftDeleteOne:
		shouldClearCache = true
	case ent.OpCreate.String():
		visibility, ok := eventqueue.ParseEnum(
			payload.ProposedChanges[trustcenterdoc.FieldVisibility],
			enums.ToTrustCenterDocumentVisibility,
			enums.TrustCenterDocumentVisibilityInvalid,
		)
		if ok {
			if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible ||
				visibility == enums.TrustCenterDocumentVisibilityProtected {
				shouldClearCache = true
			}
		}
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		if eventqueue.MutationFieldChanged(payload, trustcenterdoc.FieldVisibility) {
			shouldClearCache = true
		}
	}

	if !shouldClearCache {
		return nil
	}

	trustCenterID, _ := eventqueue.MutationStringValue(payload, trustcenterdoc.FieldTrustCenterID)
	if trustCenterID == "" {
		docID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
		if ok && docID != "" {
			doc, err := client.TrustCenterDoc.Query().
				Where(trustcenterdoc.ID(docID)).
				Select(trustcenterdoc.FieldTrustCenterID).
				Only(ctx.Context)
			if err != nil || doc == nil {
				logx.FromContext(ctx.Context).Warn().Err(err).Str("doc_id", docID).Msg("failed to query trust center doc for cache invalidation")

				return nil
			}

			trustCenterID = doc.TrustCenterID
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context, client, trustCenterID)
}

// handleNoteMutationGala processes Note mutations and invalidates cache when trust center linkage changes.
func handleNoteMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	if !eventqueue.MutationFieldChanged(payload, notegen.FieldTrustCenterID) {
		return nil
	}

	tcIDs := eventqueue.MutationStringSliceValue(payload, notegen.FieldTrustCenterID)

	if len(tcIDs) == 0 {
		noteID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
		if ok && noteID != "" {
			note, err := client.Note.Query().
				Where(notegen.ID(noteID)).
				Select(notegen.FieldTrustCenterID).
				Only(ctx.Context)
			if err == nil && note != nil && note.TrustCenterID != "" {
				tcIDs = []string{note.TrustCenterID}
			}
		}
	}

	if len(tcIDs) == 0 {
		return nil
	}

	for _, tcID := range tcIDs {
		if err := enqueueCacheRefresh(ctx.Context, client, tcID); err != nil {
			logx.FromContext(ctx.Context).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to trigger cache invalidation for note")
		}
	}

	return nil
}

// handleTrustCenterEntityMutationGala processes TrustCenterEntity mutations and invalidates cache.
func handleTrustCenterEntityMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	trustCenterID, _ := eventqueue.MutationStringValue(payload, trustcenterentity.FieldTrustCenterID)
	if trustCenterID == "" {
		entityID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
		if ok && entityID != "" {
			entity, err := client.TrustCenterEntity.Get(ctx.Context, entityID)
			if err == nil && entity != nil {
				trustCenterID = entity.TrustCenterID
			}
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context, client, trustCenterID)
}

// handleTrustCenterSubprocessorMutationGala processes TrustCenterSubprocessor mutations and invalidates cache.
func handleTrustCenterSubprocessorMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	trustCenterID, _ := eventqueue.MutationStringValue(payload, trustcentersubprocessor.FieldTrustCenterID)
	if trustCenterID == "" {
		entityID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
		if ok && entityID != "" {
			entity, err := client.TrustCenterSubprocessor.Get(ctx.Context, entityID)
			if err != nil || entity == nil {
				logx.FromContext(ctx.Context).Warn().Err(err).Str("subprocessor_id", entityID).Msg("failed to query trust center subprocessor for cache invalidation")
				return nil
			}

			trustCenterID = entity.TrustCenterID
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context, client, trustCenterID)
}

// handleTrustCenterComplianceMutationGala processes TrustCenterCompliance mutations and invalidates cache.
func handleTrustCenterComplianceMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	trustCenterID, _ := eventqueue.MutationStringValue(payload, trustcentercompliance.FieldTrustCenterID)
	if trustCenterID == "" {
		entityID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
		if ok && entityID != "" {
			entity, err := client.TrustCenterCompliance.Get(ctx.Context, entityID)
			if err != nil || entity == nil {
				logx.FromContext(ctx.Context).Warn().Err(err).Str("compliance_id", entityID).Msg("failed to query trust center compliance for cache invalidation")

				return nil
			}

			trustCenterID = entity.TrustCenterID
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context, client, trustCenterID)
}

// handleSubprocessorMutationGala processes Subprocessor mutations and invalidates related trust center cache.
func handleSubprocessorMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	if !shouldInvalidateCacheForSubprocessor(payload) {
		return nil
	}

	subprocessorID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || subprocessorID == "" {
		return nil
	}

	processors, err := client.TrustCenterSubprocessor.Query().
		Where(trustcentersubprocessor.SubprocessorID(subprocessorID)).
		Select(trustcentersubprocessor.FieldTrustCenterID).
		All(ctx.Context)
	if err != nil {
		logx.FromContext(ctx.Context).Warn().Err(err).Str("subprocessor_id", subprocessorID).Msg("failed to query trust center subprocessors")

		return nil
	}

	if len(processors) == 0 {
		return nil
	}

	trustCenterIDs := lo.Uniq(lo.FilterMap(processors, func(tcs *entgen.TrustCenterSubprocessor, _ int) (string, bool) {
		return tcs.TrustCenterID, tcs.TrustCenterID != ""
	}))

	for _, tcID := range trustCenterIDs {
		if err := enqueueCacheRefresh(ctx.Context, client, tcID); err != nil {
			logx.FromContext(ctx.Context).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to trigger cache invalidation for subprocessor")
		}
	}

	return nil
}

// handleStandardMutationGala processes Standard mutations and invalidates related trust center cache.
func handleStandardMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	if !shouldInvalidateCacheForStandard(payload) {
		return nil
	}

	standardID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || standardID == "" {
		return nil
	}

	trustCenterDocs, err := client.TrustCenterDoc.Query().
		Where(trustcenterdoc.StandardID(standardID)).
		Select(trustcenterdoc.FieldTrustCenterID).
		All(ctx.Context)
	if err != nil {
		logx.FromContext(ctx.Context).Warn().Err(err).Str("standard_id", standardID).Msg("failed to query trust center docs")

		return nil
	}

	if len(trustCenterDocs) == 0 {
		return nil
	}

	trustCenterIDs := lo.Uniq(lo.FilterMap(trustCenterDocs, func(tcd *entgen.TrustCenterDoc, _ int) (string, bool) {
		return tcd.TrustCenterID, tcd.TrustCenterID != ""
	}))

	for _, tcID := range trustCenterIDs {
		if err := enqueueCacheRefresh(ctx.Context, client, tcID); err != nil {
			logx.FromContext(ctx.Context).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to trigger cache invalidation for standard")
		}
	}

	return nil
}

// handleTrustCenterSettingMutationGala processes TrustCenterSetting mutations and refreshes cache.
func handleTrustCenterSettingMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	trustCenterID, _ := eventqueue.MutationStringValue(payload, trustcentersetting.FieldTrustCenterID)
	if trustCenterID == "" {
		settingID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
		if ok && settingID != "" {
			setting, err := client.TrustCenterSetting.Get(ctx.Context, settingID)
			if err != nil || setting == nil {
				logx.FromContext(ctx.Context).Warn().Err(err).Str("setting_id", settingID).Msg("failed to query trust center setting for cache invalidation")
				return nil
			}

			trustCenterID = setting.TrustCenterID
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context, client, trustCenterID)
}

// handleTrustCenterMutationGala processes TrustCenter mutations and refreshes cache.
func handleTrustCenterMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	switch strings.TrimSpace(payload.Operation) {
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), eventqueue.SoftDeleteOne:
		return nil
	}

	trustCenterID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context, client, trustCenterID)
}

// shouldInvalidateCacheForSubprocessor determines if subprocessor changes require cache invalidation.
func shouldInvalidateCacheForSubprocessor(payload eventqueue.MutationGalaPayload) bool {
	switch strings.TrimSpace(payload.Operation) {
	case ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return eventqueue.MutationFieldChanged(payload, subprocessor.FieldName) ||
			eventqueue.MutationFieldChanged(payload, subprocessor.FieldLogoFileID) ||
			eventqueue.MutationFieldChanged(payload, subprocessor.FieldLogoRemoteURL)
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), eventqueue.SoftDeleteOne:
		return true
	}

	return false
}

// shouldInvalidateCacheForStandard determines if standard changes require cache invalidation.
func shouldInvalidateCacheForStandard(payload eventqueue.MutationGalaPayload) bool {
	switch strings.TrimSpace(payload.Operation) {
	case ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return eventqueue.MutationFieldChanged(payload, standard.FieldName) ||
			eventqueue.MutationFieldChanged(payload, standard.FieldLogoFileID)
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), eventqueue.SoftDeleteOne:
		return true
	}

	return false
}

const (
	cacheRefreshTimeout        = 10 * time.Second
	cacheRefreshUserAgent      = "Openlane-CacheRefresh/1.0"
	cacheRefreshParam          = "fresh"
	cacheRefreshValue          = "1"
	cacheRefreshMaxRetries     = 3
	cacheRefreshInitialBackoff = 3 * time.Second
	cacheRefreshMaxBackoff     = 30 * time.Second
)

// enqueueCacheRefresh triggers a cache refresh by hitting the trust center URL with ?fresh=1
func enqueueCacheRefresh(ctx context.Context, client *entgen.Client, trustCenterID string) error {
	// In durable dispatch the context is reconstructed from a snapshot that does not include the
	// ent client, so interceptors relying on generated.FromContext (e.g. the FGA authz client used
	// by the modules feature-check interceptor) would get a nil client and return empty features,
	// causing the trust_center_module check to fail even when the module is enabled.
	ctx = entgen.NewContext(ctx, client)

	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(trustCenterID)).
		Select(trustcenter.FieldCustomDomainID, trustcenter.FieldSlug).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to query trust center for cache invalidation")

		return err
	}

	var customDomain string
	if tc.CustomDomainID != nil {
		cd, err := client.CustomDomain.Query().
			Where(customdomain.ID(*tc.CustomDomainID)).
			Select(customdomain.FieldCnameRecord).
			Only(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Str("custom_domain_id", *tc.CustomDomainID).Msg("failed to query custom domain for cache invalidation")

			return err
		}

		customDomain = cd.CnameRecord
	}

	targetURL := buildTrustCenterURL(customDomain, tc.Slug)
	if targetURL == "" {
		return nil
	}

	return triggerCacheRefresh(ctx, targetURL)
}

// buildTrustCenterURL constructs the trust center URL from custom domain or slug
func buildTrustCenterURL(customDomain, slug string) string {
	scheme := trustCenterConfig.CacheRefreshScheme
	if scheme == "" {
		scheme = "https"
	}

	// In test mode (http scheme), use DefaultTrustCenterDomain for all requests
	if scheme == "http" && trustCenterConfig.DefaultTrustCenterDomain != "" {
		return fmt.Sprintf("%s://%s", scheme, trustCenterConfig.DefaultTrustCenterDomain)
	}

	if customDomain != "" {
		return fmt.Sprintf("%s://%s", scheme, customDomain)
	}

	if slug != "" && trustCenterConfig.DefaultTrustCenterDomain != "" {
		return fmt.Sprintf("%s://%s/%s", scheme, trustCenterConfig.DefaultTrustCenterDomain, slug)
	}

	return ""
}

// triggerCacheRefresh makes an HTTP request to the trust center URL with the fresh query parameter
func triggerCacheRefresh(ctx context.Context, targetURL string) error {
	requester, err := httpsling.New(httpsling.Client(httpclient.Timeout(cacheRefreshTimeout)))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("target_url", targetURL).Msg("failed to create HTTP client for cache refresh")
		return err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = cacheRefreshInitialBackoff
	policy.MaxInterval = cacheRefreshMaxBackoff

	requestOpts := []httpsling.Option{
		httpsling.Get(targetURL),
		httpsling.QueryParam(cacheRefreshParam, cacheRefreshValue),
		httpsling.Header(httpsling.HeaderUserAgent, cacheRefreshUserAgent),
	}

	for attempt := range cacheRefreshMaxRetries {
		resp, err := requester.ReceiveWithContext(ctx, nil, append(requestOpts, httpsling.Header("X-Cache-Refresh-Attempt", fmt.Sprintf("%d", attempt+1)))...)

		if err == nil && resp != nil {
			defer resp.Body.Close()

			if httpsling.IsSuccess(resp) {
				logx.FromContext(ctx).Info().Str("target_url", targetURL).Int("status_code", resp.StatusCode).Msg("successfully triggered cache refresh")
				return nil
			}

			if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
				logx.FromContext(ctx).Warn().Str("target_url", targetURL).Int("status_code", resp.StatusCode).Msg("cache refresh request failed with client error, will not retry")
				return ErrCacheRefreshFailed
			}
		}

		if attempt == cacheRefreshMaxRetries-1 {
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("target_url", targetURL).Msg("failed to trigger cache refresh after maximum retries")
				return fmt.Errorf("%w: %w", ErrCacheRefreshFailed, err)
			}

			logx.FromContext(ctx).Error().Str("target_url", targetURL).Msg("failed to trigger cache refresh after maximum retries")
			return ErrCacheRefreshFailed
		}

		wait := policy.NextBackOff()
		if wait == backoff.Stop {
			wait = cacheRefreshInitialBackoff
		}

		time.Sleep(wait)
	}

	return nil
}
