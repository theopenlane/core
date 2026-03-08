package azuresecuritycenter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"golang.org/x/oauth2"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip satisfies http.RoundTripper for inline test transports.
func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// TestAzureSecurityCenterMint_ClientCredentials validates client-credential token exchange.
func TestAzureSecurityCenterMint_ClientCredentials(t *testing.T) {
	transport := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		if got := values.Get("grant_type"); got != "client_credentials" {
			t.Errorf("grant_type = %q", got)
		}
		if got := values.Get("client_id"); got != "client-id" {
			t.Errorf("client_id = %q", got)
		}
		if got := values.Get("client_secret"); got != "client-secret" {
			t.Errorf("client_secret = %q", got)
		}

		payload, err := json.Marshal(map[string]any{
			"access_token": "token-value",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(payload))),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})
	client := &http.Client{Transport: transport}
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)

	schema, err := jsonx.ToRawMessage(map[string]any{"type": "object"})
	require.NoError(t, err)

	spec := config.ProviderSpec{
		AuthType:          types.AuthKindOAuth2ClientCredentials,
		CredentialsSchema: schema,
	}

	provider, err := Builder().Build(context.Background(), spec)
	require.NoError(t, err)

	azureProvider, ok := provider.(*Provider)
	require.True(t, ok)

	var capturedTenant types.TrimmedString
	azureProvider.tokenEndpoint = func(tenantID types.TrimmedString) string {
		capturedTenant = tenantID
		return "https://example.com/token"
	}

	payload := models.CredentialSet{
		ProviderData: json.RawMessage(`{
			"tenantId":"tenant-123",
			"clientId":"client-id",
			"clientSecret":"client-secret",
			"subscriptionId":"sub-456"
		}`),
	}

	result, err := azureProvider.Mint(ctx, types.CredentialMintRequest{
		Provider:   TypeAzureSecurityCenter,
		Credential: payload,
	})
	require.NoError(t, err)
	require.Equal(t, types.TrimmedString("tenant-123"), capturedTenant)
	require.Equal(t, "token-value", result.OAuthAccessToken)
	require.Equal(t, "Bearer", result.OAuthTokenType)
	require.Equal(t, "client-id", result.ClientID)
	require.Equal(t, "client-secret", result.ClientSecret)
	require.JSONEq(t, `{"tenantId":"tenant-123","subscriptionId":"sub-456"}`, string(result.ProviderData))
}
