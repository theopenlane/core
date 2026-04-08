package githubapp

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// orgMemberNode is a single organization member record returned by the GitHub GraphQL API
type orgMemberNode struct {
	// DatabaseID is the numeric GitHub user identifier
	DatabaseID int `graphql:"databaseId"`
	// Login is the GitHub username
	Login string
	// Name is the user's display name
	Name string
	// Email is the user's public email address
	Email string
	// AvatarURL is the user's avatar URL
	AvatarURL string `graphql:"avatarUrl"`
	// Org is the organization login populated after query
	Org string `graphql:"-"`
}

// orgNode holds the login of a GitHub organization accessible to the installation
type orgNode struct {
	// Login is the organization's GitHub username
	Login string
}

// DirectorySync collects GitHub organization members for directory account ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(gitHubClient, func(ctx context.Context, _ types.OperationRequest, client GraphQLClient) ([]types.IngestPayloadSet, error) {
		return d.Run(ctx, client)
	})
}

// Run collects GitHub organization members and emits directory account ingest payloads
func (DirectorySync) Run(ctx context.Context, client GraphQLClient) ([]types.IngestPayloadSet, error) {
	orgs, err := queryViewerOrganizations(ctx, client)
	if err != nil {
		return nil, err
	}

	envelopes := make([]types.MappingEnvelope, 0)

	for _, org := range orgs {
		members, err := queryOrganizationMembers(ctx, client, org.Login)
		if err != nil {
			return nil, err
		}

		for _, member := range members {
			member.Org = org.Login

			resource := fmt.Sprintf("%s/%s", org.Login, member.Login)

			envelope, err := providerkit.MarshalEnvelope(resource, member, ErrIngestPayloadEncode)
			if err != nil {
				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: envelopes,
		},
	}, nil
}

// queryViewerOrganizations lists organizations accessible to the GitHub App installation
func queryViewerOrganizations(ctx context.Context, client GraphQLClient) ([]orgNode, error) {
	orgs := make([]orgNode, 0)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Viewer struct {
				Organizations struct {
					Nodes    []orgNode
					PageInfo pageInfo
				} `graphql:"organizations(first: $first, after: $after)"`
			}
		}

		variables := map[string]any{
			"first": githubv4.Int(defaultPageSize),
			"after": after,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}

		orgs = append(orgs, query.Viewer.Organizations.Nodes...)

		if !query.Viewer.Organizations.PageInfo.HasNextPage || query.Viewer.Organizations.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Viewer.Organizations.PageInfo.EndCursor))
	}

	return orgs, nil
}

// queryOrganizationMembers lists members of one GitHub organization
func queryOrganizationMembers(ctx context.Context, client GraphQLClient, orgLogin string) ([]orgMemberNode, error) {
	members := make([]orgMemberNode, 0)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Organization struct {
				MembersWithRole struct {
					Nodes    []orgMemberNode
					PageInfo pageInfo
				} `graphql:"membersWithRole(first: $first, after: $after)"`
			} `graphql:"organization(login: $login)"`
		}

		variables := map[string]any{
			"login": githubv4.String(orgLogin),
			"first": githubv4.Int(defaultPageSize),
			"after": after,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}

		members = append(members, query.Organization.MembersWithRole.Nodes...)

		if !query.Organization.MembersWithRole.PageInfo.HasNextPage || query.Organization.MembersWithRole.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Organization.MembersWithRole.PageInfo.EndCursor))
	}

	return members, nil
}
