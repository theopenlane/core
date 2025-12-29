package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/samber/lo"
	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
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

	return enqueueCacheInvalidation(ctx.Context(), payload.Client, mut.Job, trustCenterID)
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
		if err := enqueueCacheInvalidation(ctx.Context(), payload.Client, mut.Job, tcID); err != nil {
			logx.FromContext(ctx.Context()).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to enqueue cache invalidation for note")
		}
	}

	return nil
}

// handleTrustcenterEntityMutation processes TrustcenterEntity mutations and invalidates cache
func handleTrustcenterEntityMutation(ctx *soiree.EventContext, payload *events.MutationPayload) error {
	if payload == nil || payload.Client == nil {
		return nil
	}

	mut, ok := payload.Mutation.(*entgen.TrustcenterEntityMutation)
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
			entity, err := payload.Client.TrustcenterEntity.Get(ctx.Context(), entityID)
			if err == nil && entity != nil {
				trustCenterID = entity.TrustCenterID
			}
		}
	}

	if trustCenterID == "" {
		return nil
	}

	return enqueueCacheInvalidation(ctx.Context(), payload.Client, mut.Job, trustCenterID)
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

	return enqueueCacheInvalidation(ctx.Context(), payload.Client, mut.Job, trustCenterID)
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

	return enqueueCacheInvalidation(ctx.Context(), payload.Client, mut.Job, trustCenterID)
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
		if err := enqueueCacheInvalidation(ctx.Context(), payload.Client, mut.Job, tcID); err != nil {
			logx.FromContext(ctx.Context()).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to enqueue cache invalidation for subprocessor")
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
		if err := enqueueCacheInvalidation(ctx.Context(), payload.Client, mut.Job, tcID); err != nil {
			logx.FromContext(ctx.Context()).Warn().Err(err).Str("trust_center_id", tcID).Msg("failed to enqueue cache invalidation for standard")
		}
	}

	return nil
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

// enqueueCacheInvalidation enqueues a job to invalidate the trust center cache
func enqueueCacheInvalidation(ctx context.Context, client *entgen.Client, jobClient riverqueue.JobClient, trustCenterID string) error {
	if trustCenterID == "" {
		return nil
	}

	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(trustCenterID)).
		Select(trustcenter.FieldCustomDomainID, trustcenter.FieldSlug).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to query trust center for cache invalidation")
		return nil
	}

	var customDomain string
	if tc.CustomDomainID != nil {
		cd, err := client.CustomDomain.Query().
			Where(customdomain.ID(*tc.CustomDomainID)).
			Select(customdomain.FieldCnameRecord).
			Only(ctx)
		if err == nil && cd.CnameRecord != "" {
			customDomain = cd.CnameRecord
		}
	}

	args := jobspec.ClearTrustCenterCacheArgs{}

	if customDomain != "" {
		args.CustomDomain = customDomain
	}

	if tc.Slug != "" {
		args.TrustCenterSlug = tc.Slug
	}

	if args.CustomDomain == "" && args.TrustCenterSlug == "" {
		return nil
	}

	if jobClient == nil {
		logx.FromContext(ctx).Warn().Msg("no job client available, skipping cache invalidation job")
		return nil
	}

	_, err = jobClient.Insert(ctx, args, nil)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("failed to insert clear trust center cache job")
		return nil
	}

	return nil
}
