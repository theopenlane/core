package googleworkspace

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/samber/lo"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

const (
	directoryDefaultPageSize = int64(200)
	userDirectoryFields      = "nextPageToken,users(id,primaryEmail,name/fullName,name/givenName,name/familyName,orgUnitPath,suspended,archived,isEnforcedIn2Sv,isEnrolledIn2Sv,lastLoginTime,creationTime,deletionTime,customerId)"
	groupDirectoryFields     = "nextPageToken,groups(id,email,name,description,directMembersCount,adminCreated,etag)"
	memberDirectoryFields    = "nextPageToken,members(id,email,role,type,status,delivery_settings)"
)

// DirectorySyncConfig controls the directory sync operation
type DirectorySyncConfig struct {
	// Customer scopes the listing to a specific customer identifier
	Customer string `json:"customer,omitempty" jsonschema:"title=Customer ID"`
	// Domain scopes the listing to a specific domain
	Domain string `json:"domain,omitempty" jsonschema:"title=Domain"`
	// Query is a search query to filter users and groups server-side
	Query string `json:"query,omitempty" jsonschema:"title=Directory Query"`
	// OrganizationalUnit filters collected users to one org unit path and its descendants
	OrganizationalUnit string `json:"organizationalUnitPath,omitempty" jsonschema:"title=Organizational Unit Path"`
	// IncludeSuspended overrides whether suspended users are emitted
	IncludeSuspended *bool `json:"includeSuspendedUsers,omitempty" jsonschema:"title=Include Suspended Users"`
	// IncludeGroups overrides whether groups and memberships are emitted
	IncludeGroups *bool `json:"includeGroups,omitempty" jsonschema:"title=Sync Groups"`
}

type directoryEntityRef struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
}

type directoryMembershipPayload struct {
	Group  directoryEntityRef `json:"group"`
	Member *admin.Member      `json:"member,omitempty"`
}

// DirectorySync collects Google Workspace directory users for ingest
type DirectorySync struct{}

// Handle adapts directory sync to the generic operation registration boundary
func (d DirectorySync) Handle(client Client) types.OperationHandler {
	return func(ctx context.Context, request types.OperationRequest) (json.RawMessage, error) {
		svc, err := client.FromAny(request.Client)
		if err != nil {
			return nil, err
		}

		var cfg DirectorySyncConfig
		if err := jsonx.UnmarshalIfPresent(request.Config, &cfg); err != nil {
			return nil, ErrDirectorySyncConfigInvalid
		}

		return d.Run(ctx, svc, resolveDirectorySyncConfig(request.Integration, cfg))
	}
}

