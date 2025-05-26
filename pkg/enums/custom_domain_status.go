package enums

import (
	"fmt"
	"io"
	"strings"
)

type DNSVerificationStatus string

var (
	DNSVerificationStatusActive             DNSVerificationStatus = "active"
	DNSVerificationStatusPending            DNSVerificationStatus = "pending"
	DNSVerificationStatusActiveRedeploying  DNSVerificationStatus = "active_redeploying"
	DNSVerificationStatusMoved              DNSVerificationStatus = "moved"
	DNSVerificationStatusPendingDeletion    DNSVerificationStatus = "pending_deletion"
	DNSVerificationStatusDeleted            DNSVerificationStatus = "deleted"
	DNSVerificationStatusPendingBlocked     DNSVerificationStatus = "pending_blocked"
	DNSVerificationStatusPendingMigration   DNSVerificationStatus = "pending_migration"
	DNSVerificationStatusPendingProvisioned DNSVerificationStatus = "pending_provisioned"
	DNSVerificationStatusTestPending        DNSVerificationStatus = "test_pending"
	DNSVerificationStatusTestActive         DNSVerificationStatus = "test_active"
	DNSVerificationStatusTestActiveApex     DNSVerificationStatus = "test_active_apex"
	DNSVerificationStatusTestBlocked        DNSVerificationStatus = "test_blocked"
	DNSVerificationStatusTestFailed         DNSVerificationStatus = "test_failed"
	DNSVerificationStatusProvisioned        DNSVerificationStatus = "provisioned"
	DNSVerificationStatusBlocked            DNSVerificationStatus = "blocked"
	DNSVerificationStatusInvalid            DNSVerificationStatus = "invalid"
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
	switch r := strings.ToLower(r); r {
	case DNSVerificationStatusPending.String():
		return &DNSVerificationStatusPending
	case DNSVerificationStatusActive.String():
		return &DNSVerificationStatusActive
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
	SSLVerificationStatusInitializing         SSLVerificationStatus = "initializing"
	SSLVerificationStatusPendingValidation    SSLVerificationStatus = "pending_validation"
	SSLVerificationStatusDeleted              SSLVerificationStatus = "deleted"
	SSLVerificationStatusPendingIssuance      SSLVerificationStatus = "pending_issuance"
	SSLVerificationStatusPendingDeployment    SSLVerificationStatus = "pending_deployment"
	SSLVerificationStatusPendingDeletion      SSLVerificationStatus = "pending_deletion"
	SSLVerificationStatusPendingExpiration    SSLVerificationStatus = "pending_expiration"
	SSLVerificationStatusExpired              SSLVerificationStatus = "expired"
	SSLVerificationStatusActive               SSLVerificationStatus = "active"
	SSLVerificationStatusInitializingTimedOut SSLVerificationStatus = "initializing_timed_out"
	SSLVerificationStatusValidationTimedOut   SSLVerificationStatus = "validation_timed_out"
	SSLVerificationStatusIssuanceTimedOut     SSLVerificationStatus = "issuance_timed_out"
	SSLVerificationStatusDeploymentTimedOut   SSLVerificationStatus = "deployment_timed_out"
	SSLVerificationStatusDeletionTimedOut     SSLVerificationStatus = "deletion_timed_out"
	SSLVerificationStatusPendingCleanup       SSLVerificationStatus = "pending_cleanup"
	SSLVerificationStatusStagingDeployment    SSLVerificationStatus = "staging_deployment"
	SSLVerificationStatusStagingActive        SSLVerificationStatus = "staging_active"
	SSLVerificationStatusDeactivating         SSLVerificationStatus = "deactivating"
	SSLVerificationStatusInactive             SSLVerificationStatus = "inactive"
	SSLVerificationStatusBackupIssued         SSLVerificationStatus = "backup_issued"
	SSLVerificationStatusHoldingDeployment    SSLVerificationStatus = "holding_deployment"
	SSLVerificationStatusInvalid              SSLVerificationStatus = "invalid"
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
	switch r := strings.ToLower(r); r {
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
