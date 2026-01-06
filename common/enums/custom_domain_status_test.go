package enums

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSVerificationStatusValues(t *testing.T) {
	expected := []string{
		"ACTIVE", "PENDING", "ACTIVE_REDEPLOYING", "MOVED", "PENDING_DELETION",
		"DELETED", "PENDING_BLOCKED", "PENDING_MIGRATION", "PENDING_PROVISIONED",
		"TEST_PENDING", "TEST_ACTIVE", "TEST_ACTIVE_APEX", "TEST_BLOCKED",
		"TEST_FAILED", "PROVISIONED", "BLOCKED",
	}

	values := DNSVerificationStatus("").Values()
	assert.Equal(t, expected, values)
}

func TestDNSVerificationStatusString(t *testing.T) {
	tests := []struct {
		status   DNSVerificationStatus
		expected string
	}{
		{DNSVerificationStatusActive, "ACTIVE"},
		{DNSVerificationStatusPending, "PENDING"},
		{DNSVerificationStatusActiveRedeploying, "ACTIVE_REDEPLOYING"},
		{DNSVerificationStatusMoved, "MOVED"},
		{DNSVerificationStatusPendingDeletion, "PENDING_DELETION"},
		{DNSVerificationStatusDeleted, "DELETED"},
		{DNSVerificationStatusInvalid, "INVALID"},
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
		{DNSVerificationStatusActive, `"ACTIVE"`},
		{DNSVerificationStatusPending, `"PENDING"`},
		{DNSVerificationStatusActiveRedeploying, `"ACTIVE_REDEPLOYING"`},
		{DNSVerificationStatusMoved, `"MOVED"`},
		{DNSVerificationStatusInvalid, `"INVALID"`},
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
		{"ACTIVE", DNSVerificationStatusActive, false},
		{"PENDING", DNSVerificationStatusPending, false},
		{"ACTIVE_REDEPLOYING", DNSVerificationStatusActiveRedeploying, false},
		{"MOVED", DNSVerificationStatusMoved, false},
		{"INVALID", DNSVerificationStatusInvalid, false},
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
		"INITIALIZING", "PENDING_VALIDATION", "DELETED", "PENDING_ISSUANCE",
		"PENDING_DEPLOYMENT", "PENDING_DELETION", "PENDING_EXPIRATION", "EXPIRED",
		"ACTIVE", "INITIALIZING_TIMED_OUT", "VALIDATION_TIMED_OUT", "ISSUANCE_TIMED_OUT",
		"DEPLOYMENT_TIMED_OUT", "DELETION_TIMED_OUT", "PENDING_CLEANUP", "STAGING_DEPLOYMENT",
		"STAGING_ACTIVE", "DEACTIVATING", "INACTIVE", "BACKUP_ISSUED", "HOLDING_DEPLOYMENT",
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
		{SSLVerificationStatusInitializing, "INITIALIZING"},
		{SSLVerificationStatusPendingValidation, "PENDING_VALIDATION"},
		{SSLVerificationStatusDeleted, "DELETED"},
		{SSLVerificationStatusPendingIssuance, "PENDING_ISSUANCE"},
		{SSLVerificationStatusActive, "ACTIVE"},
		{SSLVerificationStatusExpired, "EXPIRED"},
		{SSLVerificationStatusInvalid, "INVALID"},
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
		{SSLVerificationStatusInitializing, `"INITIALIZING"`},
		{SSLVerificationStatusPendingValidation, `"PENDING_VALIDATION"`},
		{SSLVerificationStatusDeleted, `"DELETED"`},
		{SSLVerificationStatusActive, `"ACTIVE"`},
		{SSLVerificationStatusInvalid, `"INVALID"`},
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
		{"INITIALIZING", SSLVerificationStatusInitializing, false},
		{"PENDING_VALIDATION", SSLVerificationStatusPendingValidation, false},
		{"DELETED", SSLVerificationStatusDeleted, false},
		{"ACTIVE", SSLVerificationStatusActive, false},
		{"INVALID", SSLVerificationStatusInvalid, false},
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
