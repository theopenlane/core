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

// orgMemberNodeGQL is a single organization member record returned by the GitHub GraphQL API
type orgMemberNodeGQL struct {
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
}

// orgMemberNode combines the github response with details that are populated after the query
type orgMemberNode struct {
	orgMemberNodeGQL

	// Org is the organization login, populated after query
	Org string
	// CanonicalEmail is the best-resolved email, populated after query
	CanonicalEmail string
	// GivenName is the user's first name from SAML identity when available
	GivenName string
	// FamilyName is the user's last name from SAML identity when available
	FamilyName string
}

// teamNodeGQL is a single GitHub team returned by the teams graphql query
type teamNodeGQL struct {
	// DatabaseID is the numeric GitHub team identifier
	DatabaseID int `graphql:"databaseId"`
	// Name is the team display name
	Name string
	// Slug is the team slug within the organization
	Slug string
	// Description is the team description
	Description string
	// Privacy is the GitHub team privacy value
	Privacy string
	// Members are the users assigned to the team
	Members struct {
		Nodes    []teamMemberNodeGQL
		PageInfo pageInfo
	} `graphql:"members(first: $memberFirst)"`
}

// teamNode combines the GitHub team response with org context
type teamNode struct {
	teamNodeGQL

	// Org is the organization login
	Org string
}

// teamMemberNodeGQL is a GitHub team member reference
type teamMemberNodeGQL struct {
	// DatabaseID is the numeric GitHub user identifier
	DatabaseID int `graphql:"databaseId"`
	// Login is the GitHub username
	Login string
}

// teamMembershipNode represents one GitHub team membership edge for ingest
type teamMembershipNode struct {
	// Org is the organization login
	Org string
	// Team is the GitHub team
	Team teamNode
	// Member is the GitHub user in the team
	Member teamMemberNodeGQL
	// Role is the normalized membership role
	Role string
}

// orgNode holds the login of a GitHub organization accessible to the installation
type orgNode struct {
	// Login is the organization's GitHub username
	Login string
}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(gitHubClient, func(ctx context.Context, _ types.OperationRequest, client GraphQLClient) ([]types.IngestPayloadSet, error) {
		return d.Run(ctx, client)
	})
}

// Run collects GitHub organization members, teams, and team memberships for directory ingest
func (d DirectorySync) Run(ctx context.Context, client GraphQLClient) ([]types.IngestPayloadSet, error) {
	orgs, err := queryViewerOrganizations(ctx, client)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("githubapp_directorysync: failed to discover organizations")
		return nil, err
	}

	userAccountEnvelopes := make([]types.MappingEnvelope, 0)
	groupEnvelopes := make([]types.MappingEnvelope, 0)
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

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

		orgMembers := make([]orgMemberNode, len(members))
		for i, m := range members {
			orgMembers[i] = orgMemberNode{
				orgMemberNodeGQL: m,
			}

			orgMembers[i].Org = org.Login
			resolveCanonicalEmail(&orgMembers[i], samlMap)

			resource := fmt.Sprintf("%s/%s", org.Login, orgMembers[i].Login)

			envelope, err := providerkit.MarshalEnvelope(resource, orgMembers[i], ErrIngestPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("member", orgMembers[i].Login).Msg("githubapp_directorysync: failed to marshal member into ingest envelope")
				return nil, err
			}

			userAccountEnvelopes = append(userAccountEnvelopes, envelope)
		}

		logx.FromContext(ctx).Info().Str("org", org.Login).Int("member_count", len(orgMembers)).Bool("saml_available", samlMap != nil).Msg("githubapp_directorysync: queried organization members")

		if d.DisableGroupSync {
			continue
		}

		teams, memberships, err := queryOrganizationTeams(ctx, client, org.Login)
		if err != nil {
			logx.FromContext(ctx).Warn().Err(err).Str("org", org.Login).Msg("githubapp_directorysync: team query failed, continuing without group data")
			continue
		}

		for _, team := range teams {
			resource := fmt.Sprintf("%s/%s", org.Login, team.Slug)

			envelope, err := providerkit.MarshalEnvelope(resource, team, ErrIngestPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("team", team.Slug).Msg("githubapp_directorysync: failed to marshal team into ingest envelope")
				return nil, err
			}

			groupEnvelopes = append(groupEnvelopes, envelope)
		}

		for _, membership := range memberships {
			resource := fmt.Sprintf("%s/%s", org.Login, membership.Team.Slug)

			envelope, err := providerkit.MarshalEnvelope(resource, membership, ErrIngestPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("team", membership.Team.Slug).Str("member", membership.Member.Login).Msg("githubapp_directorysync: failed to marshal team membership into ingest envelope")
				return nil, err
			}

			membershipEnvelopes = append(membershipEnvelopes, envelope)
		}

		logx.FromContext(ctx).Info().Str("org", org.Login).Int("team_count", len(teams)).Int("membership_count", len(memberships)).Msg("githubapp_directorysync: queried organization teams")
	}

	logx.FromContext(ctx).Info().
		Int("org_count", len(orgs)).
		Int("member_count", len(userAccountEnvelopes)).
		Int("team_count", len(groupEnvelopes)).
		Int("membership_count", len(membershipEnvelopes)).
		Bool("group_sync_disabled", d.DisableGroupSync).
		Msg("githubapp_directorysync: collected directory records")

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: userAccountEnvelopes,
		},
	}

	if !d.DisableGroupSync {
		payloadSets = append(payloadSets,
			types.IngestPayloadSet{
				Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
				Envelopes: groupEnvelopes,
			},
			types.IngestPayloadSet{
				Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
				Envelopes: membershipEnvelopes,
			},
		)
	}

	return payloadSets, nil
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
func queryOrganizationMembers(ctx context.Context, client GraphQLClient, orgLogin string) ([]orgMemberNodeGQL, error) {
	members := make([]orgMemberNodeGQL, 0)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("githubapp: error checking context for github member query")
			return nil, err
		}

		var query struct {
			Organization struct {
				MembersWithRole struct {
					Nodes    []orgMemberNodeGQL
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
			logx.FromContext(ctx).Error().Err(err).Interface("org_member_node", members).Interface("variables", variables).Msg("githubapp: error querying github organization members")

			return nil, fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}

		members = append(members, query.Organization.MembersWithRole.Nodes...)

		if !query.Organization.MembersWithRole.PageInfo.HasNextPage {
			break
		}

		cursor := githubv4.String(query.Organization.MembersWithRole.PageInfo.EndCursor)

		after = &cursor
	}

	return members, nil
}

