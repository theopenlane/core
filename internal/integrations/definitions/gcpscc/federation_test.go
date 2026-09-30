package gcpscc

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
)

const (
	// testProjectNumber is a representative numeric GCP project number hosting the workload identity pool
	testProjectNumber = "123456789"
	// testRSAKeySize is the smallest RSA key size the signing key loader accepts
	testRSAKeySize = 2048
)

// testBuildRequest builds a client build request with a signing token manager
func testBuildRequest(t *testing.T) types.ClientBuildRequest {
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

// TestWorkloadIdentityAudience verifies the provider resource name construction
func TestWorkloadIdentityAudience(t *testing.T) {
	require.Equal(t,
		"//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/openlane/providers/openlane",
		workloadIdentityAudience(testProjectNumber),
	)
}

// TestFederationSource verifies the workload identity token source construction
func TestFederationSource(t *testing.T) {
	t.Run("builds a federated source without impersonation", func(t *testing.T) {
		source, err := federationSource(context.Background(), testBuildRequest(t), WorkloadIdentityCredentialSchema{
			ProjectNumber: testProjectNumber,
		})
		require.NoError(t, err)
		require.NotNil(t, source)
	})

	t.Run("builds an impersonated source when a service account is configured", func(t *testing.T) {
		source, err := federationSource(context.Background(), testBuildRequest(t), WorkloadIdentityCredentialSchema{
			ProjectNumber:       testProjectNumber,
			ServiceAccountEmail: "collector@project-123.iam.gserviceaccount.com",
		})
		require.NoError(t, err)
		require.NotNil(t, source)
	})

	t.Run("requires a project number", func(t *testing.T) {
		_, err := federationSource(context.Background(), testBuildRequest(t), WorkloadIdentityCredentialSchema{})
		require.ErrorIs(t, err, ErrProjectNumberRequired)
	})
}
