package githubapp

import (
	"context"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// repositoryAssetVariant is the mapping variant for repository asset payloads
const repositoryAssetVariant = "repository"

// repositorySyncOperation is the operation ref for the GitHub repository sync operation
var repositorySyncOperation = types.OperationRefOf[RepositorySync]().Ingests(gitHubClient, runRepositorySync)

// runRepositorySync enumerates repositories accessible to the installation and emits Asset ingest payloads
func runRepositorySync(ctx context.Context, request types.OperationRequest, client GraphQLClient, _ RepositorySync) ([]types.IngestPayloadSet, error) {
	repositories, err := queryRepositories(ctx, client, defaultPageSize, request.LastRunAt)
	if err != nil {
		return nil, err
	}

	envelopes := make([]types.MappingEnvelope, 0, len(repositories))

	for _, repo := range repositories {
		envelope, err := providerkit.MarshalEnvelopeVariant(repositoryAssetVariant, repo.NameWithOwner, repo, ErrIngestPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaAsset.Name,
			Envelopes: envelopes,
		},
	}, nil
}