// Run collects Google Workspace directory users, groups, and memberships
func (DirectorySync) Run(ctx context.Context, svc *admin.Service, cfg DirectorySyncConfig) (json.RawMessage, error) {
	users, err := listDirectoryUsers(ctx, svc, cfg)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))
	includedUsers := make(map[string]struct{}, len(users)*2)

	for _, user := range users {
		if !isUserIncluded(user, cfg) {
			continue
		}

		envelope, err := providerkit.MarshalEnvelope(directoryUserResource(user), user, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)

		if user.Id != "" {
			includedUsers[user.Id] = struct{}{}
		}

		if user.PrimaryEmail != "" {
			includedUsers[strings.ToLower(user.PrimaryEmail)] = struct{}{}
		}
	}

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
	}

	if !lo.FromPtrOr(cfg.IncludeGroups, false) {
		return providerkit.EncodeResult(payloadSets, ErrResultEncode)
	}

	groups, err := listDirectoryGroups(ctx, svc, cfg)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, group := range groups {
		envelope, err := providerkit.MarshalEnvelope(directoryGroupResource(group), group, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)

		members, err := listGroupMembers(ctx, svc, group)
		if err != nil {
			return nil, err
		}

		for _, member := range members {
			if !isDirectoryUserMember(member) || !isIncludedMembershipMember(member, includedUsers) {
				continue
			}

			membershipPayload := directoryMembershipPayload{
				Group: directoryEntityRef{
					ID:    group.Id,
					Email: group.Email,
				},
				Member: member,
			}

			resource := directoryGroupResource(group)
			if memberRef := directoryMemberResource(member); memberRef != "" {
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

	return providerkit.EncodeResult(payloadSets, ErrResultEncode)
}

// resolveDirectorySyncConfig merges stored installation defaults with per-run overrides
func resolveDirectorySyncConfig(installation *generated.Integration, cfg DirectorySyncConfig) DirectorySyncConfig {
	var defaults UserInput
	if installation != nil {
		_ = jsonx.UnmarshalIfPresent(installation.Config.ClientConfig, &defaults)
	}

	settings := DirectorySyncConfig{
		Customer:           firstNonEmpty(cfg.Customer, defaults.CustomerID),
		Domain:             cfg.Domain,
		Query:              cfg.Query,
		OrganizationalUnit: firstNonEmpty(cfg.OrganizationalUnit, defaults.OrganizationalUnit),
		IncludeSuspended:   lo.ToPtr(defaults.IncludeSuspended),
		IncludeGroups:      lo.ToPtr(defaults.EnableGroupSync),
	}

	if cfg.IncludeSuspended != nil {
		settings.IncludeSuspended = lo.ToPtr(*cfg.IncludeSuspended)
	}

	if cfg.IncludeGroups != nil {
		settings.IncludeGroups = lo.ToPtr(*cfg.IncludeGroups)
	}

	return settings
}

// listDirectoryUsers pages through Google Workspace users using the resolved sync settings
func listDirectoryUsers(ctx context.Context, svc *admin.Service, cfg DirectorySyncConfig) ([]*admin.User, error) {
	customer := firstNonEmpty(cfg.Customer, "my_customer")
	users := make([]*admin.User, 0)
	pageToken := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Users.List().
			MaxResults(directoryDefaultPageSize).
			Projection("full").
			ViewType("admin_view").
			Fields(googleapi.Field(userDirectoryFields)).
			Context(ctx)

		if cfg.Domain != "" {
			call = call.Domain(cfg.Domain)
		} else {
			call = call.Customer(customer)
		}

		if cfg.Query != "" {
			call = call.Query(cfg.Query)
		}

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

// listDirectoryGroups pages through Google Workspace groups using the resolved sync settings
func listDirectoryGroups(ctx context.Context, svc *admin.Service, cfg DirectorySyncConfig) ([]*admin.Group, error) {
	customer := firstNonEmpty(cfg.Customer, "my_customer")
	groups := make([]*admin.Group, 0)
	pageToken := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Groups.List().
			MaxResults(directoryDefaultPageSize).
			Fields(googleapi.Field(groupDirectoryFields)).
			Context(ctx)

		if cfg.Domain != "" {
			call = call.Domain(cfg.Domain)
		} else {
			call = call.Customer(customer)
		}

		if cfg.Query != "" {
			call = call.Query(cfg.Query)
		}

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

// listGroupMembers pages through Google Workspace group members for one group
func listGroupMembers(ctx context.Context, svc *admin.Service, group *admin.Group) ([]*admin.Member, error) {
	if group == nil {
		return nil, nil
	}

	groupKey := firstNonEmpty(group.Id, group.Email)
	if groupKey == "" {
		return nil, nil
	}

	members := make([]*admin.Member, 0)
	pageToken := ""

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		call := svc.Members.List(groupKey).
			MaxResults(directoryDefaultPageSize).
			Fields(googleapi.Field(memberDirectoryFields)).
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

// isUserIncluded applies client-side filters that are not directly represented in the ingest mapping
func isUserIncluded(user *admin.User, cfg DirectorySyncConfig) bool {
	if user == nil {
		return false
	}

	if !lo.FromPtrOr(cfg.IncludeSuspended, false) && user.Suspended {
		return false
	}

	return matchesOrganizationalUnit(cfg.OrganizationalUnit, user.OrgUnitPath)
}

// isDirectoryUserMember reports whether a Google group member refers to a user account
func isDirectoryUserMember(member *admin.Member) bool {
	if member == nil {
		return false
	}

	if member.Type == "" {
		return true
	}

	return strings.EqualFold(member.Type, "USER")
}

// isIncludedMembershipMember reports whether the member was included in the account ingest set
func isIncludedMembershipMember(member *admin.Member, includedUsers map[string]struct{}) bool {
	if member == nil || len(includedUsers) == 0 {
		return false
	}

	if member.Id != "" {
		if _, ok := includedUsers[member.Id]; ok {
			return true
		}
	}

	if member.Email != "" {
		_, ok := includedUsers[strings.ToLower(member.Email)]
		return ok
	}

	return false
}

// matchesOrganizationalUnit reports whether a user org unit is inside the configured scope
func matchesOrganizationalUnit(filter string, candidate string) bool {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return true
	}

	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return false
	}

	normalizedFilter := strings.TrimSuffix(filter, "/")
	normalizedCandidate := strings.TrimSuffix(candidate, "/")

	if normalizedCandidate == normalizedFilter {
		return true
	}

	return strings.HasPrefix(normalizedCandidate, normalizedFilter+"/")
}

// directoryUserResource returns the best stable resource identifier for one user
func directoryUserResource(user *admin.User) string {
	if user == nil {
		return ""
	}

	return firstNonEmpty(user.PrimaryEmail, user.Id)
}

// directoryGroupResource returns the best stable resource identifier for one group
func directoryGroupResource(group *admin.Group) string {
	if group == nil {
		return ""
	}

	return firstNonEmpty(group.Email, group.Id)
}

// directoryMemberResource returns the best stable resource identifier for one group member
func directoryMemberResource(member *admin.Member) string {
	if member == nil {
		return ""
	}

	return firstNonEmpty(member.Email, member.Id)
}

// firstNonEmpty returns the first non-empty string after trimming whitespace
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}
