package oci

import "errors"

var (
	// ErrConfigurationProviderInvalid indicates the API signing key inputs are invalid
	ErrConfigurationProviderInvalid = errors.New("oci: configuration provider invalid")
	// ErrIdentityClientCreate indicates the OCI Identity client could not be created
	ErrIdentityClientCreate = errors.New("oci: identity client creation failed")
	// ErrCloudGuardClientCreate indicates the OCI Cloud Guard client could not be created
	ErrCloudGuardClientCreate = errors.New("oci: cloud guard client creation failed")
	// ErrTenancyLookupFailed indicates the tenancy read request failed
	ErrTenancyLookupFailed = errors.New("oci: tenancy lookup failed")
	// ErrCompartmentRequired indicates no compartment or tenancy OCID was available
	ErrCompartmentRequired = errors.New("oci: compartment or tenancy OCID required")
	// ErrListProblemsFailed indicates the Cloud Guard problem listing request failed
	ErrListProblemsFailed = errors.New("oci: list problems failed")
	// ErrPayloadEncode indicates a provider payload could not be serialized
	ErrPayloadEncode = errors.New("oci: payload encode failed")
)
