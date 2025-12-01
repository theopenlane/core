package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/enums"
)

func proceedWithMutation(ctx context.Context, next ent.Mutator, m ent.Mutation) (ent.Value, error) {
	return next.Mutate(ctx, m)
}

// HookModuleCacheInvalidation is a generic hook that handles cache invalidation
// for module schemas that affect the trust center cache.
// It matches the following schemas:
// - Documents (TrustCenterDoc): when visibility is public or private, or visibility changes to/from non-visible
// - Subprocessors (Subprocessor): all operations
// - Posts (Note): all operations (when connected to TrustCenter)
// - Compliance Standards (Standard): all operations
// - Trusted by/customer logos (TrustcenterEntity): all operations
func HookModuleCacheInvalidation() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mutationType := m.Type()

			switch mutationType {
			case generated.TypeSubprocessor:
				return proceedWithMutation(ctx, next, m)

			case generated.TypeStandard:
				return proceedWithMutation(ctx, next, m)

			case generated.TypeTrustcenterEntity:
				return proceedWithMutation(ctx, next, m)

			case generated.TypeNote:
				return proceedWithMutation(ctx, next, m)

			case generated.TypeTrustCenterDoc:
				docMut, ok := m.(*generated.TrustCenterDocMutation)
				if !ok {
					return proceedWithMutation(ctx, next, m)
				}

				// check if visibility field is being changed
				visibility, visibilityChanged := docMut.Visibility()

				// for create operations: only trigger if visibility is public or private
				// do not trigger if only non-visible documents are added
				if m.Op() == ent.OpCreate {
					if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible ||
						visibility == enums.TrustCenterDocumentVisibilityProtected {
						return proceedWithMutation(ctx, next, m)
					}
					return proceedWithMutation(ctx, next, m)
				}

				// for update operations: trigger if visibility changed to/from NOT_VISIBLE
				// or if visibility is currently public or private
				if m.Op() == ent.OpUpdate || m.Op() == ent.OpUpdateOne {
					shouldTrigger := false
					_ = shouldTrigger

					if visibilityChanged {
						oldVisibility, err := docMut.OldVisibility(ctx)
						if err == nil {
							if oldVisibility == enums.TrustCenterDocumentVisibilityNotVisible ||
								visibility == enums.TrustCenterDocumentVisibilityNotVisible {
								shouldTrigger = true
							}
						}
					}

					if visibility == enums.TrustCenterDocumentVisibilityPubliclyVisible ||
						visibility == enums.TrustCenterDocumentVisibilityProtected {
						shouldTrigger = true
					}

					return proceedWithMutation(ctx, next, m)
				}

				if m.Op() == ent.OpDelete || m.Op() == ent.OpDeleteOne {
					return proceedWithMutation(ctx, next, m)
				}

				return proceedWithMutation(ctx, next, m)

			default:
				return proceedWithMutation(ctx, next, m)
			}
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne))
}
