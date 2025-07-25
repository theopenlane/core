package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customdomain"
	"github.com/theopenlane/core/internal/ent/generated/dnsverification"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/mappabledomain"
	"github.com/theopenlane/core/pkg/corejobs"
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

				_, err = m.Job.Insert(ctx, corejobs.CreateCustomDomainArgs{
					CustomDomainID: id,
				}, nil)

				return v, err
			})
		},
		hook.HasOp(ent.OpCreate),
	)
}

// HookCustomDomain runs on create mutations
func HookDeleteCustomDomain() ent.Hook {
	return hook.If(
		func(next ent.Mutator) ent.Mutator {
			return hook.CustomDomainFunc(func(ctx context.Context, m *generated.CustomDomainMutation) (generated.Value, error) {
				if !isDeleteOp(ctx, m) {
					return next.Mutate(ctx, m)
				}

				id, ok := m.ID()
				if !ok {
					return nil, fmt.Errorf("%w: %s", ErrInvalidInput, "id is required")
				}

				cd, err := m.Client().CustomDomain.Query().Where(customdomain.ID(id)).
					Select(customdomain.FieldDNSVerificationID, customdomain.FieldMappableDomainID, customdomain.FieldDNSVerificationID).
					Only(ctx)
				if err != nil || cd.DNSVerificationID == "" {
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

				_, err = m.Job.Insert(ctx, corejobs.DeleteCustomDomainArgs{
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
