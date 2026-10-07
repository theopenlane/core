package awssecurityhub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAssumeRoleMetadataARNAccountUsedWhenBlank verifies the account id is derived from the role ARN when the credential omits it
func TestAssumeRoleMetadataARNAccountUsedWhenBlank(t *testing.T) {
	metadata, err := assumeRoleMetadata(AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "123456789012", metadata.AccountID)
	assert.Equal(t, "us-east-1", metadata.HomeRegion)
}

// TestAssumeRoleMetadataMatchingAccountIDAccepted verifies a configured account id matching the role ARN is accepted
func TestAssumeRoleMetadataMatchingAccountIDAccepted(t *testing.T) {
	metadata, err := assumeRoleMetadata(AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
		AccountID:  "123456789012",
	})
	require.NoError(t, err)
	assert.Equal(t, "123456789012", metadata.AccountID)
}

// TestAssumeRoleMetadataMismatchedAccountIDRejected verifies a configured account id that disagrees with the role ARN is rejected
func TestAssumeRoleMetadataMismatchedAccountIDRejected(t *testing.T) {
	_, err := assumeRoleMetadata(AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
		AccountID:  "999999999999",
	})
	require.ErrorIs(t, err, ErrAccountIDMismatch)
}

// TestAssumeRoleMetadataMalformedARNRejected verifies a role ARN without an account segment is rejected
func TestAssumeRoleMetadataMalformedARNRejected(t *testing.T) {
	_, err := assumeRoleMetadata(AssumeRoleCredentialSchema{
		RoleARN:    "not-a-valid-arn",
		HomeRegion: "us-east-1",
	})
	require.ErrorIs(t, err, ErrRoleARNInvalid)
}
