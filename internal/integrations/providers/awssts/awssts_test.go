package awssts

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/config"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestProviderMint verifies AWS STS metadata is normalized and persisted
func TestProviderMint(t *testing.T) {
	schema, err := jsonx.ToRawMessage(map[string]any{"type": "object"})
	require.NoError(t, err)

	spec := config.ProviderSpec{
		Name:              "aws_test",
		AuthType:          types.AuthKindAWSFederation,
		CredentialsSchema: schema,
	}

	builder := Builder(types.ProviderType(spec.Name))
	provider, err := builder.Build(context.Background(), spec)
	require.NoError(t, err)

	subjectPayload := models.CredentialSet{
		ProviderData: json.RawMessage(`{
				"roleArn":"arn:aws:iam::123456789012:role/Openlane",
				"region":"us-east-1",
				"accessKeyId":"AKIA123",
				"secretAccessKey":"secret"
			}`),
	}

	payload, err := provider.Mint(context.Background(), types.CredentialMintRequest{
		Provider:   types.ProviderType(spec.Name),
		Credential: subjectPayload,
	})
	require.NoError(t, err)

	require.Equal(t, "AKIA123", payload.AccessKeyID)
	require.Equal(t, "secret", payload.SecretAccessKey)
	require.JSONEq(t, `{"roleArn":"arn:aws:iam::123456789012:role/Openlane","region":"us-east-1","homeRegion":"us-east-1","accountScope":"all"}`, string(payload.ProviderData))
}
