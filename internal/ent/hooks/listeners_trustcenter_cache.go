package hooks

import (
	"context"
	"errors"
	"fmt"
	"net"
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
	"github.com/theopenlane/core/internal/ent/generated/dnsverification"
	notegen "github.com/theopenlane/core/internal/ent/generated/note"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcentercompliance"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterentity"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterfaq"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessor"
	"github.com/theopenlane/core/internal/trustcenterurl"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// RegisterGalaTrustCenterCacheListeners registers trust center cache listeners on Gala.
func RegisterGalaTrustCenterCacheListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g,
		eventqueue.MutationListener{
			Schema: entgen.TypeTrustCenterDoc,
			Name:   "trustcenter.cache.doc",
			Handle: handleTrustCenterDocMutationGala,
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeNote,
			Name:   "trustcenter.cache.note",
			Fields: []string{notegen.FieldTrustCenterID},
			Handle: handleNoteMutationGala,
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeTrustCenterEntity,
			Name:   "trustcenter.cache.entity",
			Handle: func(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
				return refreshResolvedTrustCenter(inv, payload, trustcenterentity.FieldTrustCenterID, "entity mutation",
					inv.Client.TrustCenterEntity.Get, func(e *entgen.TrustCenterEntity) string { return e.TrustCenterID })
			},
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeTrustCenterSubprocessor,
			Name:   "trustcenter.cache.trustcenter_subprocessor",
			Handle: func(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
				return refreshResolvedTrustCenter(inv, payload, trustcentersubprocessor.FieldTrustCenterID, "trust center subprocessor mutation",
					inv.Client.TrustCenterSubprocessor.Get, func(e *entgen.TrustCenterSubprocessor) string { return e.TrustCenterID })
			},
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeTrustCenterCompliance,
			Name:   "trustcenter.cache.compliance",
			Handle: func(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
				return refreshResolvedTrustCenter(inv, payload, trustcentercompliance.FieldTrustCenterID, "compliance mutation",
					inv.Client.TrustCenterCompliance.Get, func(e *entgen.TrustCenterCompliance) string { return e.TrustCenterID })
			},
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeTrustCenterFAQ,
			Name:   "trustcenter.cache.faq",
			Handle: func(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
				return refreshResolvedTrustCenter(inv, payload, trustcenterfaq.FieldTrustCenterID, "faq mutation",
					inv.Client.TrustCenterFAQ.Get, func(e *entgen.TrustCenterFAQ) string { return e.TrustCenterID })
			},
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeTrustCenterSetting,
			Name:   "trustcenter.cache.setting",
			Handle: func(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
				return refreshResolvedTrustCenter(inv, payload, trustcentersetting.FieldTrustCenterID, "setting mutation",
					inv.Client.TrustCenterSetting.Get, func(e *entgen.TrustCenterSetting) string { return e.TrustCenterID })
			},
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeSubprocessor,
			Name:   "trustcenter.cache.subprocessor",
			Handle: handleSubprocessorMutationGala,
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeStandard,
			Name:   "trustcenter.cache.standard",
			Handle: handleStandardMutationGala,
		},
		eventqueue.MutationListener{
			Schema: entgen.TypeTrustCenter,
			Name:   "trustcenter.cache.trust_center",
			Handle: handleTrustCenterMutationGala,
		},
	)
}

// trustCenterIDForMutation resolves the mutated row's trust center from the payload value
// when present, falling back to loading the row; a missing linkage or failed load skips
// the refresh
func trustCenterIDForMutation[T any](inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload, field string, load func(context.Context, string) (T, error), trustCenterID func(T) string) string {
	if id, ok := eventqueue.MutationStringValue(payload, field); ok {
		return id
	}

	entity, err := load(inv.Context, inv.EntityID)
	if err != nil {
		logx.FromContext(inv.Context).Warn().Err(err).Str("entity_id", inv.EntityID).Msg("failed to load entity for trust center cache invalidation")

		return ""
	}

	return trustCenterID(entity)
}

