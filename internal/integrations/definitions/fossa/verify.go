package fossa

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify derives the FOSSA organization identity and confirms the token can read issues
func verify(ctx context.Context, req types.ConnectionRequest[CredentialSchema], c *APIClient) (InstallationMetadata, error) {
	metadata, err := c.CheckAccess(ctx)
	if err != nil {
		return InstallationMetadata{}, err
	}

	metadata.BaseURL = baseURLOrDefault(req.Credential.BaseURL)

	return metadata, nil
}

// probeAccess confirms the token can still read issues
func probeAccess(ctx context.Context, _ types.OperationRequest, c *APIClient) error {
	_, err := c.CheckAccess(ctx)

	return err
}

// CheckAccess reads the organization details and issue category counts, both read data so a push-only token fails here rather than at collection time
func (c *APIClient) CheckAccess(ctx context.Context) (InstallationMetadata, error) {
	org, err := c.organization(ctx)
	if err != nil {
		return InstallationMetadata{}, fmt.Errorf("%w: %w", ErrOrganizationFetchFailed, err)
	}

	if org.identifier() == "" {
		return InstallationMetadata{}, ErrOrganizationIDMissing
	}

	if _, err := c.issueCategories(ctx); err != nil {
		return InstallationMetadata{}, fmt.Errorf("%w: %w", ErrCategoriesFetchFailed, err)
	}

	return InstallationMetadata{OrganizationID: org.identifier(), Subscription: org.Subscription}, nil
}
