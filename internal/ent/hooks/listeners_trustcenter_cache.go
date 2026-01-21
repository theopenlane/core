package hooks

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"entgo.io/ent"
	"github.com/cenkalti/backoff/v5"
	"github.com/samber/lo"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessor"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/logx"
)

// handleTrustCenterDocMutation processes TrustCenterDoc mutations and invalidates cache when necessary
func handleTrustCenterDocMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.TrustCenterDocMutation)
	if !ok {
		return nil
	}

	shouldClearCache := false

	switch payload.Operation {
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne:
		shouldClearCache = true
	case ent.OpCreate.String():
		if visibility, ok := mut.Visibility(); ok {
			if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible ||
				visibility == enums.TrustCenterDocumentVisibilityProtected {
				shouldClearCache = true
			}
		}
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		if _, ok := mut.Visibility(); ok {
			// any visibility change should clear cache to ensure consistency
			shouldClearCache = true
		}
	}

	if !shouldClearCache {
		return nil
	}

	var trustCenterID string
	if tcID, exists := mut.TrustCenterID(); exists {
		trustCenterID = tcID
	}

	if trustCenterID == "" {
		docID := payload.EntityID
		if docID == "" {
			if id, ok := mut.ID(); ok {
				docID = id
			}
		}

		if docID != "" {
			doc, err := payload.Client.TrustCenterDoc.Query().Where(trustcenterdoc.ID(docID)).Select(trustcenterdoc.FieldTrustCenterID).Only(ctx.Context())
			if err == nil && doc != nil {
				trustCenterID = doc.TrustCenterID
			}
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context(), payload.Client, trustCenterID)
}

// handleNoteMutation processes Note mutations and invalidates cache
func handleNoteMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.NoteMutation)
	if !ok {
		return nil
	}

	tcIDs := mut.TrustCenterIDs()
	if len(tcIDs) == 0 {
		return nil
	}

	for _, tcID := range tcIDs {
		if err := enqueueCacheRefresh(ctx.Context(), payload.Client, tcID); err != nil {
			logx.FromContext(ctx.Context()).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to trigger cache invalidation for note")
		}
	}

	return nil
}

// handleTrustCenterEntityMutation processes TrustCenterEntity mutations and invalidates cache
func handleTrustCenterEntityMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.TrustCenterEntityMutation)
	if !ok {
		return nil
	}

	var trustCenterID string
	if tcID, exists := mut.TrustCenterID(); exists {
		trustCenterID = tcID
	}

	if trustCenterID == "" {
		entityID := payload.EntityID
		if entityID == "" {
			if id, ok := mut.ID(); ok {
				entityID = id
			}
		}

		if entityID != "" {
			entity, err := payload.Client.TrustCenterEntity.Get(ctx.Context(), entityID)
			if err == nil && entity != nil {
				trustCenterID = entity.TrustCenterID
			}
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context(), payload.Client, trustCenterID)
}

// handleTrustCenterSubprocessorMutation processes TrustCenterSubprocessor mutations and invalidates cache
func handleTrustCenterSubprocessorMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.TrustCenterSubprocessorMutation)
	if !ok {
		return nil
	}

	var trustCenterID string
	if tcID, exists := mut.TrustCenterID(); exists {
		trustCenterID = tcID
	}

	if trustCenterID == "" {
		entityID := payload.EntityID
		if entityID == "" {
			if id, ok := mut.ID(); ok {
				entityID = id
			}
		}

		if entityID != "" {
			entity, err := payload.Client.TrustCenterSubprocessor.Get(ctx.Context(), entityID)
			if err == nil && entity != nil {
				trustCenterID = entity.TrustCenterID
			}
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context(), payload.Client, trustCenterID)
}

// handleTrustCenterComplianceMutation processes TrustCenterCompliance mutations and invalidates cache
func handleTrustCenterComplianceMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.TrustCenterComplianceMutation)
	if !ok {
		return nil
	}

	var trustCenterID string
	if tcID, exists := mut.TrustCenterID(); exists {
		trustCenterID = tcID
	}

	if trustCenterID == "" {
		entityID := payload.EntityID
		if entityID == "" {
			if id, ok := mut.ID(); ok {
				entityID = id
			}
		}

		if entityID != "" {
			entity, err := payload.Client.TrustCenterCompliance.Get(ctx.Context(), entityID)
			if err == nil && entity != nil {
				trustCenterID = entity.TrustCenterID
			}
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context(), payload.Client, trustCenterID)
}

// handleSubprocessorMutation processes Subprocessor mutations and invalidates cache for related trust centers
func handleSubprocessorMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.SubprocessorMutation)
	if !ok {
		return nil
	}

	if !shouldInvalidateCacheForSubprocessor(mut, payload.Operation) {
		return nil
	}

	subprocessorID := payload.EntityID
	if subprocessorID == "" {
		if id, ok := mut.ID(); ok {
			subprocessorID = id
		}
	}

	if subprocessorID == "" {
		return nil
	}

	processors, err := payload.Client.TrustCenterSubprocessor.Query().
		Where(trustcentersubprocessor.SubprocessorID(subprocessorID)).
		Select(trustcentersubprocessor.FieldTrustCenterID).
		All(ctx.Context())
	if err != nil {
		logx.FromContext(ctx.Context()).Warn().Err(err).Str("subprocessor_id", subprocessorID).Msg("failed to query trust center subprocessors")
		return nil
	}

	if len(processors) == 0 {
		return nil
	}

	trustCenterIDs := lo.Uniq(lo.FilterMap(processors, func(tcs *entgen.TrustCenterSubprocessor, _ int) (string, bool) {
		return tcs.TrustCenterID, tcs.TrustCenterID != ""
	}))

	for _, tcID := range trustCenterIDs {
		if err := enqueueCacheRefresh(ctx.Context(), payload.Client, tcID); err != nil {
			logx.FromContext(ctx.Context()).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to trigger cache invalidation for subprocessor")
		}
	}

	return nil
}

