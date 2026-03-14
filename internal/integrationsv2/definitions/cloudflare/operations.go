package cloudflare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/user"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrationsv2/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

type cloudflareHealthDetails struct {
	Status    string `json:"status,omitempty"`
	ExpiresOn string `json:"expiresOn,omitempty"`
}

// buildCloudflareClient builds the Cloudflare API client for one installation
func buildCloudflareClient(_ context.Context, req types.ClientBuildRequest) (any, error) {
	var cred credential
	if err := jsonx.UnmarshalIfPresent(req.Credential.ProviderData, &cred); err != nil {
		return nil, err
	}

	if cred.APIToken == "" {
		return nil, ErrAPITokenMissing
	}

	return cf.NewClient(option.WithAPIToken(cred.APIToken)), nil
}

// runHealthOperation verifies the Cloudflare API token via /user/tokens/verify
func runHealthOperation(ctx context.Context, _ *generated.Integration, _ types.CredentialSet, client any, _ json.RawMessage) (json.RawMessage, error) {
	cfClient, ok := client.(*cf.Client)
	if !ok {
		return nil, ErrClientType
	}

	res, err := cfClient.User.Tokens.Verify(ctx)
	if err != nil {
		return nil, fmt.Errorf("cloudflare: token verification failed: %w", err)
	}

	if res.Status != user.TokenVerifyResponseStatusActive {
		return nil, errors.New("cloudflare: token is not active")
	}

	details := cloudflareHealthDetails{
		Status: string(res.Status),
	}

	if !res.ExpiresOn.IsZero() {
		details.ExpiresOn = res.ExpiresOn.String()
	}

	return jsonx.ToRawMessage(details)
}
