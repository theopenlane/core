package objectstore

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
)

// serviceAccountIdentity represents the identity fields extracted from a service account key
type serviceAccountIdentity struct {
	// ClientEmail is the email address of the service account
	ClientEmail string `json:"client_email"`
}

// resolveInstallationMetadata derives installation metadata from whichever credential slot the installation bound
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	slot, ok := boundSlot(req.Credentials)
	if !ok {
		return InstallationMetadata{}, false, ErrCredentialMetadataRequired
	}

	var (
		meta InstallationMetadata
		err  error
	)

	switch slot {
	case workloadIdentityCredential.ID():
		meta, err = installationMetadataFrom(req.Credentials, workloadIdentityCredential, func(cred WorkloadIdentityCredentialSchema) InstallationMetadata {
			return InstallationMetadata{
				Provider:            string(storage.GCSProvider),
				Bucket:              cred.Bucket,
				ProjectID:           cred.ProjectID,
				ServiceAccountEmail: cred.ServiceAccountEmail,
			}
		})
	case serviceAccountCredential.ID():
		meta, err = installationMetadataFrom(req.Credentials, serviceAccountCredential, func(cred ServiceAccountCredentialSchema) InstallationMetadata {
			return InstallationMetadata{
				Provider:            string(storage.GCSProvider),
				Bucket:              cred.Bucket,
				ProjectID:           cred.ProjectID,
				ServiceAccountEmail: serviceAccountEmail(cred.ServiceAccountKey),
			}
		})
	case awsAssumeRoleCredential.ID():
		meta, err = installationMetadataFrom(req.Credentials, awsAssumeRoleCredential, func(cred AWSAssumeRoleCredentialSchema) InstallationMetadata {
			return InstallationMetadata{
				Provider: string(storage.S3Provider),
				Bucket:   cred.Bucket,
				Region:   cred.Region,
			}
		})
	case awsAccessKeyCredential.ID():
		meta, err = installationMetadataFrom(req.Credentials, awsAccessKeyCredential, func(cred AWSAccessKeyCredentialSchema) InstallationMetadata {
			return InstallationMetadata{
				Provider: string(storage.S3Provider),
				Bucket:   cred.Bucket,
				Region:   cred.Region,
			}
		})
	case r2Credential.ID():
		meta, err = installationMetadataFrom(req.Credentials, r2Credential, func(cred R2CredentialSchema) InstallationMetadata {
			return InstallationMetadata{
				Provider: string(storage.R2Provider),
				Bucket:   cred.Bucket,
			}
		})
	}

	if err != nil {
		return InstallationMetadata{}, false, err
	}

	if meta.Bucket == "" {
		return InstallationMetadata{}, false, ErrBucketRequired
	}

	return meta, true, nil
}

// installationMetadataFrom decodes the credential bound to the slot and projects it to installation metadata
func installationMetadataFrom[T any](bindings types.CredentialBindings, slot types.CredentialRef[T], project func(T) InstallationMetadata) (InstallationMetadata, error) {
	cred, err := decodeCredential(bindings, slot)
	if err != nil {
		return InstallationMetadata{}, err
	}

	return project(cred), nil
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
