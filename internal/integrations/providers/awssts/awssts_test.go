package awssts

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

// TestProviderMint handles test provider mint
func TestProviderMint(t *testing.T) {
	spec := config.ProviderSpec{
		Name:              "aws_test",
		AuthType:          types.AuthKindAWSFederation,
		CredentialsSchema: map[string]any{"type": "object"},
	}

	builder := Builder(types.ProviderType(spec.Name))
	provider, err := builder.Build(context.Background(), spec)
	require.NoError(t, err)

	subjectPayload, err := types.NewCredentialBuilder(types.ProviderType(spec.Name)).
		With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(models.CredentialSet{
				ProviderData: map[string]any{
					"roleArn":         "arn:aws:iam::123456789012:role/Openlane",
					"region":          "us-east-1",
					"accessKeyId":     "AKIA123",
					"secretAccessKey": "secret",
				},
			}),
		).Build()
	require.NoError(t, err)

	payload, err := provider.Mint(context.Background(), types.CredentialSubject{
		Provider:   types.ProviderType(spec.Name),
		Credential: subjectPayload,
	})
	require.NoError(t, err)

	require.Equal(t, types.CredentialKindMetadata, payload.Kind)
	require.Equal(t, "AKIA123", payload.Data.AccessKeyID)
	require.Equal(t, "secret", payload.Data.SecretAccessKey)
	require.Equal(t, "us-east-1", payload.Data.ProviderData["region"])
}
