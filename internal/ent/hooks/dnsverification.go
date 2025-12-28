package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/jobspec"
)

// HookDNSVerificationDelete cleans up preview domain DNS records when a verification record is deleted
func HookDNSVerificationDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.DNSVerificationFunc(func(ctx context.Context, m *generated.DNSVerificationMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			id, ok := m.ID()
			if !ok {
				return next.Mutate(ctx, m)
			}

			if trustCenterConfig.PreviewZoneID == "" {
				return next.Mutate(ctx, m)
			}

			customDomains, err := m.Client().CustomDomain.Query().
				Where(customdomain.DNSVerificationID(id)).
				Select(customdomain.FieldID).
				All(ctx)
			if err != nil {
				return nil, err
			}

			for _, cd := range customDomains {
				exists, err := m.Client().TrustCenter.Query().
					Where(trustcenter.PreviewDomainID(cd.ID)).
					Exist(ctx)
				if err != nil || !exists {
					if err != nil {
						return nil, err
					}
					continue
				}

				if err := enqueueJob(ctx, m.Job, jobspec.DeletePreviewDomainArgs{
					CustomDomainID:           cd.ID,
					TrustCenterPreviewZoneID: trustCenterConfig.PreviewZoneID,
				}, nil); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpDelete|ent.OpDeleteOne)
}
