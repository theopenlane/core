package objectstore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// serviceAccountIdentity represents the identity fields extracted from a service account key
type serviceAccountIdentity struct {
	// ClientEmail is the email address of the service account
	ClientEmail string `json:"client_email"`
}

// verifyWorkloadIdentity lists the bucket through the federated client and derives the installation metadata
func verifyWorkloadIdentity(ctx context.Context, req types.ConnectionRequest[WorkloadIdentityCredentialSchema], c *Client) (InstallationMetadata, error) {
	return verifyBucket(ctx, c, InstallationMetadata{
		Provider:            storage.GCSProvider,
		Bucket:              req.Credential.Bucket,
		ProjectID:           req.Credential.ProjectID,
		ServiceAccountEmail: req.Credential.ServiceAccountEmail,
	})
}

// verifyServiceAccount lists the bucket through the service account client and derives the installation metadata
func verifyServiceAccount(ctx context.Context, req types.ConnectionRequest[ServiceAccountCredentialSchema], c *Client) (InstallationMetadata, error) {
	return verifyBucket(ctx, c, InstallationMetadata{
		Provider:            storage.GCSProvider,
		Bucket:              req.Credential.Bucket,
		ProjectID:           req.Credential.ProjectID,
		ServiceAccountEmail: serviceAccountEmail(req.Credential.ServiceAccountKey),
	})
}

// verifyAssumeRole lists the bucket through the assumed role and derives the installation metadata
func verifyAssumeRole(ctx context.Context, req types.ConnectionRequest[AWSAssumeRoleCredentialSchema], c *Client) (InstallationMetadata, error) {
	return verifyBucket(ctx, c, InstallationMetadata{
		Provider: storage.S3Provider,
		Bucket:   req.Credential.Bucket,
		Region:   req.Credential.Region,
	})
}

// verifyAccessKey lists the bucket with the static IAM keys and derives the installation metadata
func verifyAccessKey(ctx context.Context, req types.ConnectionRequest[AWSAccessKeyCredentialSchema], c *Client) (InstallationMetadata, error) {
	return verifyBucket(ctx, c, InstallationMetadata{
		Provider: storage.S3Provider,
		Bucket:   req.Credential.Bucket,
		Region:   req.Credential.Region,
	})
}

// verifyR2 lists the bucket with the R2 keys and derives the installation metadata
func verifyR2(ctx context.Context, req types.ConnectionRequest[R2CredentialSchema], c *Client) (InstallationMetadata, error) {
	return verifyBucket(ctx, c, InstallationMetadata{
		Provider: storage.R2Provider,
		Bucket:   req.Credential.Bucket,
	})
}

// verifyBucket lists one object under the bucket root to prove the credentials reach the bucket named by the metadata
func verifyBucket(ctx context.Context, c *Client, metadata InstallationMetadata) (InstallationMetadata, error) {
	if metadata.Bucket == "" {
		return InstallationMetadata{}, ErrBucketRequired
	}

	if _, err := c.Provider.ListObjects(ctx, "", 1); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("provider", metadata.Provider.String()).Msg("objectstore: bucket listing failed")

		return InstallationMetadata{}, fmt.Errorf("%w: %w", ErrVerifyFailed, err)
	}

	return metadata, nil
}

// serviceAccountEmail returns the client email carried by a service account key
func serviceAccountEmail(rawKey string) string {
	key := normalizeServiceAccountKey(rawKey)
	if key == "" {
		return ""
	}

	var identity serviceAccountIdentity

	_ = json.Unmarshal([]byte(key), &identity)

	return identity.ClientEmail
}
