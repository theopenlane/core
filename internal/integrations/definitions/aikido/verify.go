package aikido

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// accessProbePageSize is the smallest page size the issues export accepts
const accessProbePageSize = 10

// verify derives the Aikido workspace identity and confirms the credentials can read issues
func verify(ctx context.Context, req types.ConnectionRequest[CredentialSchema], c *APIClient) (InstallationMetadata, error) {
	metadata, err := c.CheckAccess(ctx)
	if err != nil {
		return InstallationMetadata{}, err
	}

	metadata.Region = req.Credential.Region
	if metadata.Region == "" {
		metadata.Region = regionEU
	}

	return metadata, nil
}

// verifyOAuth derives the Aikido workspace identity for an installation connected through the Openlane OAuth app
func verifyOAuth(ctx context.Context, _ types.ConnectionRequest[aikidoOAuthCred], c *APIClient) (InstallationMetadata, error) {
	metadata, err := c.CheckAccess(ctx)
	if err != nil {
		return InstallationMetadata{}, err
	}

	metadata.Region = regionEU

	return metadata, nil
}

// CheckAccess reads the workspace and one page of issues, so credentials missing the issues:read scope fail here rather than at collection time
func (c *APIClient) CheckAccess(ctx context.Context) (InstallationMetadata, error) {
	workspace, err := c.workspace(ctx)
	if err != nil {
		return InstallationMetadata{}, fmt.Errorf("%w: %w", ErrWorkspaceFetchFailed, err)
	}

	if workspace.identifier() == "" {
		return InstallationMetadata{}, ErrWorkspaceIDMissing
	}

	if _, _, err := c.listPage(ctx, pathIssuesExport, map[string]string{"format": "json", "filter_status": "open"}, 0, accessProbePageSize); err != nil {
		return InstallationMetadata{}, fmt.Errorf("%w: %w", ErrIssuesFetchFailed, err)
	}

	return InstallationMetadata{
		WorkspaceID:    workspace.identifier(),
		WorkspaceName:  workspace.Name,
		LinkedProvider: workspace.LinkedProvider,
	}, nil
}
