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
	case DNSVerificationStatusActive.String():
		return &DNSVerificationStatusActive
	case DNSVerificationStatusPending.String():
		return &DNSVerificationStatusPending
	case DNSVerificationStatusActiveRedeploying.String():
		return &DNSVerificationStatusActiveRedeploying
	case DNSVerificationStatusMoved.String():
		return &DNSVerificationStatusMoved
	case DNSVerificationStatusPendingDeletion.String():
		return &DNSVerificationStatusPendingDeletion
	case DNSVerificationStatusDeleted.String():
		return &DNSVerificationStatusDeleted
	case DNSVerificationStatusPendingBlocked.String():
		return &DNSVerificationStatusPendingBlocked
	case DNSVerificationStatusPendingMigration.String():
		return &DNSVerificationStatusPendingMigration
	case DNSVerificationStatusPendingProvisioned.String():
		return &DNSVerificationStatusPendingProvisioned
	case DNSVerificationStatusTestPending.String():
		return &DNSVerificationStatusTestPending
	case DNSVerificationStatusTestActive.String():
		return &DNSVerificationStatusTestActive
	case DNSVerificationStatusTestActiveApex.String():
		return &DNSVerificationStatusTestActiveApex
	case DNSVerificationStatusTestBlocked.String():
		return &DNSVerificationStatusTestBlocked
	case DNSVerificationStatusTestFailed.String():
		return &DNSVerificationStatusTestFailed
	case DNSVerificationStatusProvisioned.String():
		return &DNSVerificationStatusProvisioned
	case DNSVerificationStatusBlocked.String():
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
	case SSLVerificationStatusInitializing.String():
		return &SSLVerificationStatusInitializing
	case SSLVerificationStatusPendingValidation.String():
		return &SSLVerificationStatusPendingValidation
	case SSLVerificationStatusDeleted.String():
		return &SSLVerificationStatusDeleted
	case SSLVerificationStatusPendingIssuance.String():
		return &SSLVerificationStatusPendingIssuance
	case SSLVerificationStatusPendingDeployment.String():
		return &SSLVerificationStatusPendingDeployment
	case SSLVerificationStatusPendingDeletion.String():
		return &SSLVerificationStatusPendingDeletion
	case SSLVerificationStatusPendingExpiration.String():
		return &SSLVerificationStatusPendingExpiration
	case SSLVerificationStatusExpired.String():
		return &SSLVerificationStatusExpired
	case SSLVerificationStatusActive.String():
		return &SSLVerificationStatusActive
	case SSLVerificationStatusInitializingTimedOut.String():
		return &SSLVerificationStatusInitializingTimedOut
	case SSLVerificationStatusValidationTimedOut.String():
		return &SSLVerificationStatusValidationTimedOut
	case SSLVerificationStatusIssuanceTimedOut.String():
		return &SSLVerificationStatusIssuanceTimedOut
	case SSLVerificationStatusDeploymentTimedOut.String():
		return &SSLVerificationStatusDeploymentTimedOut
	case SSLVerificationStatusDeletionTimedOut.String():
		return &SSLVerificationStatusDeletionTimedOut
	case SSLVerificationStatusPendingCleanup.String():
		return &SSLVerificationStatusPendingCleanup
	case SSLVerificationStatusStagingDeployment.String():
		return &SSLVerificationStatusStagingDeployment
	case SSLVerificationStatusStagingActive.String():
		return &SSLVerificationStatusStagingActive
	case SSLVerificationStatusDeactivating.String():
		return &SSLVerificationStatusDeactivating
	case SSLVerificationStatusInactive.String():
		return &SSLVerificationStatusInactive
	case SSLVerificationStatusBackupIssued.String():
		return &SSLVerificationStatusBackupIssued
	case SSLVerificationStatusHoldingDeployment.String():
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
