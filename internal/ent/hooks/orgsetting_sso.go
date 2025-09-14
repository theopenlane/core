package hooks

import (
	"context"

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

// ValidateIdentityProviderConfig checks if the identity provider configuration is valid
// the intent of the function is to ensure all necessary identity provider configuration fields are present and
// valid when SSO enforcement is being set to active, while also supporting partial updates by falling back
// to existing values when appropriate
func ValidateIdentityProviderConfig(ctx context.Context, m *generated.OrganizationSettingMutation) error {
	enforced, ok := m.IdentityProviderLoginEnforced()
	if !ok || !enforced {
		return nil
	}

	// sso connection must have been tested
	// before it can be can be enforced
	tested, _ := m.OldIdentityProviderAuthTested(ctx)
	if !tested {
		return ErrSSONotEnforceable
	}

	// identity provider
	provider, providerOK := m.IdentityProvider()
	if !providerOK && m.Op() != ent.OpCreate {
		if old, err := m.OldIdentityProvider(ctx); err == nil {
			provider = old
			providerOK = true
		}
	}

	if missingProvider(providerOK, provider) {
		return ErrInvalidInput
	}

	// Client ID
	id, idOK := fallbackString(
		m.IdentityProviderClientID,
		func() (*string, error) { return m.OldIdentityProviderClientID(ctx) },
		m.Op() != ent.OpCreate,
	)

	if missingClientID(idOK, id) {
		return ErrInvalidInput
	}

	// Client Secret
	secret, secretOK := fallbackString(
		m.IdentityProviderClientSecret,
		func() (*string, error) { return m.OldIdentityProviderClientSecret(ctx) },
		m.Op() != ent.OpCreate,
	)

	if missingClientSecret(secretOK, secret) {
		return ErrInvalidInput
	}
	// OIDC Discovery Endpoint
	endpoint, endpointOK := fallbackString(
		m.OidcDiscoveryEndpoint,
		func() (*string, error) { return stringPtrFromOld(ctx, m.OldOidcDiscoveryEndpoint) },
		m.Op() != ent.OpCreate,
	)

	if missingDiscoveryEndpoint(endpointOK, endpoint) {
		return ErrInvalidInput
	}

	return nil
}

// stringPtrFromOld converts a function that returns a string and an error into a pointer to a string
// This is used to handle the case where the old value might not be set, returning nil if there's an error
// or if the value is empty
func stringPtrFromOld(ctx context.Context, old func(ctx context.Context) (string, error)) (*string, error) {
	val, err := old(ctx)
	if err != nil {
		return nil, err
	}

	return &val, nil
}

// missingProvider checks if the SSO provider is missing or set to None
func missingProvider(ok bool, provider enums.SSOProvider) bool {
	return !ok || provider == enums.SSOProviderNone
}

// missingClientID checks if the OIDC client ID is missing or empty
func missingClientID(ok bool, id string) bool {
	return !ok || id == ""
}

// missingClientSecret checks if the OIDC client secret is missing or empty
func missingClientSecret(ok bool, secret string) bool {
	return !ok || secret == ""
}

// missingDiscoveryEndpoint checks if the OIDC discovery endpoint is missing or empty
func missingDiscoveryEndpoint(ok bool, endpoint string) bool {
	return !ok || endpoint == ""
}

// fallbackString attempts to get a value from the primary function, and if it fails, it tries the fallback function
// the pattern is intended to ensure that when updating a record, the system can use the existing Client ID if a new one isn't
// provided, but when creating a new record, the value must be explicitly set
func fallbackString(primary func() (string, bool), fallback func() (*string, error), allowFallback bool) (string, bool) {
	val, ok := primary()

	if ok {
		return val, true
	}

	if allowFallback {
		if old, err := fallback(); err == nil && old != nil {
			return *old, true
		}
	}

	return "", false
}
