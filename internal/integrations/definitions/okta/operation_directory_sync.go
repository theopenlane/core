package okta

import (
	"context"
	"net/url"
	"strings"

	oktagosdk "github.com/okta/okta-sdk-golang/v5/okta"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const directoryDefaultPageSize = int32(200)

// directoryGroupRef is a lightweight identifier for the group side of a membership payload
type directoryGroupRef struct {
	// ID is the stable Okta group identifier
	ID string `json:"id,omitempty"`
}

// directoryMembershipPayload is the envelope payload for a single group membership record
type directoryMembershipPayload struct {
	// Group identifies the group the member belongs to
	Group directoryGroupRef `json:"group"`
	// Member is the group member record returned by the Okta API
	Member *oktagosdk.GroupMember `json:"member,omitempty"`
}

// DirectorySync collects Okta directory users, groups, and memberships for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.IngestWithClientRequest(
		OktaClient,
		func(ctx context.Context, request types.OperationRequest, c *oktagosdk.APIClient) ([]types.IngestPayloadSet, error) {
			var cfg UserInput
			if request.Integration != nil {
				_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
			}

			return d.Run(ctx, c, cfg)
		},
	)
}

// Run collects Okta directory users, groups, and memberships
func (DirectorySync) Run(ctx context.Context, c *oktagosdk.APIClient, cfg UserInput) ([]types.IngestPayloadSet, error) {
	users, err := listDirectoryUsers(ctx, c, cfg)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))
	includedUsers := make(map[string]struct{}, len(users))

	for _, user := range users {
		envelope, err := providerkit.MarshalEnvelope(directoryUserResource(&user), user, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)

		if id := user.GetId(); id != "" {
			includedUsers[id] = struct{}{}
		}
	}

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
	}

	if !cfg.EnableGroupSync {
		return payloadSets, nil
	}

	groups, err := listDirectoryGroups(ctx, c)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, group := range groups {
		envelope, err := providerkit.MarshalEnvelope(directoryGroupResource(&group), group, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)

		members, err := listGroupMembers(ctx, c, &group)
		if err != nil {
			return nil, err
		}

		for _, member := range members {
			if !isIncludedMember(&member, includedUsers) {
				continue
			}

			membershipPayload := directoryMembershipPayload{
				Group:  directoryGroupRef{ID: group.GetId()},
				Member: &member,
			}

			resource := directoryGroupResource(&group)
			if memberRef := directoryMemberResource(&member); memberRef != "" {
				resource = resource + ":" + memberRef
			}

			envelope, err := providerkit.MarshalEnvelope(resource, membershipPayload, ErrPayloadEncode)
			if err != nil {
				return nil, err
			}

			membershipEnvelopes = append(membershipEnvelopes, envelope)
		}
	}

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

	return payloadSets, nil
}

// listDirectoryUsers pages through Okta users using the resolved sync settings
func listDirectoryUsers(ctx context.Context, c *oktagosdk.APIClient, cfg UserInput) ([]oktagosdk.User, error) {
	users := make([]oktagosdk.User, 0)
	cursor := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Default to `status pr` so that DEPROVISIONED users are included in the result set;
		// Okta's list API omits DEPROVISIONED users unless a search expression is provided.
		search := cfg.Search
		if search == "" {
			search = "status pr"
		}

		req := c.UserAPI.ListUsers(ctx).Limit(directoryDefaultPageSize).Search(search)

		if cursor != "" {
			req = req.After(cursor)
		}

		batch, resp, err := req.Execute()
		if err != nil {
			return nil, ErrDirectoryUsersFetchFailed
		}

		users = append(users, batch...)

		cursor = nextPageCursor(resp)
		if cursor == "" {
			break
		}
	}

	return users, nil
}

// listDirectoryGroups pages through all Okta groups
func listDirectoryGroups(ctx context.Context, c *oktagosdk.APIClient) ([]oktagosdk.Group, error) {
	groups := make([]oktagosdk.Group, 0)
	cursor := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req := c.GroupAPI.ListGroups(ctx).Limit(directoryDefaultPageSize)

		if cursor != "" {
			req = req.After(cursor)
		}

		batch, resp, err := req.Execute()
		if err != nil {
			return nil, ErrDirectoryGroupsFetchFailed
		}

		groups = append(groups, batch...)

		cursor = nextPageCursor(resp)
		if cursor == "" {
			break
		}
	}

	return groups, nil
}

// listGroupMembers pages through all members of one Okta group
func listGroupMembers(ctx context.Context, c *oktagosdk.APIClient, group *oktagosdk.Group) ([]oktagosdk.GroupMember, error) {
	groupID := group.GetId()
	if groupID == "" {
		return nil, nil
	}

	members := make([]oktagosdk.GroupMember, 0)
	cursor := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req := c.GroupAPI.ListGroupUsers(ctx, groupID).Limit(directoryDefaultPageSize)

		if cursor != "" {
			req = req.After(cursor)
		}

		batch, resp, err := req.Execute()
		if err != nil {
			return nil, ErrDirectoryGroupMembersFetchFailed
		}

		members = append(members, batch...)

		cursor = nextPageCursor(resp)
		if cursor == "" {
			break
		}
	}

	return members, nil
}

// isIncludedMember reports whether the member was included in the account ingest set
func isIncludedMember(member *oktagosdk.GroupMember, includedUsers map[string]struct{}) bool {
	if member == nil || len(includedUsers) == 0 {
		return false
	}

	id := member.GetId()
	if id == "" {
		return false
	}

	_, ok := includedUsers[id]

	return ok
}

// directoryUserResource returns the best stable resource identifier for one user
func directoryUserResource(user *oktagosdk.User) string {
	if user == nil {
		return ""
	}

	if profile := user.Profile; profile != nil {
		if login := profile.GetLogin(); login != "" {
			return strings.ToLower(login)
		}
	}

	return user.GetId()
}

// directoryGroupResource returns the stable resource identifier for one group
func directoryGroupResource(group *oktagosdk.Group) string {
	if group == nil {
		return ""
	}

	return group.GetId()
}

// directoryMemberResource returns the best stable resource identifier for one group member
func directoryMemberResource(member *oktagosdk.GroupMember) string {
	if member == nil {
		return ""
	}

	if profile := member.Profile; profile != nil {
		if login := profile.GetLogin(); login != "" {
			return strings.ToLower(login)
		}
	}

	return member.GetId()
}

// nextPageCursor extracts the Okta cursor value from the Link response header
func nextPageCursor(resp *oktagosdk.APIResponse) string {
	if resp == nil {
		return ""
	}

	nextURL := resp.NextPage()
	if nextURL == "" {
		return ""
	}

	u, err := url.Parse(nextURL)
	if err != nil {
		return ""
	}

	return u.Query().Get("after")
}
