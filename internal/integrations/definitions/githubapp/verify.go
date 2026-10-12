package githubapp

import (
	"context"
	"strconv"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// verify probes the GitHub GraphQL client and derives the installation metadata from the credential
func verify(ctx context.Context, req types.ConnectionRequest[githubAppCredential], client GraphQLClient) (InstallationMetadata, error) {
	if _, err := queryRepositories(ctx, client, 1, nil); err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{
		InstallationID:   strconv.FormatInt(req.Credential.InstallationID, 10),
		OrganizationName: req.Credential.OrganizationName,
	}, nil
}
