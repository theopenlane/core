package keycloak

import (
	"context"

	gocloak "github.com/Nerzal/gocloak/v13"
	"github.com/samber/lo"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// directorySyncOperation is the Keycloak directory sync operation
var directorySyncOperation = types.OperationRefOf[DirectorySync]().Ingests(keycloakClient, runDirectorySync)

// runDirectorySync collects Keycloak directory users, groups, and memberships
func runDirectorySync(ctx context.Context, request types.OperationRequest, gc *gocloak.GoCloak, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
	cred, err := resolveCredential(request.Credentials)
	if err != nil {
		return nil, err
	}

	token, err := gc.LoginClient(ctx, cred.ClientID, cred.ClientSecret, cred.Realm)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("error acquiring keycloak token")

		return nil, ErrTokenAcquireFailed
	}

	return collectDirectory(ctx, gc, token.AccessToken, cred.Realm, cfg)
}

// collectDirectory collects Keycloak directory users, groups, and memberships
func collectDirectory(ctx context.Context, gc *gocloak.GoCloak, token, realm string, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
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

		resourceID := lo.FromPtr(user.ID)

		envelope, err := providerkit.MarshalEnvelope(resourceID, user, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)
		includedUsers[resourceID] = struct{}{}
	}

	payloadSets := providerkit.DirectoryAccountPayloadSets(accountEnvelopes)

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

		groupID := lo.FromPtr(group.ID)

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

			memberID := lo.FromPtr(member.ID)

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

	return append(payloadSets, providerkit.DirectoryGroupPayloadSets(groupEnvelopes, membershipEnvelopes, true)...), nil
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
