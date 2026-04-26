package cloudflare

import (
	"context"
	"time"

	cf "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/accounts"
	"github.com/cloudflare/cloudflare-go/v6/iam"
	"github.com/cloudflare/cloudflare-go/v6/packages/pagination"
	"github.com/cloudflare/cloudflare-go/v6/shared"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/logx"
)

// cloudflareMemberPayload is the normalized payload for a Cloudflare account member
type cloudflareMemberPayload struct {
	// ID is the Cloudflare membership identifier
	ID string `json:"id,omitempty"`
	// Email is the contact email of the member
	Email string `json:"email,omitempty"`
	// Status is the membership status (accepted, pending)
	Status string `json:"status,omitempty"`
	// UserID is the Cloudflare user identifier
	UserID string `json:"user_id,omitempty"`
	// FirstName is the user's first name
	FirstName string `json:"first_name,omitempty"`
	// LastName is the user's last name
	LastName string `json:"last_name,omitempty"`
	// TwoFactorEnabled indicates whether 2FA is enabled for the user
	TwoFactorEnabled bool `json:"two_factor_enabled,omitempty"`
	// Payload is the full profile from the source
	Payload any `json:"payload"`
	// AccountID is the account id from the source
	AccountID string `json:"account_id"`
	// Memberships are the groups the user is a member of
	Memberships []cloudflareGroupMemberPayload `json:"memberships,omitempty"`
}

// cloudflareGroupPayload is the normalized payload for a Cloudflare groups (roles, user groups, permission groups)
type cloudflareGroupPayload struct {
	// ID is the Cloudflare group identifier
	ID string `json:"id,omitempty"`
	// Name is the name of the group
	Name string `json:"name,omitempty"`
	// Description is the description of the group
	Description string `json:"description,omitempty"`
	// LastModifiedTime is the last update timestamp of the group from the source
	LastModifiedTime time.Time `json:"last_modified_time,omitempty"`
	// CreatedTime is the timestamp the group was created from the source
	CreatedTime time.Time `json:"created_time,omitempty"`
	// Payload is the full profile from the source
	Payload any `json:"payload"`
}

// cloudflareGroupMemberPayload is the normalized payload for a Cloudflare group member payload
type cloudflareGroupMemberPayload struct {
	// GroupID is the Cloudflare group identifier
	GroupID string `json:"group_id"`
	// UserID is the clouldflare user id
	UserID string `json:"user_id"`
	// Meta is metadata from the policy group
	Meta any `json:"meta,omitempty"`
	// Payload is the full profile from the source
	Payload any `json:"payload"`
}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(cloudflareClient, func(ctx context.Context, request types.OperationRequest, client *cf.Client) ([]types.IngestPayloadSet, error) {
		var cfg DirectorySync
		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
		}

		return d.Run(ctx, request.Credentials, client, cfg)
	})
}

