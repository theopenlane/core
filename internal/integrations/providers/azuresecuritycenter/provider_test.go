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

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
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

	spec := config.ProviderSpec{
		AuthType:          types.AuthKindOAuth2,
		CredentialsSchema: map[string]any{"type": "object"},
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

	payload, err := types.NewCredentialBuilder(TypeAzureSecurityCenter).With(
		types.WithCredentialKind(types.CredentialKindMetadata),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: map[string]any{
				"tenantId":       "tenant-123",
				"clientId":       "client-id",
				"clientSecret":   "client-secret",
				"subscriptionId": "sub-456",
			},
		}),
	).Build()
	require.NoError(t, err)

	result, err := azureProvider.Mint(ctx, types.CredentialSubject{
		Provider:   TypeAzureSecurityCenter,
		Credential: payload,
	})
	require.NoError(t, err)
	require.Equal(t, types.TrimmedString("tenant-123"), capturedTenant)
	require.NotNil(t, result.Token)
	require.Equal(t, "token-value", result.Token.AccessToken)
	require.Equal(t, "Bearer", result.Token.TokenType)
	require.Equal(t, "sub-456", result.Data.ProviderData["subscriptionId"])
}
