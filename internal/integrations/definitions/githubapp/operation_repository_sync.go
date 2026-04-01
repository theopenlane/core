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

// RepositorySync collects repositories accessible to the installation as assets
type RepositorySync struct{}

// RepositoryAssetPayload is the normalized repository payload emitted for Asset ingest
type RepositoryAssetPayload struct {
	// NameWithOwner is the full repository name in owner/repo format
	NameWithOwner string `json:"nameWithOwner"`
	// IsPrivate reports whether the repository is private
	IsPrivate bool `json:"isPrivate"`
	// UpdatedAt is the timestamp of the most recent push or metadata update
	UpdatedAt time.Time `json:"updatedAt"`
	// URL is the canonical web URL of the repository
	URL string `json:"url"`
}

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
		payload := RepositoryAssetPayload(repo)

		envelope, err := providerkit.MarshalEnvelopeVariant(repositoryAssetVariant, repo.NameWithOwner, payload, ErrIngestPayloadEncode)
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
