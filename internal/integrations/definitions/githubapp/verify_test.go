package githubapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/httpsling"
)

// TestVerifyDerivesMetadataFromCredential verifies the verification probes the client and returns the credential's installation identity
func TestVerifyDerivesMetadataFromCredential(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
		_, _ = w.Write([]byte(`{"data":{"viewer":{"repositories":{"nodes":[],"pageInfo":{"endCursor":"","hasNextPage":false}}}}}`))
	}))
	defer server.Close()

	client, err := newGraphQLClient(server.Client(), server.URL)
	require.NoError(t, err)

	meta, err := verify(context.Background(), types.ConnectionRequest[githubAppCredential]{
		Credential: githubAppCredential{AppID: 1, InstallationID: 222, AccessToken: "token", OrganizationName: "cred-org"},
	}, client)
	require.NoError(t, err)
	require.Equal(t, "222", meta.InstallationID)
	require.Equal(t, "cred-org", meta.OrganizationName)
	require.Equal(t, "222", meta.InstallationIdentity().ExternalID)
}
