package hooks

import (
	"context"

	"entgo.io/ent"
	"github.com/samber/lo"

	generated "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// sendDomainScanCreate emits a domain scan create event via the integration runtime on the ent client,
// deferring emission until commit if a transaction is active
func sendDomainScanCreate(ctx context.Context, client *generated.Client, organizationID string, domains []string, forceRefresh bool) {
	rt := intruntime.FromClient(ctx, client)
	if rt == nil {
		return
	}

	emit := func() {
		receipt := rt.Gala().EmitWithHeaders(ctx, operations.DomainScanCreateTopic, operations.DomainScanCreateEnvelope{
			OrganizationID: organizationID,
			Domains:        domains,
			ForceRefresh:   forceRefresh,
		}, gala.Headers{})
		if receipt.Err != nil {
			logx.FromContext(ctx).Error().Err(receipt.Err).Msg("unable to emit domain scan create event")
		}
	}

	tx := transactionFromContext(ctx)
	if tx == nil {
		emit()
		return
	}

	tx.OnCommit(func(next generated.Committer) generated.Committer {
		return generated.CommitFunc(func(ctx context.Context, tx *generated.Tx) error {
			err := next.Commit(ctx, tx)
			if err == nil {
				defer emit()
			}

			return err
		})
	})
}

// HookDomainScanUpdate triggers a new domain scan whenever domains are added to an organization's settings
func HookDomainScanUpdate() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationSettingFunc(func(ctx context.Context, m *generated.OrganizationSettingMutation) (ent.Value, error) {
			oldDomains, err := m.OldDomains(ctx)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get old domains")
				return nil, err
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			var candidates []string

			if domains, ok := m.Domains(); ok {
				candidates = append(candidates, domains...)
			}

			if appended, ok := m.AppendedDomains(); ok {
				candidates = append(candidates, appended...)
			}

			addedDomains := newDomains(oldDomains, candidates)
			if len(addedDomains) == 0 {
				return retVal, nil
			}

			orgID, err := getOrgIDFromSettingMutation(ctx, m, retVal)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("failed to get organization ID for domain scan update")
				return retVal, nil
			}

			sendDomainScanCreate(ctx, m.Client(), orgID, addedDomains, false)

			return retVal, nil
		})
	},
		hook.And(
			hook.Or(
				hook.HasFields(organizationsetting.FieldDomains),
				hook.HasAddedFields((organizationsetting.FieldDomains)),
			),
			hook.HasOp(ent.OpUpdateOne),
		),
	)
}

// newDomains returns the deduplicated domains in candidates that are not already present in oldDomains
func newDomains(oldDomains, candidates []string) []string {
	added, _ := lo.Difference(lo.Uniq(candidates), oldDomains)
	return added
}
