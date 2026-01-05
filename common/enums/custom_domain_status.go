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
	rUpper := strings.ToUpper(r)
	switch rUpper {
	case "ACTIVE":
		return &DNSVerificationStatusActive
	case "PENDING":
		return &DNSVerificationStatusPending
	case "ACTIVE_REDEPLOYING":
		return &DNSVerificationStatusActiveRedeploying
	case "MOVED":
		return &DNSVerificationStatusMoved
	case "PENDING_DELETION":
		return &DNSVerificationStatusPendingDeletion
	case "DELETED":
		return &DNSVerificationStatusDeleted
	case "PENDING_BLOCKED":
		return &DNSVerificationStatusPendingBlocked
	case "PENDING_MIGRATION":
		return &DNSVerificationStatusPendingMigration
	case "PENDING_PROVISIONED":
		return &DNSVerificationStatusPendingProvisioned
	case "TEST_PENDING":
		return &DNSVerificationStatusTestPending
	case "TEST_ACTIVE":
		return &DNSVerificationStatusTestActive
	case "TEST_ACTIVE_APEX":
		return &DNSVerificationStatusTestActiveApex
	case "TEST_BLOCKED":
		return &DNSVerificationStatusTestBlocked
	case "TEST_FAILED":
		return &DNSVerificationStatusTestFailed
	case "PROVISIONED":
		return &DNSVerificationStatusProvisioned
	case "BLOCKED":
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
	rUpper := strings.ToUpper(r)
	switch rUpper {
	case "INITIALIZING":
		return &SSLVerificationStatusInitializing
	case "PENDING_VALIDATION":
		return &SSLVerificationStatusPendingValidation
	case "DELETED":
		return &SSLVerificationStatusDeleted
	case "PENDING_ISSUANCE":
		return &SSLVerificationStatusPendingIssuance
	case "PENDING_DEPLOYMENT":
		return &SSLVerificationStatusPendingDeployment
	case "PENDING_DELETION":
		return &SSLVerificationStatusPendingDeletion
	case "PENDING_EXPIRATION":
		return &SSLVerificationStatusPendingExpiration
	case "EXPIRED":
		return &SSLVerificationStatusExpired
	case "ACTIVE":
		return &SSLVerificationStatusActive
	case "INITIALIZING_TIMED_OUT":
		return &SSLVerificationStatusInitializingTimedOut
	case "VALIDATION_TIMED_OUT":
		return &SSLVerificationStatusValidationTimedOut
	case "ISSUANCE_TIMED_OUT":
		return &SSLVerificationStatusIssuanceTimedOut
	case "DEPLOYMENT_TIMED_OUT":
		return &SSLVerificationStatusDeploymentTimedOut
	case "DELETION_TIMED_OUT":
		return &SSLVerificationStatusDeletionTimedOut
	case "PENDING_CLEANUP":
		return &SSLVerificationStatusPendingCleanup
	case "STAGING_DEPLOYMENT":
		return &SSLVerificationStatusStagingDeployment
	case "STAGING_ACTIVE":
		return &SSLVerificationStatusStagingActive
	case "DEACTIVATING":
		return &SSLVerificationStatusDeactivating
	case "INACTIVE":
		return &SSLVerificationStatusInactive
	case "BACKUP_ISSUED":
		return &SSLVerificationStatusBackupIssued
	case "HOLDING_DEPLOYMENT":
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
