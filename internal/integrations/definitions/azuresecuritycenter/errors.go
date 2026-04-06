package azuresecuritycenter

import "errors"

var (
	// ErrTenantIDMissing indicates the Azure tenant ID is missing
	ErrTenantIDMissing = errors.New("azuresecuritycenter: tenant ID required")
	// ErrClientIDMissing indicates the Azure client ID is missing
	ErrClientIDMissing = errors.New("azuresecuritycenter: client ID required")
	// ErrClientSecretMissing indicates the Azure client secret is missing
	ErrClientSecretMissing = errors.New("azuresecuritycenter: client secret required")
	// ErrSubscriptionIDMissing indicates the Azure subscription ID is missing
	ErrSubscriptionIDMissing = errors.New("azuresecuritycenter: subscription ID required")
	// ErrCredentialInvalid indicates credential metadata could not be decoded
	ErrCredentialInvalid = errors.New("azuresecuritycenter: credential invalid")
	// ErrCredentialBuildFailed indicates the Azure credential could not be constructed
	ErrCredentialBuildFailed = errors.New("azuresecuritycenter: credential build failed")
	// ErrAssessmentsClientBuildFailed indicates an armsecurity assessments client could not be constructed
	ErrAssessmentsClientBuildFailed = errors.New("azuresecuritycenter: assessments client build failed")
	// ErrAssessmentFetchFailed indicates the assessments list request failed
	ErrAssessmentFetchFailed = errors.New("azuresecuritycenter: assessment fetch failed")
	// ErrSubAssessmentFetchFailed indicates the sub-assessments list request failed
	ErrSubAssessmentFetchFailed = errors.New("azuresecuritycenter: sub-assessment fetch failed")
	// ErrIngestPayloadEncode indicates an ingest envelope payload could not be serialized
	ErrIngestPayloadEncode = errors.New("azuresecuritycenter: ingest payload encode failed")
	// ErrResultEncode indicates an operation result could not be serialized
	ErrResultEncode = errors.New("azuresecuritycenter: result encode failed")
)
