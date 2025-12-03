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

// HookTrustcenterCacheInvalidation handles cache invalidation for the trustcenter data
// cached in objectstorage
func HookTrustcenterCacheInvalidation() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mutationType := m.Type()
			genericMut, ok := m.(utils.GenericMutation)
			if !ok {
				return next.Mutate(ctx, m)
			}

			var (
				trustCenterIDs              []string
				jobClient                   riverqueue.JobClient
				shouldClearCache            bool
				shouldUseTrustCenterFromOrg bool
			)

			switch mutationType {
			case generated.TypeTrustCenterDoc:
				docMut, ok := m.(*generated.TrustCenterDocMutation)
				if !ok {
					return next.Mutate(ctx, m)
				}

				jobClient = docMut.Job

				if tcID, exists := docMut.TrustCenterID(); exists {
					trustCenterIDs = append(trustCenterIDs, tcID)
				} else {
					if oldTCID, err := docMut.OldTrustCenterID(ctx); err == nil {
						trustCenterIDs = append(trustCenterIDs, oldTCID)
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
						trustCenterIDs = append(trustCenterIDs, tcIDs...)
					}
				}

			case generated.TypeTrustcenterEntity:
				entityMut, ok := m.(*generated.TrustcenterEntityMutation)
				if ok {
					jobClient = entityMut.Job
					shouldClearCache = true

					if tcID, exists := entityMut.TrustCenterID(); exists {
						trustCenterIDs = append(trustCenterIDs, tcID)
					} else {
						if oldTCID, err := entityMut.OldTrustCenterID(ctx); err == nil {
							trustCenterIDs = append(trustCenterIDs, oldTCID)
						}
					}
				}

			case generated.TypeSubprocessor:
				subMut := m.(*generated.SubprocessorMutation)
				jobClient = subMut.Job
				shouldClearCache = true
				shouldUseTrustCenterFromOrg = true

			case generated.TypeStandard:
				stdMut := m.(*generated.StandardMutation)
				jobClient = stdMut.Job
				shouldClearCache = true
				shouldUseTrustCenterFromOrg = true
			}

			// tests have multiple trust centers because they bypass
			// with privacy checks. But in reality, if coming in through
			// graphapi ( which everything does ), it is impossible to create
			// multiple.
			//
			// TODO(adelowo): fix tests sometime. Will be a fairly large change set.
			// or we really could just do TrustCenter.First(). OnlyID won't work here really
			if len(trustCenterIDs) == 0 && shouldUseTrustCenterFromOrg {
				query := genericMut.Client().TrustCenter.Query()

				var err error
				trustCenterIDs, err = query.IDs(ctx)
				if generated.IsNotFound(err) {
					return next.Mutate(ctx, m)
				}

				if err != nil {
					logx.FromContext(ctx).Error().Err(err).
						Str("mutation_type", mutationType).
						Msg("unable to fetch trustcenter")
					return nil, ErrTrustCenterIDRequired
				}
			}

			if shouldClearCache && len(trustCenterIDs) > 0 {
				for _, tcID := range trustCenterIDs {
					_ = insertClearCacheJob(ctx, jobClient, genericMut, tcID)
				}
			}

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne))
}

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
	if tc.CustomDomainID != nil {
		cd, err := client.CustomDomain.Query().
			Where(customdomain.ID(*tc.CustomDomainID)).
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
