package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/iam/tokens"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/oidc"
)

const (
	// testFederationAudience is a representative workload identity pool provider resource name
	testFederationAudience = "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/openlane/providers/openlane"
	// testRSAKeySize is the smallest RSA key size the signing key loader accepts
	testRSAKeySize = 2048
)

// testFederationBuildRequest builds a client build request with a signing token manager
func testFederationBuildRequest(t *testing.T) types.ClientBuildRequest {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, testRSAKeySize)
	require.NoError(t, err)

	manager, err := tokens.NewWithKey(key, tokens.Config{
		Audience:        "https://api.theopenlane.io",
		Issuer:          "https://api.theopenlane.io",
		AccessDuration:  time.Hour,
		RefreshDuration: 2 * time.Hour,
		RefreshOverlap:  -15 * time.Minute,
	})
	require.NoError(t, err)

	return types.ClientBuildRequest{
		Integration:  &ent.Integration{OwnerID: "01JQZX9K8N7M6P5R4T3V2W1Y0Z"},
		TokenManager: manager,
	}
}

// TestFederatedTokenSource verifies the shared federation helper wiring and validation
func TestFederatedTokenSource(t *testing.T) {
	t.Run("builds a token source from the installation identity", func(t *testing.T) {
		source, err := FederatedTokenSource(context.Background(), testFederationBuildRequest(t), FederationSpec{
			Audience: testFederationAudience,
			Endpoint: "https://sts.googleapis.com/v1/token",
		})
		require.NoError(t, err)
		require.NotNil(t, source)
	})

	t.Run("propagates validation errors at build time", func(t *testing.T) {
		_, err := FederatedTokenSource(context.Background(), testFederationBuildRequest(t), FederationSpec{
			Endpoint: "https://sts.googleapis.com/v1/token",
		})
		require.ErrorIs(t, err, oidc.ErrAudienceRequired)

		_, err = FederatedTokenSource(context.Background(), testFederationBuildRequest(t), FederationSpec{
			Audience: testFederationAudience,
		})
		require.ErrorIs(t, err, oidc.ErrEndpointRequired)
	})
}
