package azureentraid

import (
	"context"
	"strings"

	"github.com/samber/lo"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// directoryUserPayload is the JSON-serializable representation of one Entra ID user
type directoryUserPayload struct {
	// ID is the stable Azure object identifier
	ID string `json:"id,omitempty"`
	// DisplayName is the user's display name
	DisplayName string `json:"displayName,omitempty"`
	// Mail is the primary SMTP email address
	Mail string `json:"mail,omitempty"`
	// UserPrincipalName is the UPN, used as email fallback
	UserPrincipalName string `json:"userPrincipalName,omitempty"`
	// AccountEnabled indicates whether the account is enabled
	AccountEnabled bool `json:"accountEnabled"`
	// UserType is "Member" or "Guest"
	UserType string `json:"userType,omitempty"`
	// Department is the department the user belongs to
	Department string `json:"department,omitempty"`
	// GivenName is the user's given name
	GivenName string `json:"givenName,omitempty"`
	// Surname is the user's surname
	Surname string `json:"surname,omitempty"`
	// JobTitle is the user's job title
	JobTitle string `json:"jobTitle,omitempty"`
}

// directoryGroupPayload is the JSON-serializable representation of one Entra ID group
type directoryGroupPayload struct {
	// ID is the stable Azure object identifier
	ID string `json:"id,omitempty"`
	// DisplayName is the group's display name
	DisplayName string `json:"displayName,omitempty"`
	// Mail is the group email address when mail-enabled
	Mail string `json:"mail,omitempty"`
	// GroupTypes lists group type labels (e.g. "Unified" for Microsoft 365 groups)
	GroupTypes []string `json:"groupTypes,omitempty"`
	// SecurityEnabled indicates whether the group is a security group
	SecurityEnabled bool `json:"securityEnabled"`
	// MailEnabled indicates whether the group is mail-enabled
	MailEnabled bool `json:"mailEnabled"`
}

// directoryEntityRef is a lightweight reference to a directory user or group
type directoryEntityRef struct {
	// ID is the stable Azure object identifier
	ID string `json:"id,omitempty"`
	// Email is the primary email for the entity
	Email string `json:"email,omitempty"`
}

// directoryMembershipPayload is the envelope payload for one group membership record
type directoryMembershipPayload struct {
	// Group is the group side of the membership
	Group directoryEntityRef `json:"group"`
	// Member is the member side of the membership
	Member directoryEntityRef `json:"member"`
}

// DirectorySync collects Azure Entra ID directory users, groups, and memberships for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(
		EntraClient,
		func(ctx context.Context, request types.OperationRequest, c *msgraphsdk.GraphServiceClient) ([]types.IngestPayloadSet, error) {
			var cfg UserInput
			if request.Integration != nil {
				_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
			}

			return d.Run(ctx, c, cfg)
		},
	)
}

