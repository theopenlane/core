package azureentraid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	iamauth "github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

const adminConsentBaseURL = "https://login.microsoftonline.com/%s/v2.0/adminconsent"

// adminConsentState carries the CSRF token between consent start and callback
type adminConsentState struct {
	State string `json:"state"`
}

// adminConsentRegistration builds an auth registration that uses the Azure admin consent endpoint
// which grants application permissions so the client credentials flow can access the tenant directory
func adminConsentRegistration(cfg Config) *types.AuthRegistration {
	return &types.AuthRegistration{
		CredentialRef: entraTenantCredential.ID(),
		Start: func(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
			csrfState, err := iamauth.GenerateOAuthState(0)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error generating oauth state")
				return types.AuthStartResult{}, ErrConsentStateGeneration
			}

			params := url.Values{
				"client_id":    {cfg.ClientID},
				"redirect_uri": {cfg.RedirectURL},
				"state":        {csrfState},
				"scope":        {graphScope},
			}

			stateData, err := jsonx.ToRawMessage(adminConsentState{State: csrfState})
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error encoding oauth state")
				return types.AuthStartResult{}, ErrConsentStateGeneration
			}

			tenant := cfg.DefaultTenant
			if tenant == "" {
				tenant = "organizations"
			}

			consentURL := fmt.Sprintf(adminConsentBaseURL, tenant) + "?" + params.Encode()

			logx.FromContext(ctx).Debug().Str("url", consentURL).Msg("admin consent started")

			return types.AuthStartResult{
				URL:   consentURL,
				State: stateData,
			}, nil
		},
		Complete: func(ctx context.Context, state json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
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

			data, err := jsonx.ToRawMessage(entraIDCred{TenantID: tenantID})
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Msg("error encoding auth result")
				return types.AuthCompleteResult{}, ErrCredentialEncode
			}

			logx.FromContext(ctx).Debug().Str("tenant_id", tenantID).Msg("admin consent complete")

			return types.AuthCompleteResult{
				Credential: types.CredentialSet{Data: data},
			}, nil
		},
	}
}