// refreshResolvedTrustCenter refreshes the cache for the trust center resolved from a mutated row
func refreshResolvedTrustCenter[T any](inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload, field, source string, load func(context.Context, string) (T, error), trustCenterID func(T) string) error {
	if id := trustCenterIDForMutation(inv, payload, field, load, trustCenterID); id != "" {
		refreshTrustCenterCache(inv.Context, inv.Client, id, source)
	}

	return nil
}

// handleTrustCenterDocMutationGala processes TrustCenterDoc mutations and invalidates cache when needed.
func handleTrustCenterDocMutationGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
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

	return refreshResolvedTrustCenter(inv, payload, trustcenterdoc.FieldTrustCenterID, "doc mutation",
		inv.Client.TrustCenterDoc.Get, func(doc *entgen.TrustCenterDoc) string { return doc.TrustCenterID })
}

// handleNoteMutationGala processes Note mutations and invalidates cache when trust center linkage changes.
func handleNoteMutationGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	tcIDs := eventqueue.MutationStringSliceValue(payload, notegen.FieldTrustCenterID)

	if len(tcIDs) == 0 {
		if id := trustCenterIDForMutation(inv, payload, notegen.FieldTrustCenterID,
			inv.Client.Note.Get, func(note *entgen.Note) string { return note.TrustCenterID }); id != "" {
			tcIDs = []string{id}
		}
	}

	for _, tcID := range tcIDs {
		refreshTrustCenterCache(inv.Context, inv.Client, tcID, "note mutation")
	}

	return nil
}

// handleSubprocessorMutationGala processes Subprocessor mutations and invalidates related trust center cache.
func handleSubprocessorMutationGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	if !cacheFieldsChanged(payload, subprocessor.FieldName, subprocessor.FieldLogoFileID, subprocessor.FieldLogoRemoteURL) {
		return nil
	}

	processors, err := inv.Client.TrustCenterSubprocessor.Query().
		Where(trustcentersubprocessor.SubprocessorID(inv.EntityID)).
		Select(trustcentersubprocessor.FieldTrustCenterID).
		All(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Warn().Err(err).Str("subprocessor_id", inv.EntityID).Msg("failed to query trust center subprocessors")

		return nil
	}

	trustCenterIDs := lo.Uniq(lo.FilterMap(processors, func(tcs *entgen.TrustCenterSubprocessor, _ int) (string, bool) {
		return tcs.TrustCenterID, tcs.TrustCenterID != ""
	}))

	for _, tcID := range trustCenterIDs {
		refreshTrustCenterCache(inv.Context, inv.Client, tcID, "subprocessor mutation")
	}

	return nil
}

// handleStandardMutationGala processes Standard mutations and invalidates related trust center cache.
func handleStandardMutationGala(inv eventqueue.Invocation, payload eventqueue.MutationGalaPayload) error {
	if !cacheFieldsChanged(payload, standard.FieldName, standard.FieldLogoFileID) {
		return nil
	}

	trustCenterDocs, err := inv.Client.TrustCenterDoc.Query().
		Where(trustcenterdoc.StandardID(inv.EntityID)).
		Select(trustcenterdoc.FieldTrustCenterID).
		All(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Warn().Err(err).Str("standard_id", inv.EntityID).Msg("failed to query trust center docs")

		return nil
	}

	trustCenterIDs := lo.Uniq(lo.FilterMap(trustCenterDocs, func(tcd *entgen.TrustCenterDoc, _ int) (string, bool) {
		return tcd.TrustCenterID, tcd.TrustCenterID != ""
	}))

	for _, tcID := range trustCenterIDs {
		refreshTrustCenterCache(inv.Context, inv.Client, tcID, "standard mutation")
	}

	return nil
}

// handleTrustCenterMutationGala processes TrustCenter mutations and refreshes cache.
func handleTrustCenterMutationGala(inv eventqueue.Invocation, _ eventqueue.MutationGalaPayload) error {
	refreshTrustCenterCache(inv.Context, inv.Client, inv.EntityID, "trust center mutation")

	return nil
}

