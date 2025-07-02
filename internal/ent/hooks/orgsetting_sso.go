package hooks

import (
	"context"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/enums"
)

// HookValidateIdentityProviderConfig ensures identity provider configuration is present when SSO login is enforced
func HookValidateIdentityProviderConfig() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationSettingFunc(func(ctx context.Context, m *generated.OrganizationSettingMutation) (generated.Value, error) {
			if err := ValidateIdentityProviderConfig(ctx, m); err != nil {
				return nil, err
			}
			return next.Mutate(ctx, m)
		})
	}, hook.And(
		hook.HasFields("identity_provider_login_enforced"),
		hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne),
	))
}

// ValidateIdentityProviderConfig validates that SSO fields are set when login enforcement is enabled
func ValidateIdentityProviderConfig(ctx context.Context, m *generated.OrganizationSettingMutation) error {
	enforced, ok := m.IdentityProviderLoginEnforced()
	if !ok || !enforced {
		return nil
	}

	provider, providerOK := m.IdentityProvider()
	if !providerOK && m.Op() != ent.OpCreate {
		if old, err := m.OldIdentityProvider(ctx); err == nil {
			provider = old
			providerOK = true
		}
	}

	id, idOK := m.IdentityProviderClientID()
	if !idOK && m.Op() != ent.OpCreate {
		if old, err := m.OldIdentityProviderClientID(ctx); err == nil && old != nil {
			id = *old
			idOK = true
		}
	}

	secret, secretOK := m.IdentityProviderClientSecret()
	if !secretOK && m.Op() != ent.OpCreate {
		if old, err := m.OldIdentityProviderClientSecret(ctx); err == nil && old != nil {
			secret = *old
			secretOK = true
		}
	}

	endpoint, endpointOK := m.OidcDiscoveryEndpoint()
	if !endpointOK && m.Op() != ent.OpCreate {
		if old, err := m.OldOidcDiscoveryEndpoint(ctx); err == nil && old != "" {
			endpoint = old
			endpointOK = true
		}
	}

	if !providerOK || provider == enums.SSOProviderNone || !idOK || id == "" || !secretOK || secret == "" || !endpointOK || endpoint == "" {
		return fmt.Errorf("%w: identity provider configuration missing", ErrInvalidInput)
	}

	return nil
}
