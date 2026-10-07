package auth

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// OAuthRegistrationOptions describes how one definition maps shared OAuth mechanics to its local credential type
type OAuthRegistrationOptions[T any] struct {
	// Config describes the provider OAuth endpoints and request parameters
	Config OAuthConfig
	// Material maps shared OAuth material to the definition-local credential payload
	Material func(OAuthMaterial) (T, error)
	// EncodeCredentialError is returned when the typed credential cannot be serialized
	EncodeCredentialError error
}

// OAuthRegistration adapts the shared OAuth transport flow to one typed auth flow
func OAuthRegistration[T any](opts OAuthRegistrationOptions[T]) types.AuthFlow[T] {
	return types.NewAuthFlow[T](
		func(ctx context.Context, _ json.RawMessage) (types.AuthStartResult, error) {
			return StartOAuth(ctx, opts.Config)
		},
		func(ctx context.Context, state json.RawMessage, input types.AuthCallbackInput) (types.AuthCompleteResult, error) {
			material, err := CompleteOAuth(ctx, opts.Config, state, input)
			if err != nil {
				return types.AuthCompleteResult{}, err
			}

			credential, err := opts.Material(material)
			if err != nil {
				return types.AuthCompleteResult{}, err
			}

			data, err := jsonx.ToRawMessage(credential)
			if err != nil {
				if opts.EncodeCredentialError != nil {
					return types.AuthCompleteResult{}, opts.EncodeCredentialError
				}

				return types.AuthCompleteResult{}, err
			}

			return types.AuthCompleteResult{
				Credential: types.CredentialSet{Data: data},
			}, nil
		},
	)
}
