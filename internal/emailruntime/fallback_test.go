package emailruntime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
)

func TestSystemFallbackTemplates_AllKeysPresent(t *testing.T) {
	keys := []string{
		TemplateKeyVerifyEmail,
		TemplateKeyWelcome,
		TemplateKeyInvite,
		TemplateKeyInviteJoined,
		TemplateKeyPasswordResetRequest,
		TemplateKeyPasswordResetSuccess,
		TemplateKeySubscribe,
		TemplateKeyVerifyBilling,
		TemplateKeyTrustCenterNDARequest,
		TemplateKeyTrustCenterNDASigned,
		TemplateKeyTrustCenterAuth,
		TemplateKeyQuestionnaireAuth,
		TemplateKeyBillingEmailChanged,
	}

	for _, key := range keys {
		_, ok := systemFallbackTemplates[key]
		assert.True(t, ok, "missing fallback template for key %q", key)
	}
}

func TestSystemFallbackTemplates_NoEmptyContent(t *testing.T) {
	for key, fb := range systemFallbackTemplates {
		assert.NotEmpty(t, fb.SubjectTemplate, "empty SubjectTemplate for key %q", key)
		assert.NotEmpty(t, fb.BodyTemplate, "empty BodyTemplate for key %q", key)
	}
}

func TestFallbackTemplate_ToNotificationTemplate(t *testing.T) {
	const key = TemplateKeyVerifyEmail

	fb := systemFallbackTemplates[key]
	record := fb.toNotificationTemplate(key)

	require.NotNil(t, record)
	assert.Equal(t, key, record.Key)
	assert.Equal(t, key, record.Name)
	assert.Equal(t, enums.ChannelEmail, record.Channel)
	assert.Equal(t, enums.NotificationTemplateFormatHTML, record.Format)
	assert.Equal(t, fb.SubjectTemplate, record.SubjectTemplate)
	assert.Equal(t, fb.BodyTemplate, record.BodyTemplate)
	assert.True(t, record.Active)
	assert.True(t, record.SystemOwned)
	assert.Empty(t, record.EmailTemplateID)
}

func TestFallbackTemplate_UnknownKey(t *testing.T) {
	_, ok := systemFallbackTemplates["unknown_key_xyzzy"]
	assert.False(t, ok)
}
