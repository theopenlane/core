package oci

import (
	"github.com/oracle/oci-go-sdk/v65/cloudguard"
	"github.com/oracle/oci-go-sdk/v65/identity"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

var (
	// definitionID is the stable identifier for the Oracle Cloud Infrastructure integration definition
	definitionID = types.NewDefinitionRef("def_01K0OCI00000000000000000001")
	// installation is the typed installation metadata handle for the Oracle Cloud Infrastructure definition
	installation = types.NewInstallationRef(resolveInstallationMetadata)
	// ociSchema is the credential schema for Oracle Cloud Infrastructure API signing key credentials
	ociSchema, ociCredential = providerkit.CredentialSchema[CredentialSchema]()
	// identityClient is the client ref for the OCI Identity client used by the health check
	identityClient = types.NewClientRef[*identity.IdentityClient]()
	// cloudGuardClient is the client ref for the OCI Cloud Guard client used by findings collection
	cloudGuardClient = types.NewClientRef[*cloudguard.CloudGuardClient]()
	// healthCheckSchema is the operation schema for the Oracle Cloud Infrastructure health check operation
	healthCheckSchema, healthCheckOperation = providerkit.OperationSchema[HealthCheck]()
	// findingsSyncSchema is the operation schema for the Cloud Guard findings collection operation
	findingsSyncSchema, findingsSyncOperation = providerkit.OperationSchema[FindingsSync]()
)

// UserInput holds installation-specific configuration collected from the user
type UserInput struct {
	// FindingsSync includes the configuration for findings from OCI Cloud Guard
	FindingsSync FindingsSync `json:"findingsSync" jsonschema:"title=Cloud Guard Findings Sync"`
}

// FindingsSync holds installation-specific configuration for OCI Cloud Guard problem collection
type FindingsSync struct {
	// Disable is used to disable the findings sync operation from Cloud Guard
	Disable bool `json:"disable,omitempty" jsonschema:"title=Disable,description=Disable the syncing of findings from OCI Cloud Guard"`
	// FilterExpr limits imported records to envelopes matching the CEL expression
	FilterExpr string `json:"filterExpr,omitempty" jsonschema:"title=Filter Expression,description=Optional CEL expression to apply to records before ingesting,example=Example: payload.riskLevel == 'CRITICAL' || payload.riskLevel == 'HIGH'"`
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
