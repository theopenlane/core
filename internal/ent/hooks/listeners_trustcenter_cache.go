package hooks

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/entityops"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/dnsverification"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterdoc"
	"github.com/theopenlane/core/internal/ent/generated/trustcentersubprocessor"
	"github.com/theopenlane/core/internal/trustcenterurl"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// trustCenterSoftDeleteKeys lets cache listeners load soft-deleted rows
var trustCenterSoftDeleteKeys = []func(context.Context) context.Context{entx.SkipSoftDelete}

// trustCenterIDField is the trust center linkage field name shared by every
// trust-center-linked schema
const trustCenterIDField = "trust_center_id"

// TrustCenterCacheListeners refreshes the trust center cache when trust-center-linked
// records change, including soft deletes
func TrustCenterCacheListeners() []gala.Registration {
	return append(entityops.ForSchemas([]*entityops.Schema{
		entityops.SchemaTrustCenterEntity,
		entityops.SchemaTrustCenterSubprocessor,
		entityops.SchemaTrustCenterCompliance,
		entityops.SchemaTrustCenterFAQ,
		entityops.SchemaTrustCenterSetting,
	}, entityops.MutationListener{
		Operations:  entityops.AllMutationOps,
		ContextKeys: trustCenterSoftDeleteKeys,
		Handle:      refreshResolvedTrustCenter,
	}),
		entityops.MutationListener{
			Schema:      entityops.SchemaTrustCenterDoc,
			Operations:  entityops.AllMutationOps,
			ContextKeys: trustCenterSoftDeleteKeys,
			Handle:      handleTrustCenterDocMutationGala,
		},
		entityops.MutationListener{
			Schema:      entityops.SchemaNote,
			Operations:  entityops.AllMutationOps,
			Fields:      []string{trustCenterIDField},
			ContextKeys: trustCenterSoftDeleteKeys,
			Handle:      handleNoteMutationGala,
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaSubprocessor,
			Operations: entityops.AllMutationOps,
			Fields: []string{
				subprocessor.FieldName,
				subprocessor.FieldLogoFileID,
				subprocessor.FieldLogoRemoteURL,
			},
			ContextKeys: trustCenterSoftDeleteKeys,
			Handle:      handleSubprocessorMutationGala,
		},
		entityops.MutationListener{
			Schema:     entityops.SchemaStandard,
			Operations: entityops.AllMutationOps,
			Fields: []string{
				standard.FieldName,
				standard.FieldLogoFileID,
			},
			ContextKeys: trustCenterSoftDeleteKeys,
			Handle:      handleStandardMutationGala,
		},
		entityops.MutationListener{
			Schema:      entityops.SchemaTrustCenter,
			Operations:  entityops.AllMutationOps,
			ContextKeys: trustCenterSoftDeleteKeys,
			Handle:      handleTrustCenterMutationGala,
		},
	)
}

// trustCenterIDForMutation resolves the mutated row's trust center from the payload,
// falling back to loading the row; a missing linkage skips the refresh
func trustCenterIDForMutation(inv entityops.Invocation, payload entityops.MutationPayload) (string, error) {
	if id, ok := payload.StringValue(trustCenterIDField); ok {
		return id, nil
	}

	// hard deletes stash the removed row's values; the row itself is gone
	if id, ok := payload.OldStringValue(trustCenterIDField); ok && id != "" {
		return id, nil
	}

	row, err := inv.Schema.Load(inv.Context, inv.Client, inv.EntityID)
	if err != nil {
		if entgen.IsNotFound(err) {
			return "", nil
		}
		logx.FromContext(inv.Context).Warn().Err(err).Str("entity_id", inv.EntityID).Msg("failed to load entity for trust center cache invalidation")

		return "", err
	}

	fields, err := jsonx.Decode[map[string]any](row)
	if err != nil {
		logx.FromContext(inv.Context).Warn().Err(err).Str("entity_id", inv.EntityID).Msg("failed to decode entity for trust center cache invalidation")

		return "", err
	}

	id, _ := fields[trustCenterIDField].(string)

	return id, nil
}

