package enums

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomDomainStatusValues(t *testing.T) {
	expected := []string{"INVALID", "VERIFIED", "FAILED_VERIFY", "PENDING"}
	values := CustomDomainStatus("").Values()

	assert.Equal(t, len(expected), len(values))

	for i, v := range values {
		assert.Equal(t, expected[i], v)
	}
}

func TestCustomDomainStatusString(t *testing.T) {
	tests := []struct {
		status   CustomDomainStatus
		expected string
	}{
		{CustomDomainStatusVerified, "VERIFIED"},
		{CustomDomainStatusFailedVerify, "FAILED_VERIFY"},
		{CustomDomainStatusPending, "PENDING"},
		{CustomDomainStatusInvalid, "INVALID"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.status.String())
	}
}

func TestToCustomDomainStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected *CustomDomainStatus
	}{
		{"VERIFIED", &CustomDomainStatusVerified},
		{"FAILED_VERIFY", &CustomDomainStatusFailedVerify},
		{"PENDING", &CustomDomainStatusPending},
		{"verified", &CustomDomainStatusVerified},
		{"failed_verify", &CustomDomainStatusFailedVerify},
		{"pending", &CustomDomainStatusPending},
		{"unknown", &CustomDomainStatusInvalid},
		{"", &CustomDomainStatusInvalid},
	}

	for _, test := range tests {
		result := ToCustomDomainStatus(test.input)
		assert.Equal(t, test.expected, result, "ToCustomDomainStatus(%q)", test.input)
	}
}

func TestCustomDomainStatusMarshalGQL(t *testing.T) {
	tests := []struct {
		status   CustomDomainStatus
		expected string
	}{
		{CustomDomainStatusVerified, `"VERIFIED"`},
		{CustomDomainStatusFailedVerify, `"FAILED_VERIFY"`},
		{CustomDomainStatusPending, `"PENDING"`},
		{CustomDomainStatusInvalid, `"INVALID"`},
	}

	for _, test := range tests {
		var writer strings.Builder
		test.status.MarshalGQL(&writer)

		assert.Equal(t, test.expected, writer.String())
	}
}

func TestCustomDomainStatusUnmarshalGQL(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected CustomDomainStatus
		hasError bool
	}{
		{"VERIFIED", CustomDomainStatusVerified, false},
		{"FAILED_VERIFY", CustomDomainStatusFailedVerify, false},
		{"PENDING", CustomDomainStatusPending, false},
		{"INVALID", CustomDomainStatusInvalid, false},
		{123, "", true},
	}

	for _, test := range tests {
		var status CustomDomainStatus
		err := status.UnmarshalGQL(test.input)
		if test.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, test.expected, status)
		}
	}
}
