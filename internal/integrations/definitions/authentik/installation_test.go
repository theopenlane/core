package authentik

import (
	"testing"

	"gotest.tools/v3/assert"
)

// TestInstallationIdentity verifies the default brand uuid is the external id and the brand name is display only
func TestInstallationIdentity(t *testing.T) {
	identity := InstallationMetadata{
		Brand:   "authentik",
		BrandID: "6b1a0c8e-2f4d-4c8a-9a3e-1d2b3c4d5e6f",
		Host:    "auth.example.com",
		BaseURL: "https://auth.example.com",
	}.InstallationIdentity()

	assert.Equal(t, identity.ExternalID, "6b1a0c8e-2f4d-4c8a-9a3e-1d2b3c4d5e6f")
	assert.Equal(t, identity.ExternalName, "authentik")
}
