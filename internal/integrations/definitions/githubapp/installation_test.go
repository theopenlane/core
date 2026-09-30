package githubapp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// TestResolveInstallationMetadataFromCredential verifies the bound credential's installation id wins over
// callback input and stored metadata, so a credential for a different installation resolves to that installation
func TestResolveInstallationMetadataFromCredential(t *testing.T) {
	t.Parallel()

	meta, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{
		Integration: &generated.Integration{
			InstallationMetadata: types.IntegrationInstallationMetadata{
				Attributes: json.RawMessage(`{"installationId":"11","organizationName":"stored-org"}`),
			},
		},
		Credentials: types.CredentialBindings{{
			Ref: gitHubAppCredential.ID(),
			Credential: types.CredentialSet{
				Data: json.RawMessage(`{"appId":1,"installationId":222,"accessToken":"token","organizationName":"cred-org"}`),
			},
		}},
		Input: json.RawMessage(`{"installationId":"77","organizationName":"input-org"}`),
	})
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "222", meta.InstallationID)
	require.Equal(t, "cred-org", meta.OrganizationName)
}

// TestResolveInstallationMetadataFromInput verifies callback input wins when it carries an installation id
func TestResolveInstallationMetadataFromInput(t *testing.T) {
	t.Parallel()

	meta, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{
		Integration: &generated.Integration{
			InstallationMetadata: types.IntegrationInstallationMetadata{
				Attributes: json.RawMessage(`{"installationId":"11","organizationName":"stored-org"}`),
			},
		},
		Input: json.RawMessage(`{"installationId":"77","organizationName":"input-org"}`),
	})
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "77", meta.InstallationID)
	require.Equal(t, "input-org", meta.OrganizationName)
}

// TestResolveInstallationMetadataStoredFallback verifies stored installation metadata is used when input is nil
func TestResolveInstallationMetadataStoredFallback(t *testing.T) {
	t.Parallel()

	meta, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{
		Integration: &generated.Integration{
			InstallationMetadata: types.IntegrationInstallationMetadata{
				Attributes: json.RawMessage(`{"installationId":"11","organizationName":"stored-org"}`),
			},
		},
	})
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "11", meta.InstallationID)
	require.Equal(t, "stored-org", meta.OrganizationName)
	require.Equal(t, "11", meta.InstallationIdentity().ExternalID)
}

// TestResolveInstallationMetadataNoneAvailable verifies ok is false when neither input nor stored metadata carries an installation id
func TestResolveInstallationMetadataNoneAvailable(t *testing.T) {
	t.Parallel()

	_, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{
		Integration: &generated.Integration{},
	})
	require.NoError(t, err)
	require.False(t, ok)
}
