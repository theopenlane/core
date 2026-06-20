package email

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/integrations/templatekit"
)

// TestUnsubscribeTokenFromMetadata verifies the per-recipient unsubscribe token is read from
// campaign target metadata only when present and string-typed
func TestUnsubscribeTokenFromMetadata(t *testing.T) {
	assert.Equal(t, "tok_abc", unsubscribeTokenFromMetadata(map[string]any{MetadataUnsubscribeTokenKey: "tok_abc"}))
	assert.Empty(t, unsubscribeTokenFromMetadata(map[string]any{}))
	assert.Empty(t, unsubscribeTokenFromMetadata(nil))
	assert.Empty(t, unsubscribeTokenFromMetadata(map[string]any{MetadataUnsubscribeTokenKey: 123}))
}

// TestUnsubscribeTokenInterpolatedIntoPayload verifies the per-recipient unsubscribe token is an
// available template variable, so a template's unsubscribeURL default resolves to a unique link
// for each recipient at render time
func TestUnsubscribeTokenInterpolatedIntoPayload(t *testing.T) {
	client := &Client{Config: *MockRuntimeConfig()}

	defaults := map[string]any{
		"unsubscribeURL": "https://tc.example.com/unsubscribe?token={{ .unsubscribeToken }}",
	}

	payload, err := templatekit.BuildDispatchPayload(defaults, RecipientInfo{
		Email:            "person@example.com",
		UnsubscribeToken: "tok_abc123",
	})
	assert.NoError(t, err)

	resolved, err := interpolatePayload(client, payload)
	assert.NoError(t, err)

	var req BrandedMessageRequest
	assert.NoError(t, json.Unmarshal(resolved, &req))
	assert.Equal(t, "https://tc.example.com/unsubscribe?token=tok_abc123", req.UnsubscribeURL)
}

func TestSplitFullName_FirstAndLast(t *testing.T) {
	first, last := splitFullName("Alice Smith")

	assert.Equal(t, "Alice", first)
	assert.Equal(t, "Smith", last)
}

func TestSplitFullName_FirstOnly(t *testing.T) {
	first, last := splitFullName("Alice")

	assert.Equal(t, "Alice", first)
	assert.Equal(t, "", last)
}

func TestSplitFullName_Empty(t *testing.T) {
	first, last := splitFullName("")

	assert.Equal(t, "", first)
	assert.Equal(t, "", last)
}

func TestSplitFullName_Whitespace(t *testing.T) {
	first, last := splitFullName("   ")

	assert.Equal(t, "", first)
	assert.Equal(t, "", last)
}

func TestSplitFullName_LeadingTrailingSpaces(t *testing.T) {
	first, last := splitFullName("  Alice Smith  ")

	assert.Equal(t, "Alice", first)
	assert.Equal(t, "Smith", last)
}

func TestSplitFullName_MultipleNames(t *testing.T) {
	// Only splits on the first space; "Mary Smith" stays in last
	first, last := splitFullName("Jane Mary Smith")

	assert.Equal(t, "Jane", first)
	assert.Equal(t, "Mary Smith", last)
}

func TestSplitFullName_Unicode(t *testing.T) {
	first, last := splitFullName("José García")

	assert.Equal(t, "José", first)
	assert.Equal(t, "García", last)
}