// queryOrganizationTeams lists teams and team memberships for the provided organization
func queryOrganizationTeams(ctx context.Context, client GraphQLClient, orgLogin string) ([]teamNode, []teamMembershipNode, error) {
	teams := make([]teamNode, 0)
	memberships := make([]teamMembershipNode, 0)
	var after *githubv4.String

	for {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}

		var query struct {
			Organization struct {
				Teams struct {
					Nodes    []teamNodeGQL
					PageInfo pageInfo
				} `graphql:"teams(first: $first, after: $after)"`
			} `graphql:"organization(login: $login)"`
		}

		variables := map[string]any{
			"login":       githubv4.String(orgLogin),
			"first":       githubv4.Int(defaultPageSize),
			"after":       after,
			"memberFirst": githubv4.Int(maxPageSize),
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, nil, fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}

		for _, team := range query.Organization.Teams.Nodes {
			members := team.Members.Nodes
			if team.Members.PageInfo.HasNextPage && team.Slug != "" {
				remaining, err := queryOrganizationTeamMembers(ctx, client, orgLogin, team.Slug, team.Members.PageInfo.EndCursor)
				if err != nil {
					return nil, nil, err
				}

				members = append(members, remaining...)
			}

			team := teamNode{
				teamNodeGQL: team,
				Org:         orgLogin,
			}
			teams = append(teams, team)

			for _, member := range members {
				memberships = append(memberships, teamMembershipNode{
					Org:    orgLogin,
					Team:   team,
					Member: member,
					Role:   "MEMBER",
				})
			}
		}

		if !query.Organization.Teams.PageInfo.HasNextPage || query.Organization.Teams.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Organization.Teams.PageInfo.EndCursor))
	}

	return teams, memberships, nil
}

// queryOrganizationTeamMembers fetches additional member pages for the requested team
func queryOrganizationTeamMembers(ctx context.Context, client GraphQLClient, orgLogin string, teamSlug string, firstAfter string) ([]teamMemberNodeGQL, error) {
	members := make([]teamMemberNodeGQL, 0)
	after := githubv4.NewString(githubv4.String(firstAfter))

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var query struct {
			Organization struct {
				Team struct {
					Members struct {
						Nodes    []teamMemberNodeGQL
						PageInfo pageInfo
					} `graphql:"members(first: $first, after: $after)"`
				} `graphql:"team(slug: $slug)"`
			} `graphql:"organization(login: $login)"`
		}

		variables := map[string]any{
			"login": githubv4.String(orgLogin),
			"slug":  githubv4.String(teamSlug),
			"first": githubv4.Int(maxPageSize),
			"after": after,
		}

		if err := client.Query(ctx, &query, variables); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrAPIRequest, err)
		}

		members = append(members, query.Organization.Team.Members.Nodes...)

		if !query.Organization.Team.Members.PageInfo.HasNextPage || query.Organization.Team.Members.PageInfo.EndCursor == "" {
			break
		}

		after = githubv4.NewString(githubv4.String(query.Organization.Team.Members.PageInfo.EndCursor))
	}

	return members, nil
}
