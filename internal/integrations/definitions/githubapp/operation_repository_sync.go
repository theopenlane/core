package githubapp

import (
	"context"
	"time"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// repositoryAssetVariant is the mapping variant for repository asset payloads
const repositoryAssetVariant = "repository"

// IngestHandle adapts repository sync to the ingest operation registration boundary
func (r RepositorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(gitHubClient, func(ctx context.Context, request types.OperationRequest, client GraphQLClient) ([]types.IngestPayloadSet, error) {
		return r.Run(ctx, client, request.LastRunAt)
	})
}

// Run enumerates repositories accessible to the installation and emits Asset ingest payloads
func (RepositorySync) Run(ctx context.Context, client GraphQLClient, lastRunAt *time.Time) ([]types.IngestPayloadSet, error) {
	repositories, err := queryRepositories(ctx, client, defaultPageSize, lastRunAt)
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
