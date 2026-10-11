package awssecurityhub

import (
	"testing"

	"gotest.tools/v3/assert"
)

// TestAssumeRoleMetadataARNAccountUsedWhenBlank verifies the account id is derived from the role ARN when the credential omits it
func TestAssumeRoleMetadataARNAccountUsedWhenBlank(t *testing.T) {
	metadata, err := assumeRoleMetadata(AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
	})
	assert.NilError(t, err)
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
	assert.NilError(t, err)
	assert.Equal(t, "123456789012", metadata.AccountID)
}

// TestAssumeRoleMetadataMismatchedAccountIDRejected verifies a configured account id that disagrees with the role ARN is rejected
func TestAssumeRoleMetadataMismatchedAccountIDRejected(t *testing.T) {
	_, err := assumeRoleMetadata(AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
		AccountID:  "999999999999",
	})
	assert.ErrorIs(t, err, ErrAccountIDMismatch)
}

// TestAssumeRoleMetadataMalformedARNRejected verifies a role ARN without an account segment is rejected
func TestAssumeRoleMetadataMalformedARNRejected(t *testing.T) {
	_, err := assumeRoleMetadata(AssumeRoleCredentialSchema{
		RoleARN:    "not-a-valid-arn",
		HomeRegion: "us-east-1",
	})
	assert.ErrorIs(t, err, ErrRoleARNInvalid)
}
