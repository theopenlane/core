package oci

import (
	"github.com/oracle/oci-go-sdk/v65/cloudguard"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Oracle Cloud Infrastructure integration definition
	definitionID = types.NewDefinitionRef("def_01K0OCI00000000000000000001")
	// apiKey is the connection for OCI API signing key credentials
	apiKey = types.ConnectionOf[CredentialSchema]()
	// installation is the typed installation metadata handle for the definition
	installation = types.InstallationOf[InstallationMetadata]()
)

// Client holds the OCI SDK clients with the tenancy scope for one installation
type Client struct {
	// Identity is the OCI Identity client used by verification
	Identity *identity.IdentityClient
	// CloudGuard is the OCI Cloud Guard client used by findings collection
	CloudGuard *cloudguard.CloudGuardClient
	// TenancyOCID is the OCID of the tenancy the credentials belong to
	TenancyOCID string
	// CompartmentOCID is the compartment collection is rooted at when one was configured
	CompartmentOCID string
	// Region is the OCI region API calls are issued against
	Region string
}

// Compartment returns the compartment collection is rooted at, defaulting to the tenancy
func (c Client) Compartment() string {
	return lo.CoalesceOrEmpty(c.CompartmentOCID, c.TenancyOCID)
}

// FindingsSync holds installation-specific configuration for OCI Cloud Guard problem collection
type FindingsSync struct {
	types.OperationSettings
	// SkipProblemDetails collects only the list response and skips the per-problem detail lookup
	SkipProblemDetails bool `json:"skipProblemDetails,omitempty" jsonschema:"title=Skip Problem Details,description=Skip the per-problem detail lookup. Far fewer API calls on large tenancies, but findings arrive without a description or recommendation"`
}

// CredentialSchema holds the OCI API signing key credentials for one installation
type CredentialSchema struct {
	// TenancyOCID is the OCID of the tenancy the credentials belong to
	TenancyOCID string `json:"tenancyOcid" jsonschema:"required,title=Tenancy OCID,description=OCID of the tenancy Openlane should read from (e.g. ocid1.tenancy.oc1..aaaa)"`
	// UserOCID is the OCID of the user the API signing key is registered against
	UserOCID string `json:"userOcid" jsonschema:"required,title=User OCID,description=OCID of the user the API signing key is registered against (e.g. ocid1.user.oc1..aaaa)"`
	// Fingerprint is the fingerprint of the uploaded API signing key
	Fingerprint string `json:"fingerprint" jsonschema:"required,title=API Key Fingerprint,description=Fingerprint shown in the OCI console for the uploaded public key"`
	// PrivateKey is the PEM encoded API signing private key
	PrivateKey string `json:"privateKey" jsonschema:"required,title=API Private Key,secret=true,description=PEM encoded private key matching the uploaded public key"`
	// PrivateKeyPassphrase is the passphrase protecting the private key when it is encrypted
	PrivateKeyPassphrase string `json:"privateKeyPassphrase,omitempty" jsonschema:"title=Private Key Passphrase,secret=true,description=Only required when the private key is encrypted"`
	// Region is the OCI region API calls are issued against
	Region string `json:"region" jsonschema:"required,title=Region,description=OCI region identifier used for API calls (e.g. us-ashburn-1)"`
	// CompartmentOCID scopes collection to one compartment and everything beneath it
	CompartmentOCID string `json:"compartmentOcid,omitempty" jsonschema:"title=Compartment OCID,description=Compartment to collect from including its subcompartments. Defaults to the tenancy root compartment"`
}

// InstallationMetadata holds the stable OCI tenancy identity for one installation
type InstallationMetadata struct {
	// TenancyOCID is the OCID of the tenancy collection is scoped to
	TenancyOCID string `json:"tenancyOcid,omitempty" jsonschema:"title=Tenancy OCID"`
	// CompartmentOCID is the compartment collection is rooted at when one was configured
	CompartmentOCID string `json:"compartmentOcid,omitempty" jsonschema:"title=Compartment OCID"`
	// Region is the OCI region API calls are issued against
	Region string `json:"region,omitempty" jsonschema:"title=Region"`
}

// InstallationIdentity implements types.InstallationIdentifiable
func (m InstallationMetadata) InstallationIdentity() types.IntegrationInstallationIdentity {
	return types.IntegrationInstallationIdentity{
		ExternalID: m.TenancyOCID,
	}
}
