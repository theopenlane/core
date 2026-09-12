package tailscale

import (
	"testing"

	"gotest.tools/v3/assert"
)

// TestInstallationIdentity verifies the tailnet name is the external id rather than the rotatable OAuth client id
func TestInstallationIdentity(t *testing.T) {
	identity := InstallationMetadata{
		ClientID: "k123456CNTRL",
		Tailnet:  "example.com",
	}.InstallationIdentity()

	assert.Equal(t, identity.ExternalID, "example.com")
	assert.Equal(t, identity.ExternalName, "example.com")
}
