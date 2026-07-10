package runtime

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
	emaildef "github.com/theopenlane/core/internal/integrations/definitions/email"
)

// TestSubprocessorChange verifies a changed subprocessor join row is classified relative to the last
// notification floor: soft-deleted is a removal, created after the floor is an addition, else an update
func TestSubprocessorChange(t *testing.T) {
	t.Parallel()

	floor := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name     string
		row      *ent.TrustCenterSubprocessor
		expected string
	}{
		{
			name:     "soft-deleted row is removed",
			row:      &ent.TrustCenterSubprocessor{CreatedAt: floor.Add(-time.Hour), DeletedAt: time.Now()},
			expected: "Removed",
		},
		{
			name:     "created after floor is added",
			row:      &ent.TrustCenterSubprocessor{CreatedAt: time.Now()},
			expected: "Added",
		},
		{
			name:     "created before floor is updated",
			row:      &ent.TrustCenterSubprocessor{CreatedAt: floor.Add(-time.Hour)},
			expected: "Updated",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, subprocessorChange(tc.row, floor))
		})
	}
}

// TestSubprocessorNotificationRequest verifies the notification body maps the trust center branding from
// the setting, carries the changed vendor entries, and uses the tokenized unsubscribe link template
func TestSubprocessorNotificationRequest(t *testing.T) {
	t.Parallel()

	setting := &ent.TrustCenterSetting{
		CompanyName:              "SecureCorp",
		LogoRemoteURL:            lo.ToPtr("https://securecorp.example.com/logo.png"),
		PrimaryColor:             "#0f3d3a",
		AccentColor:              "#3fc2b4",
		BackgroundColor:          "#e8eaed",
		SecondaryBackgroundColor: "#ffffff",
		ForegroundColor:          "#14171e",
	}

	entries := []emaildef.SubprocessorEntry{
		{Name: "Amazon Web Services", Change: "Added"},
	}

	req := subprocessorNotificationRequest(setting, "trust.securecorp.com", "securecorp", entries)

	assert.Equal(t, "SecureCorp", req.CompanyName)
	assert.Equal(t, "https://securecorp.example.com/logo.png", req.LogoURL)
	assert.Equal(t, "#0f3d3a", req.PrimaryColor)
	assert.Equal(t, "#3fc2b4", req.ButtonColor)
	assert.Equal(t, "#e8eaed", req.BodyBackgroundColor)
	assert.Equal(t, "#ffffff", req.CardBackgroundColor)
	assert.Equal(t, "#14171e", req.TextColor)

	assert.Equal(t, entries, req.Subprocessors)
	assert.Equal(t, subprocessorNotificationSubject, req.Subject)
	assert.Equal(t, subprocessorNotificationTitle, req.Title)
	assert.Equal(t, subprocessorNotificationButtonText, req.ButtonText)
	assert.NotEmpty(t, req.ButtonLink)
	// the per-recipient unsubscribe link is resolved per subscriber at send time, not on the base request
	assert.Empty(t, req.UnsubscribeURL)
}

// TestTrustCenterNotificationContent verifies a post notification carries the post subject, title, and
// body along with the tokenized unsubscribe link and a trust center button
func TestTrustCenterNotificationContent(t *testing.T) {
	t.Parallel()

	content := trustCenterNotificationContent(
		"June trust center update",
		"June trust center update",
		[]string{"We published a new update."},
		"trust.securecorp.com",
		"securecorp",
	)

	assert.Equal(t, "June trust center update", content["subject"])
	assert.Equal(t, "June trust center update", content["title"])
	assert.Equal(t, []string{"We published a new update."}, content["intros"])
	assert.Equal(t, "View trust center", content["buttonText"])
	assert.NotEmpty(t, content["buttonLink"])

	unsubscribeURL, ok := content["unsubscribeURL"].(string)
	assert.True(t, ok)
	assert.Contains(t, unsubscribeURL, "{{ .unsubscribeToken }}")
}
