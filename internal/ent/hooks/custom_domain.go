package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

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

// HookDeleteCustomDomain runs on update and delete mutations
func HookDeleteCustomDomain() ent.Hook {
	return hook.If(
		func(next ent.Mutator) ent.Mutator {
			return hook.CustomDomainFunc(func(ctx context.Context, m *generated.CustomDomainMutation) (generated.Value, error) {
				logx.FromContext(ctx).Debug().Msg("custom domain delete hook")

				if !isDeleteOp(ctx, m) {
					if !auth.IsSystemAdminFromContext(ctx) {
						logx.FromContext(ctx).Warn().Msg("only system admins can update custom domains")

						return nil, generated.ErrPermissionDenied
					}

					if !m.Op().Is(ent.OpUpdateOne) {
						return next.Mutate(ctx, m)
					}

					newCname, ok := m.CnameRecord()

					var oldCname string
					if ok {
						var err error

						oldCname, err = m.OldCnameRecord(ctx)
						if err != nil {
							return nil, err
						}
					}

					v, err := next.Mutate(ctx, m)
					if err != nil {
						return v, err
					}

					if !ok || oldCname == newCname {
						return v, nil
					}

					id, err := GetObjectIDFromEntValue(v)
					if err != nil {
						return v, err
					}

					if err := enqueueJob(ctx, m.Job, jobspec.ValidateCustomDomainArgs{
						CustomDomainID: id,
					}, nil); err != nil {
						return nil, err
					}

					trustCenters, err := m.Client().TrustCenter.Query().
						Where(trustcenter.Or(
							trustcenter.HasCustomDomainWith(customdomain.ID(id)),
							trustcenter.HasPreviewDomainWith(customdomain.ID(id)),
						)).
						All(ctx)
					if err != nil {
						return nil, err
					}

					for _, tc := range trustCenters {
						if tc.PirschDomainID == "" {
							continue
						}

						if err := enqueueJob(ctx, m.Job, jobspec.UpdatePirschDomainArgs{
							TrustCenterID: tc.ID,
						}, nil); err != nil {
							return nil, err
						}
					}

					return v, nil
				}

				id, ok := m.ID()
				if !ok {
					return nil, fmt.Errorf("%w: %s", ErrInvalidInput, "id is required")
				}

				cd, err := m.Client().CustomDomain.Query().Where(customdomain.ID(id)).
					Select(customdomain.FieldDNSVerificationID, customdomain.FieldMappableDomainID, customdomain.FieldDNSVerificationID).
					Only(ctx)
				if err != nil {
					return nil, err
				}

				trustCenters, err := m.Client().TrustCenter.Query().
					Where(trustcenter.Or(
						trustcenter.HasCustomDomainWith(customdomain.ID(id)),
						trustcenter.HasPreviewDomainWith(customdomain.ID(id)),
					)).
					All(ctx)
				if err != nil {
					return nil, err
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
						return nil, err
					}
				}

				if cd.DNSVerificationID == "" {
					return next.Mutate(ctx, m)
				}

				mappableDomain, err := m.Client().MappableDomain.Query().
					Where(mappabledomain.ID(cd.MappableDomainID)).
					Select(mappabledomain.FieldZoneID).
					Only(ctx)
				if err != nil {
					return nil, err
				}

				dnsVerification, err := m.Client().DNSVerification.Query().
					Where(dnsverification.ID(cd.DNSVerificationID)).
					Select(dnsverification.FieldCloudflareHostnameID).
					Only(ctx)
				if err != nil {
					return nil, err
				}

				err = enqueueJob(ctx, m.Job, jobspec.DeleteCustomDomainArgs{
					CustomDomainID:             id,
					DNSVerificationID:          cd.DNSVerificationID,
					CloudflareCustomHostnameID: dnsVerification.CloudflareHostnameID,
					CloudflareZoneID:           mappableDomain.ZoneID,
				}, nil)
				if err != nil {
					return nil, err
				}

				return next.Mutate(ctx, m)
			})
		},
		hook.HasOp(ent.OpDeleteOne|ent.OpDelete|ent.OpUpdate|ent.OpUpdateOne),
	)
}