// refreshResolvedTrustCenter refreshes the cache for the trust center resolved from a
// mutated row; it registers directly as the handler for every trust-center-linked schema
func refreshResolvedTrustCenter(inv entityops.Invocation, payload entityops.MutationPayload) error {
	if id, err := trustCenterIDForMutation(inv, payload); err != nil {
		return err
	} else if id != "" {
		refreshTrustCenterCache(inv.Context, inv.Client, id, inv.Schema.Snake+" mutation")
	}

	return nil
}

// refreshLinkedTrustCenters refreshes the cache for every distinct non-empty trust center id
func refreshLinkedTrustCenters(inv entityops.Invocation, ids []string, source string) {
	trustCenterIDs := lo.Uniq(lo.Compact(ids))

	for _, tcID := range trustCenterIDs {
		refreshTrustCenterCache(inv.Context, inv.Client, tcID, source)
	}
}

// handleTrustCenterDocMutationGala processes TrustCenterDoc mutations and invalidates cache when needed
func handleTrustCenterDocMutationGala(inv entityops.Invocation, payload entityops.MutationPayload) error {
	shouldClearCache := false

	switch strings.TrimSpace(payload.Operation) {
	case entityops.OpDelete, entityops.OpDeleteOne, entityops.OpSoftDelete:
		shouldClearCache = true
	case entityops.OpCreate:
		rawVisibility, _ := payload.Value(trustcenterdoc.FieldVisibility)

		visibility, ok := entityops.ParseEnum(
			rawVisibility,
			enums.ToTrustCenterDocumentVisibility,
			enums.TrustCenterDocumentVisibilityInvalid,
		)
		if ok {
			if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible ||
				visibility == enums.TrustCenterDocumentVisibilityProtected {
				shouldClearCache = true
			}
		}
	case entityops.OpUpdate, entityops.OpUpdateOne:
		if payload.FieldChanged(trustcenterdoc.FieldVisibility) {
			shouldClearCache = true
		}
	}

	if !shouldClearCache {
		return nil
	}

	return refreshResolvedTrustCenter(inv, payload)
}

// handleNoteMutationGala processes Note mutations and invalidates cache when trust center linkage changes
func handleNoteMutationGala(inv entityops.Invocation, payload entityops.MutationPayload) error {
	tcIDs := payload.StringSliceValue(trustCenterIDField)

	if len(tcIDs) == 0 {
		if id, err := trustCenterIDForMutation(inv, payload); err == nil && id != "" {
			tcIDs = []string{id}
		} else if err != nil {
			return err
		}
	}

	refreshLinkedTrustCenters(inv, tcIDs, "note mutation")

	return nil
}

// handleSubprocessorMutationGala processes Subprocessor mutations and invalidates related trust center cache
func handleSubprocessorMutationGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	processors, err := inv.Client.TrustCenterSubprocessor.Query().
		Where(trustcentersubprocessor.SubprocessorID(inv.EntityID)).
		Select(trustcentersubprocessor.FieldTrustCenterID).
		All(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Warn().Err(err).Str("subprocessor_id", inv.EntityID).Msg("failed to query trust center subprocessors")

		return err
	}

	refreshLinkedTrustCenters(inv, lo.Map(processors, func(tcs *entgen.TrustCenterSubprocessor, _ int) string {
		return tcs.TrustCenterID
	}), "subprocessor mutation")

	return nil
}

// handleStandardMutationGala processes Standard mutations and invalidates related trust center cache
func handleStandardMutationGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	trustCenterDocs, err := inv.Client.TrustCenterDoc.Query().
		Where(trustcenterdoc.StandardID(inv.EntityID)).
		Select(trustcenterdoc.FieldTrustCenterID).
		All(inv.Context)
	if err != nil {
		logx.FromContext(inv.Context).Warn().Err(err).Str("standard_id", inv.EntityID).Msg("failed to query trust center docs")

		return err
	}

	refreshLinkedTrustCenters(inv, lo.Map(trustCenterDocs, func(tcd *entgen.TrustCenterDoc, _ int) string {
		return tcd.TrustCenterID
	}), "standard mutation")

	return nil
}

