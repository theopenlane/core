package azuresecuritycenter

import "errors"

var (
	// ErrTenantIDMissing indicates the Azure tenant ID is missing
	ErrTenantIDMissing = errors.New("azuresecuritycenter: tenant ID required")
	// ErrNoSubscriptions indicates no subscriptions are accessible, usually because no RBAC role was assigned to the app
	ErrNoSubscriptions = errors.New("azuresecuritycenter: no accessible subscriptions; assign the Security Reader role to the Openlane app on at least one subscription")
	// ErrCredentialBuildFailed indicates the Azure credential could not be constructed
	ErrCredentialBuildFailed = errors.New("azuresecuritycenter: credential build failed")
	// ErrAssessmentFetchFailed indicates the assessments query failed
	ErrAssessmentFetchFailed = errors.New("azuresecuritycenter: assessment fetch failed")
	// ErrResourceGraphQueryFailed indicates a Resource Graph query failed
	ErrResourceGraphQueryFailed = errors.New("azuresecuritycenter: resource graph query failed")
	// ErrAssessmentDecode indicates an assessment row could not be decoded
	ErrAssessmentDecode = errors.New("azuresecuritycenter: assessment decode failed")
	// ErrIngestPayloadEncode indicates an ingest envelope payload could not be serialized
	ErrIngestPayloadEncode = errors.New("azuresecuritycenter: ingest payload encode failed")
)