// Run collects Cloudflare account members and emits directory account ingest payloads
func (DirectorySync) Run(ctx context.Context, credentials types.CredentialBindings, client *cf.Client, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
	meta, err := resolveCredential(credentials)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: error attempting to resolve credentials")
		return nil, err
	}

	if meta.AccountID == "" {
		return nil, ErrAccountIDMissing
	}

	members, err := listDirectoryUsers(ctx, client, meta.AccountID)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0)
	includedUsers := make(map[string]struct{}, len(members))

	for _, m := range members {
		resource := meta.AccountID + "/" + m.Email

		envelope, err := providerkit.MarshalEnvelope(resource, m, ErrPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("cloudflare_directory_sync: error during cloudflare account members iteration")

			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)
		includedUsers[m.Email] = struct{}{}
	}

	groups, err := listDirectoryGroups(ctx, client, meta.AccountID)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, group := range groups {
		resource := meta.AccountID + "/" + group.ID

		envelope, err := providerkit.MarshalEnvelope(resource, group, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)
	}

	for _, member := range members {
		if !isIncludedUserMember(member, includedUsers) {
			continue
		}

		for _, membership := range member.Memberships {
			resource := meta.AccountID + "/" + membership.GroupID + ":" + membership.UserID
			envelope, err := providerkit.MarshalEnvelope(resource, membership, ErrPayloadEncode)
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

// isIncludedUserMember reports whether a group member is a user that was included in the account ingest set
func isIncludedUserMember(member cloudflareMemberPayload, includedUsers map[string]struct{}) bool {
	_, ok := includedUsers[member.Email]
	return ok
}

// listDirectoryUsers pages through all Cloudflare members for the given account
func listDirectoryUsers(ctx context.Context, client *cf.Client, accountID string) ([]cloudflareMemberPayload, error) {
	iter := client.Accounts.Members.ListAutoPaging(ctx, accounts.MemberListParams{
		AccountID: cf.F(accountID),
	})

	members, err := parseResponseData(ctx, iter, func(m shared.Member) cloudflareMemberPayload {
		memberships := make([]cloudflareGroupMemberPayload, 0)
		for _, policy := range m.Policies {
			for _, g := range policy.PermissionGroups {
				memberships = append(memberships, cloudflareGroupMemberPayload{
					GroupID: g.ID,
					UserID:  m.User.ID,
					Meta:    g.Meta,
					Payload: g,
				})
			}
		}
		return cloudflareMemberPayload{
			ID:               m.ID,
			Email:            m.User.Email,
			Status:           string(m.Status),
			UserID:           m.User.ID,
			FirstName:        m.User.FirstName,
			LastName:         m.User.LastName,
			TwoFactorEnabled: m.User.TwoFactorAuthenticationEnabled,
			Payload:          m,
			Memberships:      memberships,
			AccountID:        accountID,
		}
	})

	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("cloudflare: error pulling members")
		return nil, ErrMembersFetchFailed
	}

	return members, nil
}

// listDirectoryGroups pages through all Cloudflare roles and groups for the given account
func listDirectoryGroups(ctx context.Context, client *cf.Client, accountID string) ([]cloudflareGroupPayload, error) {
	iterUser := client.IAM.UserGroups.ListAutoPaging(ctx, iam.UserGroupListParams{
		AccountID: cf.F(accountID),
	})

	groups, err := parseResponseData(ctx, iterUser, func(g iam.UserGroupListResponse) cloudflareGroupPayload {
		return cloudflareGroupPayload{ID: g.ID, Name: g.Name, LastModifiedTime: g.ModifiedOn, CreatedTime: g.CreatedOn, Payload: g}
	})
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("cloudflare: error pulling user groups")
		return nil, ErrGroupsFetchFailed
	}

	iterPerm := client.IAM.PermissionGroups.ListAutoPaging(ctx, iam.PermissionGroupListParams{
		AccountID: cf.F(accountID),
	})

	permissionGroups, err := parseResponseData(ctx, iterPerm, func(g iam.PermissionGroupListResponse) cloudflareGroupPayload {
		return cloudflareGroupPayload{ID: g.ID, Name: g.Name, Payload: g}
	})

	groups = append(groups, permissionGroups...)

	iterRoles := client.Accounts.Roles.ListAutoPaging(ctx, accounts.RoleListParams{
		AccountID: cf.F(accountID),
	})

	roles, err := parseResponseData(ctx, iterRoles, func(g shared.Role) cloudflareGroupPayload {
		return cloudflareGroupPayload{ID: g.ID, Name: g.Name, Description: g.Description, Payload: g}
	})

	groups = append(groups, roles...)

	return groups, nil
}

// parseResponseData is a generic helper for iterating through results
func parseResponseData[T any, K any](
	ctx context.Context,
	iter *pagination.V4PagePaginationArrayAutoPager[T],
	mapFn func(T) K,
) ([]K, error) {
	var results []K

	for iter.Next() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		results = append(results, mapFn(iter.Current()))
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
