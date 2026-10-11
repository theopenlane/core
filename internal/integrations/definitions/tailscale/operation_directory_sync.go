package tailscale

import (
	"context"
	"fmt"

	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// tailscaleGroupPayload is the envelope payload for one Tailscale role group record
type tailscaleGroupPayload struct {
	// ID is the role identifier (e.g. "admin", "member")
	ID string `json:"id"`
	// Name is the human-readable role label
	Name string `json:"name"`
}

// tailscaleMembershipPayload is the envelope payload for one Tailscale role membership record
type tailscaleMembershipPayload struct {
	// GroupID is the role identifier the user belongs to
	GroupID string `json:"group_id"`
	// UserID is the Tailscale user identifier
	UserID string `json:"user_id"`
}

// runDirectorySync collects Tailscale users and optionally role-based groups and memberships
func runDirectorySync(ctx context.Context, _ types.OperationRequest, client *tsclient.Client, cfg providerkit.DirectorySync) ([]types.IngestPayloadSet, error) {
	users, err := listTailscaleUsers(ctx, client)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))

	for _, user := range users {
		envelope, err := providerkit.MarshalEnvelope(user.ID, user, ErrPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("user", user.LoginName).Msg("tailscale: failed to marshal user")
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)
	}

	payloadSets := providerkit.DirectoryAccountPayloadSets(accountEnvelopes)

	if cfg.DisableGroupSync {
		logx.FromContext(ctx).Debug().Int("user_count", len(accountEnvelopes)).Msg("tailscale: collected users; group sync disabled")
		return payloadSets, nil
	}

	groupEnvelopes, membershipEnvelopes, err := roleGroupEnvelopes(ctx, users)
	if err != nil {
		return nil, err
	}

	aclGroups, aclMemberships, membershipsComplete, err := aclGroupEnvelopes(ctx, client, users)
	if err != nil {
		return nil, err
	}

	groupEnvelopes = append(groupEnvelopes, aclGroups...)
	membershipEnvelopes = append(membershipEnvelopes, aclMemberships...)

	logx.FromContext(ctx).Debug().Int("user_count", len(accountEnvelopes)).Int("group_count", len(groupEnvelopes)).Int("membership_count", len(membershipEnvelopes)).Msg("tailscale: collected users, role groups, and memberships")

	return append(payloadSets, providerkit.DirectoryGroupPayloadSets(groupEnvelopes, membershipEnvelopes, membershipsComplete)...), nil
}

// roleGroupEnvelopes derives one group envelope per Tailscale role and one membership envelope per user holding it
func roleGroupEnvelopes(ctx context.Context, users []tsclient.User) ([]types.MappingEnvelope, []types.MappingEnvelope, error) {
	rolesSeen := make(map[tsclient.UserRole]struct{})
	groupEnvelopes := make([]types.MappingEnvelope, 0)
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, user := range users {
		role := user.Role
		if role == "" {
			continue
		}

		if _, seen := rolesSeen[role]; !seen {
			rolesSeen[role] = struct{}{}

			group := tailscaleGroupPayload{
				ID:   string(role),
				Name: string(role),
			}

			envelope, err := providerkit.MarshalEnvelope(string(role), group, ErrPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("role", string(role)).Msg("tailscale: failed to marshal role group")
				return nil, nil, err
			}

			groupEnvelopes = append(groupEnvelopes, envelope)
		}

		membership := tailscaleMembershipPayload{
			GroupID: string(role),
			UserID:  user.ID,
		}

		membershipKey := fmt.Sprintf("%s:%s", role, user.ID)

		envelope, err := providerkit.MarshalEnvelope(membershipKey, membership, ErrPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("user", user.LoginName).Str("role", string(role)).Msg("tailscale: failed to marshal membership")
			return nil, nil, err
		}

		membershipEnvelopes = append(membershipEnvelopes, envelope)
	}

	return groupEnvelopes, membershipEnvelopes, nil
}

// aclGroupEnvelopes derives group and membership envelopes from the tailnet policy file; complete is false when the policy file could not be fetched
func aclGroupEnvelopes(ctx context.Context, client *tsclient.Client, users []tsclient.User) ([]types.MappingEnvelope, []types.MappingEnvelope, bool, error) {
	acl, err := client.PolicyFile().Get(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("tailscale: failed to fetch policy file; skipping user-defined groups")

		return nil, nil, false, nil
	}

	userByEmail := make(map[string]string, len(users))
	for _, u := range users {
		userByEmail[u.LoginName] = u.ID
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(acl.Groups))
	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for groupName, members := range acl.Groups {
		group := tailscaleGroupPayload{
			ID:   groupName,
			Name: groupName,
		}

		groupEnvelope, err := providerkit.MarshalEnvelope(groupName, group, ErrPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("group", groupName).Msg("tailscale: failed to marshal ACL group")
			return nil, nil, false, err
		}

		groupEnvelopes = append(groupEnvelopes, groupEnvelope)

		for _, member := range members {
			userID, ok := userByEmail[member]
			if !ok {
				continue
			}

			membership := tailscaleMembershipPayload{
				GroupID: groupName,
				UserID:  userID,
			}

			membershipKey := fmt.Sprintf("%s:%s", groupName, userID)

			membershipEnvelope, err := providerkit.MarshalEnvelope(membershipKey, membership, ErrPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("group", groupName).Str("user", member).Msg("tailscale: failed to marshal ACL group membership")
				return nil, nil, false, err
			}

			membershipEnvelopes = append(membershipEnvelopes, membershipEnvelope)
		}
	}

	return groupEnvelopes, membershipEnvelopes, true, nil
}

// listTailscaleUsers fetches all users from the Tailscale API and maps them to payloads
func listTailscaleUsers(ctx context.Context, client *tsclient.Client) ([]tsclient.User, error) {
	users, err := client.Users().List(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUsersFetchFailed, err)
	}

	return users, nil
}
