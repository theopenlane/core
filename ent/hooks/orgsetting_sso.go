package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/samber/lo"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/hook"
	"github.com/theopenlane/shared/enums"
)

// HookValidateIdentityProviderConfig ensures identity provider configuration is present when SSO login is enforced
// and resets enforced/tested status when SSO configuration fields change
func HookValidateIdentityProviderConfig() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return hook.OrganizationSettingFunc(func(ctx context.Context, m *generated.OrganizationSettingMutation) (generated.Value, error) {
			if err := disableEnforcementIfConfigChanged(ctx, m); err != nil {
				return nil, err
			}

			if err := ValidateIdentityProviderConfig(ctx, m); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, hook.And(
		hook.HasFields("identity_provider", "identity_provider_client_id",
			"identity_provider_client_secret", "oidc_discovery_endpoint",
			"identity_provider_login_enforced"),
		hook.HasOp(ent.OpCreate|ent.OpUpdateOne),
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

	if isStringEmpty(idOK, id) {
		return ErrInvalidInput
	}

	// Client Secret
	secret, secretOK := fallbackString(
		m.IdentityProviderClientSecret,
		func() (*string, error) { return m.OldIdentityProviderClientSecret(ctx) },
		m.Op() != ent.OpCreate,
	)

	if isStringEmpty(secretOK, secret) {
		return ErrInvalidInput
	}
	// OIDC Discovery Endpoint
	endpoint, endpointOK := fallbackString(
		m.OidcDiscoveryEndpoint,
		func() (*string, error) { return stringPtrFromOld(ctx, m.OldOidcDiscoveryEndpoint) },
		m.Op() != ent.OpCreate,
	)

	if isStringEmpty(endpointOK, endpoint) {
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

	return lo.ToPtr(val), nil
}

// missingProvider checks if the SSO provider is missing or set to None
func missingProvider(ok bool, provider enums.SSOProvider) bool {
	return !ok || provider == enums.SSOProviderNone
}

// isStringEmpty checks if a string value is missing or empty
func isStringEmpty(ok bool, value string) bool {
	return !ok || value == ""
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

// disableEnforcementIfConfigChanged makes sure if any value changes in the idp config,
// the enforcement is disabled. The "tested" value is turned off too so the user/org is forced
// to test the connection before they can enable it again
func disableEnforcementIfConfigChanged(ctx context.Context, m *generated.OrganizationSettingMutation) error {
	if m.Op() == ent.OpCreate || didSSOConfigChange(ctx, m) {
		m.SetIdentityProviderLoginEnforced(false)
		m.SetIdentityProviderAuthTested(false)
	}

	return nil
}

func didSSOConfigChange(ctx context.Context, m *generated.OrganizationSettingMutation) bool {
	if enforced, ok := m.IdentityProviderLoginEnforced(); ok {
		if oldEnforcement, err := m.OldIdentityProviderLoginEnforced(ctx); err == nil && oldEnforcement && !enforced {
			return false
		}
	}

	if provider, ok := m.IdentityProvider(); ok {
		if oldProvider, err := m.OldIdentityProvider(ctx); err == nil && provider != oldProvider {
			return true
		}
	}

	if clientID, ok := m.IdentityProviderClientID(); ok {
		if oldClientID, err := m.OldIdentityProviderClientID(ctx); err == nil && oldClientID != nil && clientID != *oldClientID {
			return true
		}
	}

	if clientSecret, ok := m.IdentityProviderClientSecret(); ok {
		if oldClientSecret, err := m.OldIdentityProviderClientSecret(ctx); err == nil && oldClientSecret != nil && clientSecret != *oldClientSecret {
			return true
		}
	}

	if endpoint, ok := m.OidcDiscoveryEndpoint(); ok {
		if oldEndpoint, err := m.OldOidcDiscoveryEndpoint(ctx); err == nil && endpoint != oldEndpoint {
			return true
		}
	}

	return false
}
