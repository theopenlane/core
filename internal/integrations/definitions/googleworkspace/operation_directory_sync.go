package googleworkspace

import (
	"context"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// directoryDefaultPageSize is the number of records to request per page when listing users, groups, and members
const directoryDefaultPageSize = int64(200)

// DirectorySync collects Google Workspace directory users for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(workspaceClient, func(ctx context.Context, request types.OperationRequest, svc *admin.Service) ([]types.IngestPayloadSet, error) {
		var meta InstallationMetadata

		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.InstallationMetadata.Attributes, &meta)
		}

		if meta.CustomerID == "" {
			return nil, ErrCustomerIDMissing
		}

		return d.Run(ctx, svc, meta.CustomerID)
	})
}

// Run collects Google Workspace directory users, groups, and memberships
func (DirectorySync) Run(ctx context.Context, svc *admin.Service, customerID string) ([]types.IngestPayloadSet, error) {
	users, err := listDirectoryUsers(ctx, svc, customerID)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))
	includedUsers := make(map[string]struct{}, len(users))

	for _, user := range users {
		envelope, err := providerkit.MarshalEnvelope(user.PrimaryEmail, user, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)

		includedUsers[user.PrimaryEmail] = struct{}{}
	}

	groups, err := listDirectoryGroups(ctx, svc, customerID)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, group := range groups {
		groupResource := group.Email

		envelope, err := providerkit.MarshalEnvelope(groupResource, group, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)

		members, err := listGroupMembers(ctx, svc, group)
		if err != nil {
			return nil, err
		}

		for _, member := range members {
			if !isIncludedUserMember(member, includedUsers) {
				continue
			}

			envelope, err := providerkit.MarshalEnvelope(groupResource, member, ErrPayloadEncode)
			if err != nil {
				return nil, err
			}

			membershipEnvelopes = append(membershipEnvelopes, envelope)
		}
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			Envelopes: groupEnvelopes,
		},
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
			Envelopes: membershipEnvelopes,
		},
	}, nil
}

// listDirectoryUsers pages through all Google Workspace users for the given customer
func listDirectoryUsers(ctx context.Context, svc *admin.Service, customerID string) ([]*admin.User, error) {
	var users []*admin.User

	pageToken := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Users.List().
			Customer(customerID).
			MaxResults(directoryDefaultPageSize).
			Projection("full").
			ViewType("admin_view").
			OrderBy("email").
			Context(ctx)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, ErrDirectoryUsersFetchFailed
		}

		users = append(users, resp.Users...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return users, nil
}

// listDirectoryGroups pages through all Google Workspace groups for the given customer
func listDirectoryGroups(ctx context.Context, svc *admin.Service, customerID string) ([]*admin.Group, error) {
	var groups []*admin.Group

	pageToken := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Groups.List().
			Customer(customerID).
			MaxResults(directoryDefaultPageSize).
			OrderBy("email").
			Context(ctx)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, ErrDirectoryGroupsFetchFailed
		}

		groups = append(groups, resp.Groups...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return groups, nil
}

// listGroupMembers pages through all members of one Google Workspace group
func listGroupMembers(ctx context.Context, svc *admin.Service, group *admin.Group) ([]*admin.Member, error) {
	var members []*admin.Member

	pageToken := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Members.List(group.Id).
			MaxResults(directoryDefaultPageSize).
			Context(ctx)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, ErrDirectoryGroupMembersFetchFailed
		}

		members = append(members, resp.Members...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return members, nil
}

// isIncludedUserMember reports whether a group member is a user that was included in the account ingest set
func isIncludedUserMember(member *admin.Member, includedUsers map[string]struct{}) bool {
	if member.Type != "" && member.Type != "USER" {
		return false
	}

	_, ok := includedUsers[member.Email]
	return ok
}
