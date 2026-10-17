package microsoft

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
)

const (
	// armModuleName identifies this client in ARM telemetry headers
	armModuleName = "github.com/theopenlane/core/microsoft"
	// armModuleVersion is the module version reported in ARM telemetry headers
	armModuleVersion = "v1.0.0"
)

// ARMClient calls Azure Resource Manager APIs for one tenant credential
type ARMClient struct {
	// client is the ARM pipeline and endpoint shared by every request
	client *arm.Client
}

// NewARMClient builds an ARM client for the credential, honoring the app config's endpoint overrides
func NewARMClient(cfg AppConfig, cred azcore.TokenCredential) (*ARMClient, error) {
	client, err := arm.NewClient(armModuleName, armModuleVersion, cred, cfg.ARMClientOptions())
	if err != nil {
		return nil, err
	}

	return &ARMClient{client: client}, nil
}
