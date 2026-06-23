package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/jobspec"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/dnsverification"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/mappabledomain"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/pkg/logx"
)

// HookCustomDomain runs on create mutations
func HookCreateCustomDomain() ent.Hook {
	return hook.If(
		func(next ent.Mutator) ent.Mutator {
			return hook.CustomDomainFunc(func(ctx context.Context, m *generated.CustomDomainMutation) (generated.Value, error) {
				domainType, ok := m.DomainType()
				if !ok || domainType == enums.CustomDomainTypeUnknown {
					m.SetDomainType(enums.CustomDomainTypeExternal)
				}

				v, err := next.Mutate(ctx, m)
				if err != nil {
					return v, err
				}

				id, err := GetObjectIDFromEntValue(v)
				if err != nil {
					return v, err
				}

				err = enqueueJob(ctx, m.Job, jobspec.CreateCustomDomainArgs{
					CustomDomainID: id,
				}, nil)

				return v, err
			})
		},
		hook.HasOp(ent.OpCreate),
	)
}

// HookDeleteCustomDomain runs on single and bulk deletions.
func HookDeleteCustomDomain() ent.Hook {
	return hook.If(
		func(next ent.Mutator) ent.Mutator {
			return hook.CustomDomainFunc(func(ctx context.Context, m *generated.CustomDomainMutation) (generated.Value, error) {
				logx.FromContext(ctx).Debug().Msg("custom domain delete hook")

				if !isDeleteOp(ctx, m) {
					// only allow system admin to update
					if !auth.IsSystemAdminFromContext(ctx) {
						logx.FromContext(ctx).Warn().Msg("only system admins can update custom domains")

						return nil, generated.ErrPermissionDenied
					}

					return next.Mutate(ctx, m)
				}

				var ids []string

				switch m.Op() {

				case ent.OpDelete, ent.OpUpdate:
					var err error
					ids, err = m.IDs(ctx)
					if err != nil {
						return nil, err
					}

				case ent.OpDeleteOne, ent.OpUpdateOne:

					id, ok := m.ID()
					if !ok {
						return nil, fmt.Errorf("%w: %s", ErrInvalidInput, "id is required")
					}

					ids = []string{id}
				}

				for _, id := range ids {
					if err := deleteCustomDomain(ctx, m, id); err != nil {
						return nil, err
					}
				}

				return next.Mutate(ctx, m)
			})
		},
		hook.HasOp(ent.OpDeleteOne|ent.OpDelete|ent.OpUpdate|ent.OpUpdateOne),
	)
}

func deleteCustomDomain(ctx context.Context, m *generated.CustomDomainMutation, id string) error {
	cd, err := m.Client().CustomDomain.Query().Where(customdomain.ID(id)).
		Select(customdomain.FieldDNSVerificationID, customdomain.FieldMappableDomainID).
		Only(ctx)
	if err != nil {
		return err
	}

	trustCenters, err := m.Client().TrustCenter.Query().
		Where(trustcenter.Or(
			trustcenter.HasCustomDomainWith(customdomain.ID(id)),
			trustcenter.HasPreviewDomainWith(customdomain.ID(id)),
		)).
		All(ctx)
	if err != nil {
		return err
	}

	for _, tc := range trustCenters {
		update := m.Client().TrustCenter.UpdateOneID(tc.ID)
		if tc.CustomDomainID != nil && *tc.CustomDomainID == id {
			update.ClearCustomDomain()
		}
		if tc.PreviewDomainID == id {
			update.ClearPreviewDomain()
		}

		if err = update.Exec(ctx); err != nil {
			return err
		}
	}

	if cd.DNSVerificationID == "" {
		return nil
	}

	mappableDomain, err := m.Client().MappableDomain.Query().
		Where(mappabledomain.ID(cd.MappableDomainID)).
		Select(mappabledomain.FieldZoneID).
		Only(ctx)
	if err != nil {
		return err
	}

	dnsVerification, err := m.Client().DNSVerification.Query().
		Where(dnsverification.ID(cd.DNSVerificationID)).
		Select(dnsverification.FieldCloudflareHostnameID).
		Only(ctx)
	if err != nil {
		return err
	}

	return enqueueJob(ctx, m.Job, jobspec.DeleteCustomDomainArgs{
		CustomDomainID:             id,
		DNSVerificationID:          cd.DNSVerificationID,
		CloudflareCustomHostnameID: dnsVerification.CloudflareHostnameID,
		CloudflareZoneID:           mappableDomain.ZoneID,
	}, nil)
}
