package azuresecuritycenter

import "errors"

var (
	// ErrTenantIDMissing indicates the Azure tenant ID is missing
	ErrTenantIDMissing = errors.New("azuresecuritycenter: tenant ID required")
	// ErrNoSubscriptions indicates no subscriptions are accessible, usually because no RBAC role was assigned to the app
	ErrNoSubscriptions = errors.New("azuresecuritycenter: no accessible subscriptions; assign the Security Reader role to the Openlane app on at least one subscription")
	// ErrNoTenantSubscriptionsDelegated indicates no subscription in the tenant is visible through any Lighthouse delegation to the managing tenant
	ErrNoTenantSubscriptionsDelegated = errors.New("azuresecuritycenter: no subscriptions in this tenant are delegated to Openlane yet; check the tenant ID, or wait up to 15 minutes after deploying the Lighthouse template")
	// ErrNoDelegatedSubscriptions indicates subscriptions are delegated but none carry the installation's connection code
	ErrNoDelegatedSubscriptions = errors.New("azuresecuritycenter: no subscriptions delegated with this connection code; the offer name in the Service providers page must end with the connection code saved in Openlane")
	// ErrDelegationLookupFailed indicates the Lighthouse delegations of a subscription could not be read
	ErrDelegationLookupFailed = errors.New("azuresecuritycenter: lighthouse delegation lookup failed")
	// ErrLighthouseNotConfigured indicates the operator has not configured the Lighthouse managing tenant
	ErrLighthouseNotConfigured = errors.New("azuresecuritycenter: lighthouse managing tenant not configured")
	// ErrDisconnectDetailsEncode indicates the disconnect result details could not be serialized
	ErrDisconnectDetailsEncode = errors.New("azuresecuritycenter: disconnect details encode failed")
	// ErrConnectionCodeMissing indicates the Lighthouse connection code is missing
	ErrConnectionCodeMissing = errors.New("azuresecuritycenter: connection code required")
	// ErrSubscriptionNotAccessible indicates a requested subscription is not owned by the connected tenant or not accessible to the app
	ErrSubscriptionNotAccessible = errors.New("azuresecuritycenter: requested subscription is not accessible in the connected tenant")
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
