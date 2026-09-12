package zitadel

import (
	"testing"

	"gotest.tools/v3/assert"
)

// TestInstallationIdentity verifies the instance id is the external id and the domain is display only
func TestInstallationIdentity(t *testing.T) {
	identity := InstallationMetadata{
		Domain:     "my-instance.zitadel.cloud",
		InstanceID: "318612938736893956",
	}.InstallationIdentity()

	assert.Equal(t, identity.ExternalID, "318612938736893956")
	assert.Equal(t, identity.ExternalName, "my-instance.zitadel.cloud")
}
