package keycloak

import (
	"context"

	gocloak "github.com/Nerzal/gocloak/v13"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// DirectorySync collects Keycloak directory users, groups, and memberships for ingest
type DirectorySync struct{}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(keycloakClient, func(ctx context.Context, request types.OperationRequest, gc *gocloak.GoCloak) ([]types.IngestPayloadSet, error) {
		var cfg UserInput
		if request.Integration != nil {
			_ = jsonx.UnmarshalIfPresent(request.Integration.Config.ClientConfig, &cfg)
		}

		cred, err := resolveCredential(request.Credentials)
		if err != nil {
			return nil, err
		}

		token, err := gc.LoginClient(ctx, cred.ClientID, cred.ClientSecret, cred.Realm)
		if err != nil {
			return nil, ErrTokenAcquireFailed
		}

		return d.Run(ctx, gc, token.AccessToken, cred.Realm, cfg)
	})
}

// Run collects Keycloak directory users, groups, and memberships
func (DirectorySync) Run(ctx context.Context, gc *gocloak.GoCloak, token, realm string, cfg UserInput) ([]types.IngestPayloadSet, error) {
	users, err := listDirectoryUsers(ctx, gc, token, realm)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))
	includedUsers := make(map[string]struct{}, len(users))

	for _, user := range users {
		if user.ID == nil {
			continue
		}

		resourceID := derefString(user.ID)

		envelope, err := providerkit.MarshalEnvelope(resourceID, user, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)
		includedUsers[resourceID] = struct{}{}
	}

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
	}

	if cfg.DisableGroupSync {
		return payloadSets, nil
	}

	groups, err := listDirectoryGroups(ctx, gc, token, realm)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, group := range groups {
		if group.ID == nil {
			continue
		}

		groupID := derefString(group.ID)

		envelope, err := providerkit.MarshalEnvelope(groupID, group, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)

		members, err := listGroupMembers(ctx, gc, token, realm, groupID)
		if err != nil {
			return nil, err
		}

		for _, member := range members {
			if member.ID == nil {
				continue
			}

			memberID := derefString(member.ID)

			if _, ok := includedUsers[memberID]; !ok {
				continue
			}

			envelope, err := providerkit.MarshalEnvelope(groupID, member, ErrPayloadEncode)
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

// listDirectoryUsers pages through all Keycloak users in the realm
func listDirectoryUsers(ctx context.Context, gc *gocloak.GoCloak, token, realm string) ([]*gocloak.User, error) {
	users := make([]*gocloak.User, 0)
	first := 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		batch, err := gc.GetUsers(ctx, token, realm, gocloak.GetUsersParams{
			First: gocloak.IntP(first),
			Max:   gocloak.IntP(keycloakDefaultPageSize),
		})
		if err != nil {
			return nil, ErrDirectoryUsersFetchFailed
		}

		users = append(users, batch...)

		if len(batch) < keycloakDefaultPageSize {
			break
		}

		first += keycloakDefaultPageSize
	}

	return users, nil
}

// listDirectoryGroups pages through all Keycloak groups in the realm
func listDirectoryGroups(ctx context.Context, gc *gocloak.GoCloak, token, realm string) ([]*gocloak.Group, error) {
	groups := make([]*gocloak.Group, 0)
	first := 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		batch, err := gc.GetGroups(ctx, token, realm, gocloak.GetGroupsParams{
			First: gocloak.IntP(first),
			Max:   gocloak.IntP(keycloakDefaultPageSize),
			Full:  gocloak.BoolP(true),
		})
		if err != nil {
			return nil, ErrDirectoryGroupsFetchFailed
		}

		groups = append(groups, batch...)

		if len(batch) < keycloakDefaultPageSize {
			break
		}

		first += keycloakDefaultPageSize
	}

	return groups, nil
}

// listGroupMembers pages through all members of one Keycloak group
func listGroupMembers(ctx context.Context, gc *gocloak.GoCloak, token, realm, groupID string) ([]*gocloak.User, error) {
	members := make([]*gocloak.User, 0)
	first := 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		batch, err := gc.GetGroupMembers(ctx, token, realm, groupID, gocloak.GetGroupsParams{
			First: gocloak.IntP(first),
			Max:   gocloak.IntP(keycloakDefaultPageSize),
		})
		if err != nil {
			return nil, ErrDirectoryGroupMembersFetchFailed
		}

		members = append(members, batch...)

		if len(batch) < keycloakDefaultPageSize {
			break
		}

		first += keycloakDefaultPageSize
	}

	return members, nil
}