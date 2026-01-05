package enums

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSVerificationStatusValues(t *testing.T) {
	expected := []string{
		"active", "pending", "active_redeploying", "moved", "pending_deletion",
		"deleted", "pending_blocked", "pending_migration", "pending_provisioned",
		"test_pending", "test_active", "test_active_apex", "test_blocked",
		"test_failed", "provisioned", "blocked",
	}
	values := DNSVerificationStatus("").Values()

	assert.Equal(t, len(expected), len(values))

	for i, v := range values {
		assert.Equal(t, expected[i], v)
	}
}

func TestDNSVerificationStatusString(t *testing.T) {
	tests := []struct {
		status   DNSVerificationStatus
		expected string
	}{
		{DNSVerificationStatusActive, "active"},
		{DNSVerificationStatusPending, "pending"},
		{DNSVerificationStatusActiveRedeploying, "active_redeploying"},
		{DNSVerificationStatusMoved, "moved"},
		{DNSVerificationStatusPendingDeletion, "pending_deletion"},
		{DNSVerificationStatusDeleted, "deleted"},
		{DNSVerificationStatusInvalid, "invalid"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.status.String())
	}
}

func TestToDNSVerificationStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected *DNSVerificationStatus
	}{
		{"ACTIVE", &DNSVerificationStatusActive},
		{"PENDING", &DNSVerificationStatusPending},
		{"ACTIVE_REDEPLOYING", &DNSVerificationStatusActiveRedeploying},
		{"MOVED", &DNSVerificationStatusMoved},
		{"PENDING_DELETION", &DNSVerificationStatusPendingDeletion},
		{"DELETED", &DNSVerificationStatusDeleted},
		{"active", &DNSVerificationStatusActive},
		{"pending", &DNSVerificationStatusPending},
		{"unknown", &DNSVerificationStatusInvalid},
		{"", &DNSVerificationStatusInvalid},
	}

	for _, test := range tests {
		result := ToDNSVerificationStatus(test.input)
		assert.Equal(t, test.expected, result, "ToDNSVerificationStatus(%q)", test.input)
	}
}

func TestDNSVerificationStatusMarshalGQL(t *testing.T) {
	tests := []struct {
		status   DNSVerificationStatus
		expected string
	}{
		{DNSVerificationStatusActive, `"active"`},
		{DNSVerificationStatusPending, `"pending"`},
		{DNSVerificationStatusActiveRedeploying, `"active_redeploying"`},
		{DNSVerificationStatusMoved, `"moved"`},
		{DNSVerificationStatusInvalid, `"invalid"`},
	}

	for _, test := range tests {
		var writer strings.Builder
		test.status.MarshalGQL(&writer)

		assert.Equal(t, test.expected, writer.String())
	}
}

func TestDNSVerificationStatusUnmarshalGQL(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected DNSVerificationStatus
		hasError bool
	}{
		{"active", DNSVerificationStatusActive, false},
		{"pending", DNSVerificationStatusPending, false},
		{"active_redeploying", DNSVerificationStatusActiveRedeploying, false},
		{"moved", DNSVerificationStatusMoved, false},
		{"invalid", DNSVerificationStatusInvalid, false},
		{123, "", true},
	}

	for _, test := range tests {
		var status DNSVerificationStatus
		err := status.UnmarshalGQL(test.input)
		if test.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, test.expected, status)
		}
	}
}

func TestSSLVerificationStatusValues(t *testing.T) {
	expected := []string{
		"initializing", "pending_validation", "deleted", "pending_issuance",
		"pending_deployment", "pending_deletion", "pending_expiration", "expired",
		"active", "initializing_timed_out", "validation_timed_out", "issuance_timed_out",
		"deployment_timed_out", "deletion_timed_out", "pending_cleanup", "staging_deployment",
		"staging_active", "deactivating", "inactive", "backup_issued", "holding_deployment",
	}
	values := SSLVerificationStatus("").Values()

	assert.Equal(t, len(expected), len(values))

	for i, v := range values {
		assert.Equal(t, expected[i], v)
	}
}

func TestSSLVerificationStatusString(t *testing.T) {
	tests := []struct {
		status   SSLVerificationStatus
		expected string
	}{
		{SSLVerificationStatusInitializing, "initializing"},
		{SSLVerificationStatusPendingValidation, "pending_validation"},
		{SSLVerificationStatusDeleted, "deleted"},
		{SSLVerificationStatusPendingIssuance, "pending_issuance"},
		{SSLVerificationStatusActive, "active"},
		{SSLVerificationStatusExpired, "expired"},
		{SSLVerificationStatusInvalid, "invalid"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.status.String())
	}
}

func TestToSSLVerificationStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected *SSLVerificationStatus
	}{
		{"INITIALIZING", &SSLVerificationStatusInitializing},
		{"PENDING_VALIDATION", &SSLVerificationStatusPendingValidation},
		{"DELETED", &SSLVerificationStatusDeleted},
		{"PENDING_ISSUANCE", &SSLVerificationStatusPendingIssuance},
		{"ACTIVE", &SSLVerificationStatusActive},
		{"EXPIRED", &SSLVerificationStatusExpired},
		{"initializing", &SSLVerificationStatusInitializing},
		{"active", &SSLVerificationStatusActive},
		{"unknown", &SSLVerificationStatusInvalid},
		{"", &SSLVerificationStatusInvalid},
	}

	for _, test := range tests {
		result := ToSSLVerificationStatus(test.input)
		assert.Equal(t, test.expected, result, "ToSSLVerificationStatus(%q)", test.input)
	}
}

func TestSSLVerificationStatusMarshalGQL(t *testing.T) {
	tests := []struct {
		status   SSLVerificationStatus
		expected string
	}{
		{SSLVerificationStatusInitializing, `"initializing"`},
		{SSLVerificationStatusPendingValidation, `"pending_validation"`},
		{SSLVerificationStatusDeleted, `"deleted"`},
		{SSLVerificationStatusActive, `"active"`},
		{SSLVerificationStatusInvalid, `"invalid"`},
	}

	for _, test := range tests {
		var writer strings.Builder
		test.status.MarshalGQL(&writer)

		assert.Equal(t, test.expected, writer.String())
	}
}

func TestSSLVerificationStatusUnmarshalGQL(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected SSLVerificationStatus
		hasError bool
	}{
		{"initializing", SSLVerificationStatusInitializing, false},
		{"pending_validation", SSLVerificationStatusPendingValidation, false},
		{"deleted", SSLVerificationStatusDeleted, false},
		{"active", SSLVerificationStatusActive, false},
		{"invalid", SSLVerificationStatusInvalid, false},
		{123, "", true},
	}

	for _, test := range tests {
		var status SSLVerificationStatus
		err := status.UnmarshalGQL(test.input)
		if test.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, test.expected, status)
		}
	}
}