// handleTrustCenterMutationGala processes TrustCenter mutations and refreshes cache
func handleTrustCenterMutationGala(inv entityops.Invocation, _ entityops.MutationPayload) error {
	refreshTrustCenterCache(inv.Context, inv.Client, inv.EntityID, "trust center mutation")

	return nil
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

// refreshTrustCenterCache is best-effort: refresh failures are logged and never fail the
// envelope, so a broken domain or origin cannot park mutation events in retry
func refreshTrustCenterCache(ctx context.Context, client *entgen.Client, trustCenterID, source string) {
	ctx = logx.WithFields(ctx, map[string]any{"trust_center_id": trustCenterID, "caller": source})

	if err := enqueueCacheRefresh(ctx, client, trustCenterID); err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to refresh trust center cache")
	}
}

// enqueueCacheRefresh triggers a cache refresh by hitting the trust center URL with ?fresh=1
func enqueueCacheRefresh(ctx context.Context, client *entgen.Client, trustCenterID string) error {
	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(trustCenterID)).
		Select(trustcenter.FieldCustomDomainID, trustcenter.FieldSlug, trustcenter.FieldPreviewDomainID).
		Only(ctx)
	if err != nil {
		if entgen.IsNotFound(err) {
			return nil
		}
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to query trust center for cache invalidation")

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

	ctx = logx.WithFields(ctx, map[string]any{logField: domainID})

	cd, err := client.CustomDomain.Query().
		Where(customdomain.ID(domainID)).
		Select(customdomain.FieldCnameRecord, customdomain.FieldDNSVerificationID).
		WithDNSVerification(func(q *entgen.DNSVerificationQuery) {
			q.Select(dnsverification.FieldDNSVerificationStatus)
		}).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to query custom domain for cache invalidation")

		return "", err
	}

	dnsVerification, err := cd.Edges.DNSVerificationOrErr()
	if err != nil || dnsVerification == nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("dns verification not found for custom domain, skipping custom domain cache refresh url")

		return "", nil
	}

	if dnsVerification.DNSVerificationStatus != enums.DNSVerificationStatusActive {
		logx.FromContext(ctx).Info().Str("dns_verification_status", dnsVerification.DNSVerificationStatus.String()).Msg("custom domain dns verification is not active, skipping custom domain cache refresh url")

		return "", nil
	}

	return cd.CnameRecord, nil
}

// triggerCacheRefresh makes an HTTP request to the trust center URL with the fresh query parameter
func triggerCacheRefresh(ctx context.Context, targetURL string) error {
	ctx = logx.WithFields(ctx, map[string]any{"target_url": targetURL})

	requester, err := httpsling.New(httpsling.Client(httpclient.Timeout(cacheRefreshTimeout)))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create HTTP client for cache refresh")
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
		resp, err := requester.ReceiveWithContext(ctx, nil, append(requestOpts, httpsling.Header("X-Cache-Refresh-Attempt", strconv.Itoa(attempt+1)))...)

		if err != nil {
			if _, ok := errors.AsType[*net.DNSError](err); ok {
				logx.FromContext(ctx).Info().Err(err).Msg("dns lookup failed for trust center cache refresh, skipping")

				return nil
			}
		}

		if err == nil && resp != nil {
			success := httpsling.IsSuccess(resp)
			statusCode := resp.StatusCode

			// close per attempt; a deferred close inside the retry loop leaks bodies
			// across attempts
			_ = resp.Body.Close()

			if success {
				logx.FromContext(ctx).Info().Int("status_code", statusCode).Msg("successfully triggered cache refresh")
				return nil
			}

			if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
				logx.FromContext(ctx).Warn().Int("status_code", statusCode).Msg("cache refresh request failed with client error, will not retry")

				return ErrCacheRefreshFailed
			}
		}

		if attempt == cacheRefreshMaxRetries-1 {
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to trigger cache refresh after maximum retries")

				return fmt.Errorf("%w: %w", ErrCacheRefreshFailed, err)
			}

			logx.FromContext(ctx).Error().Msg("failed to trigger cache refresh after maximum retries")

			return ErrCacheRefreshFailed
		}

		wait := policy.NextBackOff()
		if wait == backoff.Stop {
			wait = cacheRefreshInitialBackoff
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	return nil
}
