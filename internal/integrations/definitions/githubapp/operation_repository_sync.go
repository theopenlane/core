package githubapp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/samber/lo"
	"github.com/shurcooL/githubv4"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// RepositorySync lists repositories accessible to the installation
type RepositorySync struct{}

type pageInfo struct {
	EndCursor   string
	HasNextPage bool
}

type repositoryNode struct {
	NameWithOwner string
	IsPrivate     bool
	UpdatedAt     time.Time
	URL           string `graphql:"url"`
}

// Handle adapts repository sync to the generic operation registration boundary
func (r RepositorySync) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		githubClient, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		return r.Run(ctx, githubClient)
	}
}

// Run enumerates repositories accessible to the installation
func (RepositorySync) Run(ctx context.Context, client GraphQLClient) (json.RawMessage, error) {
	repositories, err := queryRepositories(ctx, client, defaultPageSize)
	if err != nil {
		return nil, err
	}

	sampleSize := min(len(repositories), 10)
	samples := lo.Map(repositories[:sampleSize], func(repository repositoryNode, _ int) map[string]any {
		return map[string]any{
			"name":       repository.NameWithOwner,
			"private":    repository.IsPrivate,
			"updated_at": repository.UpdatedAt,
			"url":        repository.URL,
		}
	})

	return providerkit.EncodeResult(map[string]any{
		"count":   len(repositories),
		"samples": samples,
	}, ErrResultEncode)
}

// queryRepositories lists repositories accessible to the installation
func queryRepositories(ctx context.Context, client GraphQLClient, pageSize int) ([]repositoryNode, error) {
	repositories := make([]repositoryNode, 0)
	pageSize = clampPageSize(pageSize)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Viewer struct {
				Repositories struct {
					Nodes    []repositoryNode
					PageInfo pageInfo
				} `graphql:"repositories(first: $first, after: $after, orderBy: {field: UPDATED_AT, direction: DESC})"`
			}
		}

		variables := map[string]any{
			"first": githubv4.Int(pageSize),
			"after": after,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, normalizeGitHubAPIError(err)
		}

		repositories = append(repositories, query.Viewer.Repositories.Nodes...)
		if !query.Viewer.Repositories.PageInfo.HasNextPage || query.Viewer.Repositories.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Viewer.Repositories.PageInfo.EndCursor))
	}

	return repositories, nil
}

// clampPageSize constrains page sizes to the supported GitHub API range
func clampPageSize(value int) int {
	switch {
	case value <= 0:
		return defaultPageSize
	case value > maxPageSize:
		return maxPageSize
	default:
		return value
	}
}

// normalizeGitHubAPIError collapses provider-specific errors into integration errors
func normalizeGitHubAPIError(err error) error {
	if err == nil {
		return nil
	}

	return ErrAPIRequest
}
