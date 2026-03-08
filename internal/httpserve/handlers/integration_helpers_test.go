package handlers

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseStatePayload validates state parsing for a well-formed payload.
func TestParseStatePayload(t *testing.T) {
	orgID := "org_123"
	provider := "githubapp"
	randomBytes := []byte("randombytes")

	stateData := buildStatePayload(orgID, provider, randomBytes)
	state := base64.URLEncoding.EncodeToString([]byte(stateData))

	gotOrg, gotProvider, err := parseStatePayload(state)
	assert.NoError(t, err)
	assert.Equal(t, orgID, gotOrg)
	assert.Equal(t, provider, gotProvider)
}

// TestParseStatePayload_Invalid validates state parsing failures for invalid inputs.
func TestParseStatePayload_Invalid(t *testing.T) {
	cases := []string{
		"",
		"not-base64",
		base64.URLEncoding.EncodeToString([]byte("missing:parts")),
		base64.URLEncoding.EncodeToString([]byte("org:provider")),
		base64.URLEncoding.EncodeToString([]byte("org:provider:###")),
	}

	for _, tc := range cases {
		_, _, err := parseStatePayload(tc)
		assert.Error(t, err)
	}
}

func TestParseStatePayload_RawURLRandomPart(t *testing.T) {
	orgID := "org_123"
	provider := "githubapp"
	randomPart := base64.RawURLEncoding.EncodeToString([]byte("randombytes"))
	stateData := orgID + ":" + provider + ":" + randomPart
	state := base64.RawURLEncoding.EncodeToString([]byte(stateData))

	gotOrg, gotProvider, err := parseStatePayload(state)
	assert.NoError(t, err)
	assert.Equal(t, orgID, gotOrg)
	assert.Equal(t, provider, gotProvider)
}

func TestGenerateOAuthState_ParseRoundTrip(t *testing.T) {
	h := &Handler{}

	state, err := h.generateOAuthState("org_123", "githubapp")
	assert.NoError(t, err)
	assert.NotEmpty(t, state)

	gotOrg, gotProvider, err := parseStatePayload(state)
	assert.NoError(t, err)
	assert.Equal(t, "org_123", gotOrg)
	assert.Equal(t, "githubapp", gotProvider)
}
