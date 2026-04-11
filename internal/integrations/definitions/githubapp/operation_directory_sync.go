package githubapp

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// samlIdentity holds resolved SAML identity data for one organization member
type samlIdentity struct {
	// NameID is the SAML nameId, typically the user's SSO email
	NameID string
	// GivenName is the user's first name from the SAML assertion
	GivenName string
	// FamilyName is the user's last name from the SAML assertion
	FamilyName string
}

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
	// OrganizationVerifiedDomainEmails is the list of emails matching the org's verified domains
	OrganizationVerifiedDomainEmails []string `graphql:"organizationVerifiedDomainEmails(login: $login)"`
	// Org is the organization login, populated after query
	Org string `graphql:"-"`
	// CanonicalEmail is the best-resolved email, populated after query
	CanonicalEmail string `graphql:"-"`
	// GivenName is the user's first name from SAML identity when available
	GivenName string `graphql:"-"`
	// FamilyName is the user's last name from SAML identity when available
	FamilyName string `graphql:"-"`
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
		logx.FromContext(ctx).Error().Err(err).Msg("githubapp_directorysync: failed to discover organizations")
		return nil, err
	}

	envelopes := make([]types.MappingEnvelope, 0)

	for _, org := range orgs {
		samlMap, err := queryExternalIdentities(ctx, client, org.Login)
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("org", org.Login).Msg("githubapp_directorysync: SAML identity query failed, continuing without SSO data")
			samlMap = nil
		}

		members, err := queryOrganizationMembers(ctx, client, org.Login)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("org", org.Login).Msg("githubapp_directorysync: failed to query members")
			return nil, err
		}

		for i := range members {
			members[i].Org = org.Login
			resolveCanonicalEmail(&members[i], samlMap)

			resource := fmt.Sprintf("%s/%s", org.Login, members[i].Login)

			envelope, err := providerkit.MarshalEnvelope(resource, members[i], ErrIngestPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("member", members[i].Login).Msg("githubapp_directorysync: failed to marshal member into ingest envelope")
				return nil, err
			}

			envelopes = append(envelopes, envelope)
		}

		logx.FromContext(ctx).Info().Str("org", org.Login).Int("member_count", len(members)).Bool("saml_available", samlMap != nil).Msg("githubapp_directorysync: queried organization members")
	}

	logx.FromContext(ctx).Info().Int("org_count", len(orgs)).Int("member_count", len(envelopes)).Msg("githubapp_directorysync: collected organization members for directory sync")

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: envelopes,
		},
	}, nil
}

// resolveCanonicalEmail sets the best email for a member using the priority chain:
// SAML nameId > organization verified domain email > public profile email
func resolveCanonicalEmail(member *orgMemberNode, samlMap map[string]samlIdentity) {
	if samlMap != nil {
		if saml, ok := samlMap[member.Login]; ok {
			member.CanonicalEmail = saml.NameID
			member.GivenName = saml.GivenName
			member.FamilyName = saml.FamilyName
		}
	}

	if member.CanonicalEmail == "" && len(member.OrganizationVerifiedDomainEmails) > 0 {
		member.CanonicalEmail = member.OrganizationVerifiedDomainEmails[0]
	}

	if member.CanonicalEmail == "" {
		member.CanonicalEmail = member.Email
	}
}

// queryViewerOrganizations discovers organizations accessible to the GitHub App installation
// by extracting unique organization owners from the installation's accessible repositories.
// The viewer.organizations query returns empty for GitHub App bot users because the bot
// is not an organization member.
func queryViewerOrganizations(ctx context.Context, client GraphQLClient) ([]orgNode, error) {
	seen := make(map[string]struct{})
	var orgs []orgNode
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Viewer struct {
				Repositories struct {
					Nodes []struct {
						Owner struct {
							Login    string
							TypeName string `graphql:"__typename"`
						}
					}
					PageInfo pageInfo
				} `graphql:"repositories(first: $first, after: $after)"`
			}
		}

		variables := map[string]any{
			"first": githubv4.Int(maxPageSize),
			"after": after,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}

		for _, r := range query.Viewer.Repositories.Nodes {
			if r.Owner.TypeName != "Organization" {
				continue
			}

			if _, ok := seen[r.Owner.Login]; ok {
				continue
			}

			seen[r.Owner.Login] = struct{}{}
			orgs = append(orgs, orgNode{Login: r.Owner.Login})
		}

		if !query.Viewer.Repositories.PageInfo.HasNextPage || query.Viewer.Repositories.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Viewer.Repositories.PageInfo.EndCursor))
	}

	return orgs, nil
}

// queryExternalIdentities queries SAML external identities for an organization.
// Returns nil, nil when the organization has no SAML identity provider configured.
func queryExternalIdentities(ctx context.Context, client GraphQLClient, orgLogin string) (map[string]samlIdentity, error) {
	identities := make(map[string]samlIdentity)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Organization struct {
				SamlIdentityProvider *struct {
					ExternalIdentities struct {
						Nodes []struct {
							SamlIdentity *struct {
								NameID     string `graphql:"nameId"`
								GivenName  string
								FamilyName string
							}
							User *struct {
								Login string
							}
						}
						PageInfo pageInfo
					} `graphql:"externalIdentities(first: $first, after: $after)"`
				}
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

		if query.Organization.SamlIdentityProvider == nil {
			return nil, nil
		}

		for _, node := range query.Organization.SamlIdentityProvider.ExternalIdentities.Nodes {
			if node.User == nil || node.SamlIdentity == nil {
				continue
			}

			identities[node.User.Login] = samlIdentity{
				NameID:     node.SamlIdentity.NameID,
				GivenName:  node.SamlIdentity.GivenName,
				FamilyName: node.SamlIdentity.FamilyName,
			}
		}

		pi := query.Organization.SamlIdentityProvider.ExternalIdentities.PageInfo
		if !pi.HasNextPage || pi.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(pi.EndCursor))
	}

	return identities, nil
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
