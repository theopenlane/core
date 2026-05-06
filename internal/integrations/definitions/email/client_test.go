package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSender_Resend(t *testing.T) {
	sender, err := buildSender("resend", "test-key")
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestBuildSender_Sendgrid(t *testing.T) {
	sender, err := buildSender("sendgrid", "test-key")
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestBuildSender_Postmark(t *testing.T) {
	sender, err := buildSender("postmark", "test-key")
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestBuildSender_Mock(t *testing.T) {
	sender, err := buildSender(ProviderMock, "")
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestBuildSender_UnsupportedProvider(t *testing.T) {
	sender, err := buildSender("mailgun", "test-key")

	require.ErrorIs(t, err, ErrProviderNotSupported)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "mailgun")
}

func TestBuildSender_EmptyProvider(t *testing.T) {
	sender, err := buildSender("", "test-key")

	require.ErrorIs(t, err, ErrProviderNotSupported)
	assert.Nil(t, sender)
}

func TestEmailScrubber_NotNil(t *testing.T) {
	assert.NotNil(t, EmailScrubber())
}
