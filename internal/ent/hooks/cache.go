package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/riverboat/pkg/riverqueue"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/corejobs"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
)

func insertClearCacheJob(ctx context.Context, jobClient riverqueue.JobClient, m utils.GenericMutation, trustCenterID string) error {
	client := m.Client()

	tc, err := client.TrustCenter.Query().
		Where(trustcenter.ID(trustCenterID)).
		Select(trustcenter.FieldCustomDomainID, trustcenter.FieldSlug).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Str("trust_center_id", trustCenterID).Msg("failed to query trust center for cache invalidation")
		return err
	}

	var customDomain string
	if tc.CustomDomainID != "" {
		cd, err := client.CustomDomain.Query().
			Where(customdomain.ID(tc.CustomDomainID)).
			Select(customdomain.FieldCnameRecord).
			Only(ctx)
		if err == nil && cd.CnameRecord != "" {
			customDomain = cd.CnameRecord
		}
	}

	args := corejobs.ClearTrustCenterCacheArgs{}

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

// HookModuleCacheInvalidation handles cache invalidation for the trustcenter data
// cached in objectstorage
func HookModuleCacheInvalidation() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mutationType := m.Type()
			genericMut, ok := m.(utils.GenericMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			var (
				trustCenterID    string
				jobClient        riverqueue.JobClient
				shouldClearCache bool
			)

			switch mutationType {
			case generated.TypeTrustCenterDoc:
				docMut, ok := m.(*generated.TrustCenterDocMutation)
				if !ok {
					return next.Mutate(ctx, m)
				}

				jobClient = docMut.Job

				if tcID, exists := docMut.TrustCenterID(); exists {
					trustCenterID = tcID
				} else if m.Op() == ent.OpDelete || m.Op() == ent.OpDeleteOne {
					if oldTCID, err := docMut.OldTrustCenterID(ctx); err == nil {
						trustCenterID = oldTCID
					}
				}

				// check if document visibility is being changed
				visibility, visibilityChanged := docMut.Visibility()

				// do not trigger if only non-visible documents are added
				if m.Op().Is(ent.OpCreate) {
					if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible ||
						visibility == enums.TrustCenterDocumentVisibilityProtected {
						shouldClearCache = true
					}
				}

				// for update operations: trigger if visibility changed to/from NOT_VISIBLE
				// or if visibility is currently public or private and it's being flipped
				if m.Op() == ent.OpUpdate || m.Op() == ent.OpUpdateOne {
					if visibilityChanged {
						oldVisibility, err := docMut.OldVisibility(ctx)
						if err == nil {
							if oldVisibility == enums.TrustCenterDocumentVisibilityNotVisible ||
								visibility == enums.TrustCenterDocumentVisibilityNotVisible {
								shouldClearCache = true
							}
						}
					}

					if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible ||
						visibility == enums.TrustCenterDocumentVisibilityProtected {
						shouldClearCache = true
					}
				}

				if m.Op() == ent.OpDelete || m.Op() == ent.OpDeleteOne {
					shouldClearCache = true
				}

			case generated.TypeNote:
				noteMut, ok := m.(*generated.NoteMutation)
				if ok {
					jobClient = noteMut.Job
					shouldClearCache = true

					tcIDs := noteMut.TrustCenterIDs()
					if len(tcIDs) > 0 {
						trustCenterID = tcIDs[0]
					}
				}

			case generated.TypeTrustcenterEntity:
				entityMut, ok := m.(*generated.TrustcenterEntityMutation)
				if ok {
					jobClient = entityMut.Job
					shouldClearCache = true

					if tcID, exists := entityMut.TrustCenterID(); exists {
						trustCenterID = tcID
					}
				}

			case generated.TypeSubprocessor:
				subMut := m.(*generated.SubprocessorMutation)
				jobClient = subMut.Job
				shouldClearCache = true

			case generated.TypeStandard:
				stdMut := m.(*generated.StandardMutation)
				jobClient = stdMut.Job
				shouldClearCache = true
			}

			if trustCenterID == "" {

				var err error
				trustCenterID, err = genericMut.Client().TrustCenter.Query().OnlyID(ctx)
				if generated.IsNotFound(err) {
					return nil, err
				}

				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("unable to fetch trustcenter")
					return nil, ErrTrustCenterIDRequired
				}
			}

			if shouldClearCache && trustCenterID != "" {
				_ = insertClearCacheJob(ctx, jobClient, genericMut, trustCenterID)
			}

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne))
}
