package microsoft

import (
	"context"
	"encoding/json"
	"net/url"

	iamauth "github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// adminConsentState carries the CSRF token between consent start and callback
type adminConsentState struct {
	State string `json:"state"`
}

// adminConsent holds the per-definition inputs for the admin consent flow
type adminConsent[T any] struct {
	// cfg is the operator-level Azure application configuration
	cfg AppConfig
	// newCredential builds the definition's own credential from the consented tenant ID and the consented application ID
	newCredential func(tenantID, applicationID string) T
}

// AdminConsentFlow builds an auth flow that uses the Azure admin consent endpoint
// which grants application permissions so the client credentials flow can access the tenant
func AdminConsentFlow[T any](cfg AppConfig, newCredential func(tenantID, applicationID string) T) types.AuthFlow[T] {
	consent := adminConsent[T]{cfg: cfg, newCredential: newCredential}

	return types.NewAuthFlow[T](consent.start, consent.complete)
}

// start builds the admin consent URL and CSRF state
func (a adminConsent[T]) start(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
	csrfState, err := iamauth.GenerateOAuthState(0)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error generating oauth state")
		return types.AuthStartResult{}, ErrConsentStateGeneration
	}

	params := url.Values{
		"client_id":    {a.cfg.ApplicationID},
		"redirect_uri": {a.cfg.RedirectURL},
		"state":        {csrfState},
		"scope":        {GraphScope},
	}

	stateData, err := jsonx.ToRawMessage(adminConsentState{State: csrfState})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error encoding oauth state")
		return types.AuthStartResult{}, ErrConsentStateGeneration
	}

	tenant := a.cfg.DefaultTenant
	if tenant == "" {
		tenant = OrganizationsTenant
	}

	consentURL := AdminConsentURL(tenant) + "?" + params.Encode()

	logx.FromContext(ctx).Debug().Str("url", consentURL).Msg("admin consent started")

	return types.AuthStartResult{
		URL:   consentURL,
		State: stateData,
	}, nil
}

// complete validates the admin consent callback and stores the consented tenant as the definition credential
func (a adminConsent[T]) complete(ctx context.Context, state json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
	var savedState adminConsentState
	if err := jsonx.UnmarshalIfPresent(state, &savedState); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("unable to parse consent state")
		return types.AuthCompleteResult{}, ErrConsentStateInvalid
	}

	if savedState.State != "" && input.First("state") != savedState.State {
		return types.AuthCompleteResult{}, ErrConsentStateMismatch
	}

	if errCode := input.First("error"); errCode != "" {
		logx.FromContext(ctx).Error().Str("error", errCode).Str("description", input.First("error_description")).Msg("admin consent denied")
		return types.AuthCompleteResult{}, ErrConsentDenied
	}

	tenantID := input.First("tenant")
	if tenantID == "" {
		return types.AuthCompleteResult{}, ErrTenantIDNotFound
	}

	data, err := jsonx.ToRawMessage(a.newCredential(tenantID, a.cfg.ApplicationID))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error encoding auth result")
		return types.AuthCompleteResult{}, ErrCredentialEncode
	}

	logx.FromContext(ctx).Debug().Str("tenant_id", tenantID).Msg("admin consent complete")

	return types.AuthCompleteResult{
		Credential: types.CredentialSet{Data: data},
	}, nil
}
