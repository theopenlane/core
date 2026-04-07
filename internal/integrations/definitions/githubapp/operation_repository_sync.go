package githubapp

import (
	"context"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// repositoryAssetVariant is the mapping variant for repository asset payloads
const repositoryAssetVariant = "repository"

// RepositorySync collects repositories accessible to the installation as assets
type RepositorySync struct{}

// IngestHandle adapts repository sync to the ingest operation registration boundary
func (r RepositorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClient(gitHubClient, r.Run)
}

// Run enumerates repositories accessible to the installation and emits Asset ingest payloads
func (RepositorySync) Run(ctx context.Context, client GraphQLClient) ([]types.IngestPayloadSet, error) {
	repositories, err := queryRepositories(ctx, client, defaultPageSize)
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
			Schema:    integrationgenerated.IntegrationMappingSchemaAsset,
			Envelopes: envelopes,
		},
	}, nil
}
