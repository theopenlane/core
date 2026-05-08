package email

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBuildSender_Resend(t *testing.T) {
	sender, err := buildSender("resend", "test-key", false, "")
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestBuildSender_ResendDevMode(t *testing.T) {
	sender, err := buildSender("resend", "", true, t.TempDir())
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestBuildSender_Mock(t *testing.T) {
	sender, err := buildSender(ProviderMock, "", false, "")
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestBuildSender_UnsupportedProvider(t *testing.T) {
	sender, err := buildSender("unsupported", "test-key", false, "")

	require.ErrorIs(t, err, ErrProviderNotSupported)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestBuildSender_EmptyProvider(t *testing.T) {
	sender, err := buildSender("", "test-key", false, "")

	require.ErrorIs(t, err, ErrProviderNotSupported)
	assert.Nil(t, sender)
}

func TestEmailScrubber_NotNil(t *testing.T) {
	assert.NotNil(t, EmailScrubber())
}

func TestDevModeRuntimeClient_SendsWelcomeEmail(t *testing.T) {
	t.Parallel()

	testDir := t.TempDir()

	cfg := RuntimeEmailConfig{
		Provider:    "resend",
		FromEmail:   "test@example.com",
		CompanyName: "TestCo",
		ProductURL:  "https://app.example.com",
		DocsURL:     "https://docs.example.com",
		TestDir:     testDir,
	}

	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	built, err := runtimeClientBuilder(true)(context.Background(), raw)
	require.NoError(t, err)

	client, ok := built.(*Client)
	require.True(t, ok)
	require.NotNil(t, client.Sender)

	welcome := testDispatcher[WelcomeRequest](t, "WelcomeRequest")

	err = welcome.SendByKey(context.Background(), types.OperationRequest{}, client, mustMarshal(t, WelcomeRequest{
		RecipientInfo: RecipientInfo{
			Email:     "dev@example.com",
			FirstName: "Dev",
		},
	}))
	require.NoError(t, err)

	recipientDir := filepath.Join(testDir, "dev@example.com")
	entries, err := os.ReadDir(recipientDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.True(t, filepath.Ext(entries[0].Name()) == ".mim")
}

func TestDevModeRuntimeClient_ProvisionedIgnoresDevMode(t *testing.T) {
	t.Parallel()

	cfg := RuntimeEmailConfig{
		APIKey:      "real-key",
		Provider:    "resend",
		FromEmail:   "prod@example.com",
		CompanyName: "ProdCo",
		TestDir:     t.TempDir(),
	}

	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	built, err := runtimeClientBuilder(false)(context.Background(), raw)
	require.NoError(t, err)

	client, ok := built.(*Client)
	require.True(t, ok)
	require.NotNil(t, client.Sender)
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)

	return data
}
