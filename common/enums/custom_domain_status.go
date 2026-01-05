package enums

import (
	"fmt"
	"io"
	"strings"
)

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

func (DNSVerificationStatus) Values() (kinds []string) {
	v := []DNSVerificationStatus{
		DNSVerificationStatusActive,
		DNSVerificationStatusPending,
		DNSVerificationStatusActiveRedeploying,
		DNSVerificationStatusMoved,
		DNSVerificationStatusPendingDeletion,
		DNSVerificationStatusDeleted,
		DNSVerificationStatusPendingBlocked,
		DNSVerificationStatusPendingMigration,
		DNSVerificationStatusPendingProvisioned,
		DNSVerificationStatusTestPending,
		DNSVerificationStatusTestActive,
		DNSVerificationStatusTestActiveApex,
		DNSVerificationStatusTestBlocked,
		DNSVerificationStatusTestFailed,
		DNSVerificationStatusProvisioned,
		DNSVerificationStatusBlocked,
	}
	for _, s := range v {
		kinds = append(kinds, string(s))
	}

	return kinds
}

// String returns the DNSVerificationStatus as a string
func (r DNSVerificationStatus) String() string {
	return string(r)
}

// ToDNSVerificationStatus returns the user status enum based on string input
func ToDNSVerificationStatus(r string) *DNSVerificationStatus {
	rUpper := strings.ToUpper(r)
	switch rUpper {
	case string(DNSVerificationStatusActive):
		return &DNSVerificationStatusActive
	case string(DNSVerificationStatusPending):
		return &DNSVerificationStatusPending
	case string(DNSVerificationStatusActiveRedeploying):
		return &DNSVerificationStatusActiveRedeploying
	case string(DNSVerificationStatusMoved):
		return &DNSVerificationStatusMoved
	case string(DNSVerificationStatusPendingDeletion):
		return &DNSVerificationStatusPendingDeletion
	case string(DNSVerificationStatusDeleted):
		return &DNSVerificationStatusDeleted
	case string(DNSVerificationStatusPendingBlocked):
		return &DNSVerificationStatusPendingBlocked
	case string(DNSVerificationStatusPendingMigration):
		return &DNSVerificationStatusPendingMigration
	case string(DNSVerificationStatusPendingProvisioned):
		return &DNSVerificationStatusPendingProvisioned
	case string(DNSVerificationStatusTestPending):
		return &DNSVerificationStatusTestPending
	case string(DNSVerificationStatusTestActive):
		return &DNSVerificationStatusTestActive
	case string(DNSVerificationStatusTestActiveApex):
		return &DNSVerificationStatusTestActiveApex
	case string(DNSVerificationStatusTestBlocked):
		return &DNSVerificationStatusTestBlocked
	case string(DNSVerificationStatusTestFailed):
		return &DNSVerificationStatusTestFailed
	case string(DNSVerificationStatusProvisioned):
		return &DNSVerificationStatusProvisioned
	case string(DNSVerificationStatusBlocked):
		return &DNSVerificationStatusBlocked
	default:
		return &DNSVerificationStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r DNSVerificationStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *DNSVerificationStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for DNSVerificationStatus, got: %T", v) //nolint:err113
	}

	*r = DNSVerificationStatus(str)

	return nil
}

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

func (SSLVerificationStatus) Values() (kinds []string) {
	v := []SSLVerificationStatus{
		SSLVerificationStatusInitializing,
		SSLVerificationStatusPendingValidation,
		SSLVerificationStatusDeleted,
		SSLVerificationStatusPendingIssuance,
		SSLVerificationStatusPendingDeployment,
		SSLVerificationStatusPendingDeletion,
		SSLVerificationStatusPendingExpiration,
		SSLVerificationStatusExpired,
		SSLVerificationStatusActive,
		SSLVerificationStatusInitializingTimedOut,
		SSLVerificationStatusValidationTimedOut,
		SSLVerificationStatusIssuanceTimedOut,
		SSLVerificationStatusDeploymentTimedOut,
		SSLVerificationStatusDeletionTimedOut,
		SSLVerificationStatusPendingCleanup,
		SSLVerificationStatusStagingDeployment,
		SSLVerificationStatusStagingActive,
		SSLVerificationStatusDeactivating,
		SSLVerificationStatusInactive,
		SSLVerificationStatusBackupIssued,
		SSLVerificationStatusHoldingDeployment,
	}
	for _, s := range v {
		kinds = append(kinds, string(s))
	}

	return kinds
}

// String returns the SSLVerificationStatus as a string
func (r SSLVerificationStatus) String() string {
	return string(r)
}

// ToSSLVerificationStatus returns the user status enum based on string input
func ToSSLVerificationStatus(r string) *SSLVerificationStatus {
	rUpper := strings.ToUpper(r)
	switch rUpper {
	case string(SSLVerificationStatusInitializing):
		return &SSLVerificationStatusInitializing
	case string(SSLVerificationStatusPendingValidation):
		return &SSLVerificationStatusPendingValidation
	case string(SSLVerificationStatusDeleted):
		return &SSLVerificationStatusDeleted
	case string(SSLVerificationStatusPendingIssuance):
		return &SSLVerificationStatusPendingIssuance
	case string(SSLVerificationStatusPendingDeployment):
		return &SSLVerificationStatusPendingDeployment
	case string(SSLVerificationStatusPendingDeletion):
		return &SSLVerificationStatusPendingDeletion
	case string(SSLVerificationStatusPendingExpiration):
		return &SSLVerificationStatusPendingExpiration
	case string(SSLVerificationStatusExpired):
		return &SSLVerificationStatusExpired
	case string(SSLVerificationStatusActive):
		return &SSLVerificationStatusActive
	case string(SSLVerificationStatusInitializingTimedOut):
		return &SSLVerificationStatusInitializingTimedOut
	case string(SSLVerificationStatusValidationTimedOut):
		return &SSLVerificationStatusValidationTimedOut
	case string(SSLVerificationStatusIssuanceTimedOut):
		return &SSLVerificationStatusIssuanceTimedOut
	case string(SSLVerificationStatusDeploymentTimedOut):
		return &SSLVerificationStatusDeploymentTimedOut
	case string(SSLVerificationStatusDeletionTimedOut):
		return &SSLVerificationStatusDeletionTimedOut
	case string(SSLVerificationStatusPendingCleanup):
		return &SSLVerificationStatusPendingCleanup
	case string(SSLVerificationStatusStagingDeployment):
		return &SSLVerificationStatusStagingDeployment
	case string(SSLVerificationStatusStagingActive):
		return &SSLVerificationStatusStagingActive
	case string(SSLVerificationStatusDeactivating):
		return &SSLVerificationStatusDeactivating
	case string(SSLVerificationStatusInactive):
		return &SSLVerificationStatusInactive
	case string(SSLVerificationStatusBackupIssued):
		return &SSLVerificationStatusBackupIssued
	case string(SSLVerificationStatusHoldingDeployment):
		return &SSLVerificationStatusHoldingDeployment
	default:
		return &SSLVerificationStatusInvalid
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r SSLVerificationStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *SSLVerificationStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for SSLVerificationStatus, got: %T", v) //nolint:err113
	}

	*r = SSLVerificationStatus(str)

	return nil
}
