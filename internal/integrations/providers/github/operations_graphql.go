package github

import (
	"context"
	"fmt"
	"time"

	"github.com/shurcooL/githubv4"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/integrations/types"
)

// githubOrgRepoOperationConfig captures GraphQL organization repository collection settings.
type githubOrgRepoOperationConfig struct {
	// Pagination controls page sizing for GraphQL calls.
	operations.Pagination
	// PayloadOptions controls payload inclusion.
	operations.PayloadOptions

	// Organization is the GitHub organization login to query.
	Organization types.TrimmedString `json:"organization" jsonschema:"description=GitHub organization login used to collect repositories."`
}

// runGitHubOrganizationReposOperation collects organization repositories through the GraphQL API.
func runGitHubOrganizationReposOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	config, err := operations.Decode[githubOrgRepoOperationConfig](input.Config)
	if err != nil {
		return types.OperationResult{}, err
	}

	if config.Organization == "" {
		return types.OperationResult{}, ErrOrganizationRequired
	}

	client, err := githubGraphQLClientForOperation(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	pageSize := clampPerPage(config.EffectivePageSize(maxPerPage))
	repositories, err := listGitHubOrganizationRepositories(ctx, client, config.Organization, pageSize)
	if err != nil {
		return operations.OperationFailure("GitHub organization repository collection failed", err, map[string]any{
			"organization": config.Organization.String(),
		})
	}

	sampleCount := min(len(repositories), operations.DefaultSampleSize)
	samples := make([]githubOrganizationRepository, sampleCount)
	copy(samples, repositories[:sampleCount])

	details := map[string]any{
		"organization": config.Organization.String(),
		"count":        len(repositories),
		"samples":      samples,
	}

	details = operations.AddPayloadIf(details, config.IncludePayloads, "repositories", repositories)

	return types.OperationResult{
		Status:  types.OperationStatusOK,
		Summary: fmt.Sprintf("Collected %d repositories for organization %s", len(repositories), config.Organization.String()),
		Details: details,
	}, nil
}

// githubOrganizationRepository stores normalized GraphQL repository metadata.
type githubOrganizationRepository struct {
	// ID is the repository node identifier.
	ID string
	// Name is the repository short name.
	Name string
	// NameWithOwner is the owner/name identifier.
	NameWithOwner string
	// Description is the repository description.
	Description string
	// URL is the repository URL.
	URL string
	// Visibility is the repository visibility enum value.
	Visibility string
	// IsPrivate indicates private visibility.
	IsPrivate bool
	// IsArchived indicates archived status.
	IsArchived bool
	// IsFork indicates fork status.
	IsFork bool
	// DefaultBranch is the default branch name when available.
	DefaultBranch string
	// PushedAt is the latest push timestamp.
	PushedAt time.Time
	// UpdatedAt is the latest update timestamp.
	UpdatedAt time.Time
}

// githubOrganizationRepoNode captures repository fields returned by GraphQL.
type githubOrganizationRepoNode struct {
	// ID is the node ID.
	ID string
	// Name is the short repository name.
	Name string
	// NameWithOwner is the full owner/name value.
	NameWithOwner string
	// Description is the repository description.
	Description string
	// URL is the repository URL.
	URL string
	// Visibility is the repository visibility enum.
	Visibility string
	// IsPrivate indicates private visibility.
	IsPrivate bool
	// IsArchived indicates archived status.
	IsArchived bool
	// IsFork indicates fork status.
	IsFork bool
	// UpdatedAt reports the latest update timestamp.
	UpdatedAt time.Time
	// PushedAt reports the latest push timestamp.
	PushedAt time.Time
	// DefaultBranchRef contains default branch metadata.
	DefaultBranchRef *struct {
		// Name is the default branch name.
		Name string
	} `graphql:"defaultBranchRef"`
}

// listGitHubOrganizationRepositories paginates repositories for a GitHub organization.
func listGitHubOrganizationRepositories(ctx context.Context, client *githubv4.Client, organization types.TrimmedString, pageSize int) ([]githubOrganizationRepository, error) {
	var (
		cursor       *githubv4.String
		repositories []githubOrganizationRepository
	)

	for {
		var query struct {
			Organization struct {
				Repositories struct {
					Nodes    []githubOrganizationRepoNode
					PageInfo struct {
						HasNextPage githubv4.Boolean
						EndCursor   githubv4.String
					}
				} `graphql:"repositories(first: $pageSize, after: $cursor, orderBy: {field: UPDATED_AT, direction: DESC})"`
			} `graphql:"organization(login: $organization)"`
		}

		graphqlPageSize := githubv4.Int(pageSize) // #nosec G115 -- pageSize is clamped to maxPerPage (100)

		variables := map[string]any{
			"organization": githubv4.String(organization.String()),
			"pageSize":     graphqlPageSize,
			"cursor":       cursor,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, err
		}

		for _, node := range query.Organization.Repositories.Nodes {
			defaultBranch := ""
			if node.DefaultBranchRef != nil {
				defaultBranch = node.DefaultBranchRef.Name
			}

			repositories = append(repositories, githubOrganizationRepository{
				ID:            node.ID,
				Name:          node.Name,
				NameWithOwner: node.NameWithOwner,
				Description:   node.Description,
				URL:           node.URL,
				Visibility:    node.Visibility,
				IsPrivate:     node.IsPrivate,
				IsArchived:    node.IsArchived,
				IsFork:        node.IsFork,
				DefaultBranch: defaultBranch,
				PushedAt:      node.PushedAt,
				UpdatedAt:     node.UpdatedAt,
			})
		}

		if !bool(query.Organization.Repositories.PageInfo.HasNextPage) {
			break
		}

		next := query.Organization.Repositories.PageInfo.EndCursor
		cursor = &next
	}

	return repositories, nil
}

// githubGraphQLClientForOperation returns a pooled GraphQL client or a token-derived fallback.
func githubGraphQLClientForOperation(input types.OperationInput) (*githubv4.Client, error) {
	client := githubGraphQLClientFromAny(input.Client)
	if client != nil {
		return client, nil
	}

	token, err := auth.OAuthTokenFromPayload(input.Credential)
	if err != nil {
		return nil, err
	}

	return newGitHubGraphQLClient(token), nil
}