// Run collects Azure Entra ID directory users, groups, and memberships
func (DirectorySync) Run(ctx context.Context, c *msgraphsdk.GraphServiceClient, cfg UserInput) ([]types.IngestPayloadSet, error) {
	users, err := listEntraUsers(ctx, c)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))
	includedUsers := make(map[string]struct{}, len(users)*2)

	for _, user := range users {
		if !isEntraUserIncluded(user, cfg) {
			continue
		}

		payload := userToPayload(user)
		resource := entraUserResource(payload)

		envelope, err := providerkit.MarshalEnvelope(resource, payload, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)

		if payload.ID != "" {
			includedUsers[payload.ID] = struct{}{}
		}

		if payload.UserPrincipalName != "" {
			includedUsers[strings.ToLower(payload.UserPrincipalName)] = struct{}{}
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

	groups, err := listEntraGroups(ctx, c)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, group := range groups {
		payload := groupToPayload(group)
		resource := entraGroupResource(payload)

		envelope, err := providerkit.MarshalEnvelope(resource, payload, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)

		members, err := listEntraGroupUserMembers(ctx, c, payload.ID)
		if err != nil {
			return nil, err
		}

		groupRef := directoryEntityRef{ID: payload.ID, Email: payload.Mail}

		for _, member := range members {
			memberPayload := userToPayload(member)
			memberRef := directoryEntityRef{
				ID:    memberPayload.ID,
				Email: entraUserEmail(memberPayload),
			}

			if !isIncludedMember(memberRef, includedUsers) {
				continue
			}

			membershipResource := resource
			if memberRef.ID != "" {
				membershipResource = resource + ":" + memberRef.ID
			}

			membershipEnvelope, err := providerkit.MarshalEnvelope(membershipResource, directoryMembershipPayload{
				Group:  groupRef,
				Member: memberRef,
			}, ErrPayloadEncode)
			if err != nil {
				return nil, err
			}

			membershipEnvelopes = append(membershipEnvelopes, membershipEnvelope)
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

// odataPage is the minimal interface required from an OData collection response page
type odataPage[T any] interface {
	GetValue() []T
	GetOdataNextLink() *string
}

// paginateOData pages through all records from an OData-paginated endpoint
func paginateOData[T any, P odataPage[T]](
	ctx context.Context,
	fetchFirst func() (P, error),
	fetchNext func(nextLink string) (P, error),
	fetchErr error,
) ([]T, error) {
	result, err := fetchFirst()
	if err != nil {
		return nil, fetchErr
	}

	items := append([]T(nil), result.GetValue()...)

	for result.GetOdataNextLink() != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		result, err = fetchNext(*result.GetOdataNextLink())
		if err != nil {
			return nil, fetchErr
		}

		items = append(items, result.GetValue()...)
	}

	return items, nil
}

// listEntraUsers pages through all Azure Entra ID users
func listEntraUsers(ctx context.Context, c *msgraphsdk.GraphServiceClient) ([]models.Userable, error) {
	return paginateOData(
		ctx,
		func() (models.UserCollectionResponseable, error) { return c.Users().Get(ctx, nil) },
		func(nextLink string) (models.UserCollectionResponseable, error) {
			return c.Users().WithUrl(nextLink).Get(ctx, nil)
		},
		ErrUsersFetchFailed,
	)
}

// listEntraGroups pages through all Azure Entra ID groups
func listEntraGroups(ctx context.Context, c *msgraphsdk.GraphServiceClient) ([]models.Groupable, error) {
	return paginateOData(
		ctx,
		func() (models.GroupCollectionResponseable, error) { return c.Groups().Get(ctx, nil) },
		func(nextLink string) (models.GroupCollectionResponseable, error) {
			return c.Groups().WithUrl(nextLink).Get(ctx, nil)
		},
		ErrGroupsFetchFailed,
	)
}

// listEntraGroupUserMembers pages through user-type members for one group via the /microsoft.graph.user cast endpoint
func listEntraGroupUserMembers(ctx context.Context, c *msgraphsdk.GraphServiceClient, groupID string) ([]models.Userable, error) {
	if groupID == "" {
		return nil, nil
	}

	return paginateOData(
		ctx,
		func() (models.UserCollectionResponseable, error) {
			return c.Groups().ByGroupId(groupID).Members().GraphUser().Get(ctx, nil)
		},
		func(nextLink string) (models.UserCollectionResponseable, error) {
			return c.Groups().ByGroupId(groupID).Members().GraphUser().WithUrl(nextLink).Get(ctx, nil)
		},
		ErrMembersFetchFailed,
	)
}

// isEntraUserIncluded applies inclusion filters based on installation config
func isEntraUserIncluded(user models.Userable, cfg UserInput) bool {
	if !cfg.IncludeGuestUsers && strings.EqualFold(lo.FromPtr(user.GetUserType()), "Guest") {
		return false
	}

	return true
}

// isIncludedMember reports whether the member's ID or email appears in the included users set
func isIncludedMember(ref directoryEntityRef, includedUsers map[string]struct{}) bool {
	if len(includedUsers) == 0 {
		return false
	}

	if ref.ID != "" {
		if _, ok := includedUsers[ref.ID]; ok {
			return true
		}
	}

	if ref.Email != "" {
		_, ok := includedUsers[strings.ToLower(ref.Email)]
		return ok
	}

	return false
}

// userToPayload maps a Userable SDK model to a JSON-serializable payload struct
func userToPayload(user models.Userable) directoryUserPayload {
	return directoryUserPayload{
		ID:                lo.FromPtr(user.GetId()),
		DisplayName:       lo.FromPtr(user.GetDisplayName()),
		Mail:              lo.FromPtr(user.GetMail()),
		UserPrincipalName: lo.FromPtr(user.GetUserPrincipalName()),
		AccountEnabled:    lo.FromPtr(user.GetAccountEnabled()),
		UserType:          lo.FromPtr(user.GetUserType()),
		Department:        lo.FromPtr(user.GetDepartment()),
		GivenName:         lo.FromPtr(user.GetGivenName()),
		Surname:           lo.FromPtr(user.GetSurname()),
		JobTitle:          lo.FromPtr(user.GetJobTitle()),
	}
}

// groupToPayload maps a Groupable SDK model to a JSON-serializable payload struct
func groupToPayload(group models.Groupable) directoryGroupPayload {
	return directoryGroupPayload{
		ID:              lo.FromPtr(group.GetId()),
		DisplayName:     lo.FromPtr(group.GetDisplayName()),
		Mail:            lo.FromPtr(group.GetMail()),
		GroupTypes:      group.GetGroupTypes(),
		SecurityEnabled: lo.FromPtr(group.GetSecurityEnabled()),
		MailEnabled:     lo.FromPtr(group.GetMailEnabled()),
	}
}

// entraUserResource returns the stable resource identifier for one user payload
func entraUserResource(p directoryUserPayload) string {
	if p.ID != "" {
		return p.ID
	}

	return entraUserEmail(p)
}

// entraGroupResource returns the stable resource identifier for one group payload
func entraGroupResource(p directoryGroupPayload) string {
	if p.ID != "" {
		return p.ID
	}

	return p.Mail
}

// entraUserEmail returns the best email for one user payload
func entraUserEmail(p directoryUserPayload) string {
	if p.Mail != "" {
		return p.Mail
	}

	return p.UserPrincipalName
}
