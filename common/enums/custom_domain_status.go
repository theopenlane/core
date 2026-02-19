package enums

import "io"

// DNSVerificationStatus is a custom type for DNS verification status.
type DNSVerificationStatus string

var (
	DNSVerificationStatusActive             DNSVerificationStatus = "ACTIVE"
	DNSVerificationStatusPending            DNSVerificationStatus = "PENDING"
	DNSVerificationStatusActiveRedeploying  DNSVerificationStatus = "ACTIVE_REDEPLOYING"
	DNSVerificationStatusMoved              DNSVerificationStatus = "MOVED"
	DNSVerificationStatusPendingDeletion    DNSVerificationStatus = "PENDING_DELETION"
	DNSVerificationStatusDeleted            DNSVerificationStatus = "DELETED"
	DNSVerificationStatusPendingBlocked     DNSVerificationStatus = "PENDING_BLOCKED"
	DNSVerificationStatusPendingMigration   DNSVerificationStatus = "PENDING_MIGRATION"
	DNSVerificationStatusPendingProvisioned DNSVerificationStatus = "PENDING_PROVISIONED"
	DNSVerificationStatusTestPending        DNSVerificationStatus = "TEST_PENDING"
	DNSVerificationStatusTestActive         DNSVerificationStatus = "TEST_ACTIVE"
	DNSVerificationStatusTestActiveApex     DNSVerificationStatus = "TEST_ACTIVE_APEX"
	DNSVerificationStatusTestBlocked        DNSVerificationStatus = "TEST_BLOCKED"
	DNSVerificationStatusTestFailed         DNSVerificationStatus = "TEST_FAILED"
	DNSVerificationStatusProvisioned        DNSVerificationStatus = "PROVISIONED"
	DNSVerificationStatusBlocked            DNSVerificationStatus = "BLOCKED"
	DNSVerificationStatusInvalid            DNSVerificationStatus = "INVALID"
)

var dnsVerificationStatusValues = []DNSVerificationStatus{
	DNSVerificationStatusActive, DNSVerificationStatusPending, DNSVerificationStatusActiveRedeploying,
	DNSVerificationStatusMoved, DNSVerificationStatusPendingDeletion, DNSVerificationStatusDeleted,
	DNSVerificationStatusPendingBlocked, DNSVerificationStatusPendingMigration, DNSVerificationStatusPendingProvisioned,
	DNSVerificationStatusTestPending, DNSVerificationStatusTestActive, DNSVerificationStatusTestActiveApex,
	DNSVerificationStatusTestBlocked, DNSVerificationStatusTestFailed, DNSVerificationStatusProvisioned,
	DNSVerificationStatusBlocked,
}

// Values returns all valid DNSVerificationStatus values as strings.
func (DNSVerificationStatus) Values() []string { return stringValues(dnsVerificationStatusValues) }

// String returns the DNSVerificationStatus as a string
func (r DNSVerificationStatus) String() string { return string(r) }

// ToDNSVerificationStatus returns the DNS verification status enum based on string input
func ToDNSVerificationStatus(r string) *DNSVerificationStatus {
	return parse(r, dnsVerificationStatusValues, &DNSVerificationStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r DNSVerificationStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *DNSVerificationStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// SSLVerificationStatus is a custom type for SSL verification status.
type SSLVerificationStatus string

var (
	SSLVerificationStatusInitializing         SSLVerificationStatus = "INITIALIZING"
	SSLVerificationStatusPendingValidation    SSLVerificationStatus = "PENDING_VALIDATION"
	SSLVerificationStatusDeleted              SSLVerificationStatus = "DELETED"
	SSLVerificationStatusPendingIssuance      SSLVerificationStatus = "PENDING_ISSUANCE"
	SSLVerificationStatusPendingDeployment    SSLVerificationStatus = "PENDING_DEPLOYMENT"
	SSLVerificationStatusPendingDeletion      SSLVerificationStatus = "PENDING_DELETION"
	SSLVerificationStatusPendingExpiration    SSLVerificationStatus = "PENDING_EXPIRATION"
	SSLVerificationStatusExpired              SSLVerificationStatus = "EXPIRED"
	SSLVerificationStatusActive               SSLVerificationStatus = "ACTIVE"
	SSLVerificationStatusInitializingTimedOut SSLVerificationStatus = "INITIALIZING_TIMED_OUT"
	SSLVerificationStatusValidationTimedOut   SSLVerificationStatus = "VALIDATION_TIMED_OUT"
	SSLVerificationStatusIssuanceTimedOut     SSLVerificationStatus = "ISSUANCE_TIMED_OUT"
	SSLVerificationStatusDeploymentTimedOut   SSLVerificationStatus = "DEPLOYMENT_TIMED_OUT"
	SSLVerificationStatusDeletionTimedOut     SSLVerificationStatus = "DELETION_TIMED_OUT"
	SSLVerificationStatusPendingCleanup       SSLVerificationStatus = "PENDING_CLEANUP"
	SSLVerificationStatusStagingDeployment    SSLVerificationStatus = "STAGING_DEPLOYMENT"
	SSLVerificationStatusStagingActive        SSLVerificationStatus = "STAGING_ACTIVE"
	SSLVerificationStatusDeactivating         SSLVerificationStatus = "DEACTIVATING"
	SSLVerificationStatusInactive             SSLVerificationStatus = "INACTIVE"
	SSLVerificationStatusBackupIssued         SSLVerificationStatus = "BACKUP_ISSUED"
	SSLVerificationStatusHoldingDeployment    SSLVerificationStatus = "HOLDING_DEPLOYMENT"
	SSLVerificationStatusInvalid              SSLVerificationStatus = "INVALID"
)

var sslVerificationStatusValues = []SSLVerificationStatus{
	SSLVerificationStatusInitializing, SSLVerificationStatusPendingValidation, SSLVerificationStatusDeleted,
	SSLVerificationStatusPendingIssuance, SSLVerificationStatusPendingDeployment, SSLVerificationStatusPendingDeletion,
	SSLVerificationStatusPendingExpiration, SSLVerificationStatusExpired, SSLVerificationStatusActive,
	SSLVerificationStatusInitializingTimedOut, SSLVerificationStatusValidationTimedOut, SSLVerificationStatusIssuanceTimedOut,
	SSLVerificationStatusDeploymentTimedOut, SSLVerificationStatusDeletionTimedOut, SSLVerificationStatusPendingCleanup,
	SSLVerificationStatusStagingDeployment, SSLVerificationStatusStagingActive, SSLVerificationStatusDeactivating,
	SSLVerificationStatusInactive, SSLVerificationStatusBackupIssued, SSLVerificationStatusHoldingDeployment,
}

// Values returns all valid SSLVerificationStatus values as strings.
func (SSLVerificationStatus) Values() []string { return stringValues(sslVerificationStatusValues) }

// String returns the SSLVerificationStatus as a string
func (r SSLVerificationStatus) String() string { return string(r) }

// ToSSLVerificationStatus returns the SSL verification status enum based on string input
func ToSSLVerificationStatus(r string) *SSLVerificationStatus {
	return parse(r, sslVerificationStatusValues, &SSLVerificationStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r SSLVerificationStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *SSLVerificationStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