// cacheFieldsChanged reports whether a mutation removes the row or changes any of the
// given cache-relevant fields
func cacheFieldsChanged(payload eventqueue.MutationGalaPayload, fields ...string) bool {
	switch strings.TrimSpace(payload.Operation) {
	case ent.OpCreate.String(), ent.OpUpdate.String(), ent.OpUpdateOne.String():
		return lo.SomeBy(fields, func(field string) bool {
			return eventqueue.MutationFieldChanged(payload, field)
		})
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

func refreshTrustCenterCache(ctx context.Context, client *entgen.Client, trustCenterID, source string) {
	if err := enqueueCacheRefresh(ctx, client, trustCenterID); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("trust_center_id", trustCenterID).
			Str("caller", source).Msg("failed to refresh trust center cache")
	}
}

// enqueueCacheRefresh triggers a cache refresh by hitting the trust center URL with ?fresh=1
func enqueueCacheRefresh(ctx context.Context, client *entgen.Client, trustCenterID string) error {
	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(trustCenterID)).
		Select(trustcenter.FieldCustomDomainID, trustcenter.FieldSlug, trustcenter.FieldPreviewDomainID).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to query trust center for cache invalidation")

		return err
	}

	var customDomain string
	if tc.CustomDomainID != nil {
		customDomain, err = getVerifiedDomain(ctx, client, *tc.CustomDomainID, false)
		if err != nil {
			return err
		}
	}

	targetURL := buildTrustCenterURL(customDomain, tc.Slug)
	if targetURL != "" {
		if err := triggerCacheRefresh(ctx, targetURL); err != nil {
			return err
		}
	}

	if tc.PreviewDomainID == "" {
		return nil
	}

	previewDomain, err := getVerifiedDomain(ctx, client, tc.PreviewDomainID, true)
	if err != nil {
		return err
	}
	if previewDomain == "" {
		return nil
	}

	previewURL := buildTrustCenterURL(previewDomain, "")
	if previewURL == "" {
		return nil
	}

	return triggerCacheRefresh(ctx, previewURL)
}

// buildTrustCenterURL constructs the trust center URL from custom domain or slug, delegating to the
// shared trustcenterurl package
func buildTrustCenterURL(customDomain, slug string) string {
	return trustcenterurl.BuildURL(customDomain, slug)
}

func getVerifiedDomain(ctx context.Context, client *entgen.Client, domainID string, isPreviewDomain bool) (string, error) {
	logField := trustcenter.FieldCustomDomainID
	if isPreviewDomain {
		logField = trustcenter.FieldPreviewDomainID
	}

	cd, err := client.CustomDomain.Query().
		Where(customdomain.ID(domainID)).
		Select(customdomain.FieldCnameRecord, customdomain.FieldDNSVerificationID).
		WithDNSVerification(func(q *entgen.DNSVerificationQuery) {
			q.Select(dnsverification.FieldDNSVerificationStatus)
		}).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str(logField, domainID).Msg("failed to query custom domain for cache invalidation")

		return "", err
	}

	dnsVerification, err := cd.Edges.DNSVerificationOrErr()
	if err != nil || dnsVerification == nil {
		logx.FromContext(ctx).Warn().Err(err).Str(logField, domainID).
			Msg("dns verification not found for custom domain, skipping custom domain cache refresh url")

		return "", nil
	}

	if dnsVerification.DNSVerificationStatus != enums.DNSVerificationStatusActive {
		logx.FromContext(ctx).Info().Str(logField, domainID).
			Str("dns_verification_status", dnsVerification.DNSVerificationStatus.String()).
			Msg("custom domain dns verification is not active, skipping custom domain cache refresh url")

		return "", nil
	}

	return cd.CnameRecord, nil
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

		if err != nil {
			var dnsErr *net.DNSError
			if errors.As(err, &dnsErr) {
				logx.FromContext(ctx).Info().Err(err).Str("target_url", targetURL).Msg("dns lookup failed for trust center cache refresh, skipping")
				return nil
			}
		}

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
