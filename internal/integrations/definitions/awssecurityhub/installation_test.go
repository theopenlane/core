package awssecurityhub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// TestResolveInstallationMetadata_AssumeRole_ARNAccountUsedWhenBlank derives account id from ARN
func TestResolveInstallationMetadata_AssumeRole_ARNAccountUsedWhenBlank(t *testing.T) {
	bindings := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
	})

	metadata, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{Credentials: bindings})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "123456789012", metadata.AccountID)
}

// TestResolveInstallationMetadata_AssumeRole_MatchingAccountIDAccepted is valid
func TestResolveInstallationMetadata_AssumeRole_MatchingAccountIDAccepted(t *testing.T) {
	bindings := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
		AccountID:  "123456789012",
	})

	metadata, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{Credentials: bindings})
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "123456789012", metadata.AccountID)
}

// TestResolveInstallationMetadata_AssumeRole_MismatchedAccountIDRejected rejects mismatch
func TestResolveInstallationMetadata_AssumeRole_MismatchedAccountIDRejected(t *testing.T) {
	bindings := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
		RoleARN:    "arn:aws:iam::123456789012:role/MyRole",
		HomeRegion: "us-east-1",
		AccountID:  "999999999999",
	})

	_, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{Credentials: bindings})
	require.ErrorIs(t, err, ErrAccountIDMismatch)
	assert.False(t, ok)
}

// TestResolveInstallationMetadata_AssumeRole_MalformedARNRejected rejects malformed ARN
func TestResolveInstallationMetadata_AssumeRole_MalformedARNRejected(t *testing.T) {
	bindings := makeAssumeRoleBindings(t, AssumeRoleCredentialSchema{
		RoleARN:    "not-a-valid-arn",
		HomeRegion: "us-east-1",
	})

	_, ok, err := resolveInstallationMetadata(context.Background(), types.InstallationRequest{Credentials: bindings})
	require.ErrorIs(t, err, ErrRoleARNInvalid)
	assert.False(t, ok)
}
