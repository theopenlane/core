package oci

import "errors"

var (
	// ErrCredentialMetadataRequired indicates no credential metadata was provided
	ErrCredentialMetadataRequired = errors.New("oci: credential metadata required")
	// ErrMetadataDecode indicates credential metadata could not be decoded
	ErrMetadataDecode = errors.New("oci: failed to decode credential metadata")
	// ErrConfigurationProviderInvalid indicates the API signing key inputs do not produce a usable request signer
	ErrConfigurationProviderInvalid = errors.New("oci: configuration provider invalid")
	// ErrIdentityClientCreate indicates the OCI Identity client could not be created
	ErrIdentityClientCreate = errors.New("oci: identity client creation failed")
	// ErrCloudGuardClientCreate indicates the OCI Cloud Guard client could not be created
	ErrCloudGuardClientCreate = errors.New("oci: cloud guard client creation failed")
	// ErrTenancyLookupFailed indicates the tenancy read request failed
	ErrTenancyLookupFailed = errors.New("oci: tenancy lookup failed")
	// ErrCompartmentRequired indicates no compartment or tenancy OCID was available to scope collection
	ErrCompartmentRequired = errors.New("oci: compartment or tenancy OCID required")
	// ErrListProblemsFailed indicates the Cloud Guard problem listing request failed
	ErrListProblemsFailed = errors.New("oci: list problems failed")
	// ErrOperationConfigInvalid indicates operation config could not be decoded
	ErrOperationConfigInvalid = errors.New("oci: operation config invalid")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("oci: payload encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("oci: result encode failed")
)