// handleStandardMutation processes Standard mutations and invalidates cache for related trust centers
func handleStandardMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.StandardMutation)
	if !ok {
		return nil
	}

	if !shouldInvalidateCacheForStandard(mut, payload.Operation) {
		return nil
	}

	standardID := payload.EntityID
	if standardID == "" {
		if id, ok := mut.ID(); ok {
			standardID = id
		}
	}

	if standardID == "" {
		return nil
	}

	trustCenterDocs, err := payload.Client.TrustCenterDoc.Query().
		Where(trustcenterdoc.StandardID(standardID)).
		Select(trustcenterdoc.FieldTrustCenterID).
		All(ctx.Context())
	if err != nil {
		logx.FromContext(ctx.Context()).Warn().Err(err).Str("standard_id", standardID).Msg("failed to query trust center docs")
		return nil
	}

	if len(trustCenterDocs) == 0 {
		return nil
	}

	trustCenterIDs := lo.Uniq(lo.FilterMap(trustCenterDocs, func(tcd *entgen.TrustCenterDoc, _ int) (string, bool) {
		return tcd.TrustCenterID, tcd.TrustCenterID != ""
	}))

	for _, tcID := range trustCenterIDs {
		if err := enqueueCacheRefresh(ctx.Context(), payload.Client, tcID); err != nil {
			logx.FromContext(ctx.Context()).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to trigger cache invalidation for standard")
		}
	}

	return nil
}

// handleTrustCenterSettingMutation processes TrustCenterSetting mutations and refreshes cache
func handleTrustCenterSettingMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.TrustCenterSettingMutation)
	if !ok {
		return nil
	}

	trustCenterID, exists := mut.TrustCenterID()
	if trustCenterID == "" || !exists {
		settingID := payload.EntityID
		if settingID == "" {
			if id, ok := mut.ID(); ok {
				settingID = id
			}
		}

		if settingID != "" {
			setting, err := payload.Client.TrustCenterSetting.Get(ctx.Context(), settingID)
			if err == nil && setting != nil {
				trustCenterID = setting.TrustCenterID
			}
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context(), payload.Client, trustCenterID)
}

// handleTrustCenterMutation processes TrustCenter mutations and refreshes cache
func handleTrustCenterMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil {
		return nil
	}

	switch payload.Operation {
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne:
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.TrustCenterMutation)
	if !ok {
		return nil
	}

	trustCenterID := payload.EntityID
	if trustCenterID == "" {
		if id, ok := mut.ID(); ok {
			trustCenterID = id
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheRefresh(ctx.Context(), payload.Client, trustCenterID)
}

// shouldInvalidateCacheForSubprocessor determines if subprocessor changes warrant cache invalidation
func shouldInvalidateCacheForSubprocessor(mut *entgen.SubprocessorMutation, operation string) bool {
	switch operation {
	case ent.OpCreate.String():
		_, hasName := mut.Name()
		_, hasLogoFileID := mut.LogoFileID()
		_, hasLogoRemoteURL := mut.LogoRemoteURL()
		return hasName || hasLogoFileID || hasLogoRemoteURL
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne:
		return true
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		_, hasName := mut.Name()
		_, hasLogoFileID := mut.LogoFileID()
		_, hasLogoRemoteURL := mut.LogoRemoteURL()
		logoFileCleared := mut.LogoFileIDCleared()
		logoRemoteURLCleared := mut.LogoRemoteURLCleared()
		return hasName || hasLogoFileID || hasLogoRemoteURL || logoFileCleared || logoRemoteURLCleared
	}

	return false
}

// shouldInvalidateCacheForStandard determines if standard changes warrant cache invalidation
func shouldInvalidateCacheForStandard(mut *entgen.StandardMutation, operation string) bool {
	switch operation {
	case ent.OpCreate.String():
		_, hasName := mut.Name()
		_, hasLogoFileID := mut.LogoFileID()
		return hasName || hasLogoFileID
	case ent.OpDelete.String(), ent.OpDeleteOne.String(), SoftDeleteOne:
		return true
	case ent.OpUpdate.String(), ent.OpUpdateOne.String():
		_, hasName := mut.Name()
		_, hasLogoFileID := mut.LogoFileID()
		logoFileCleared := mut.LogoFileIDCleared()
		return hasName || hasLogoFileID || logoFileCleared
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
				return nil
			}

			if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
				return ErrCacheRefreshFailed
			}
		}

		if attempt == cacheRefreshMaxRetries-1 {
			if err != nil {
				return fmt.Errorf("%w: %w", ErrCacheRefreshFailed, err)
			}

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
