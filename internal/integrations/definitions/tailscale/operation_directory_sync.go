package tailscale

import (
	"context"
	"fmt"

	tsclient "github.com/tailscale/tailscale-client-go/v2"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
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

// IngestHandle adapts Tailscale directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(tailscaleClient, directorySyncOperation, ErrOperationConfigInvalid, func(ctx context.Context, _ types.OperationRequest, client *tsclient.Client, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
		if cfg.Disable {
			logx.FromContext(ctx).Debug().Msg("tailscale: directory sync is disabled")
			return nil, nil
		}

		return d.Run(ctx, client, cfg)
	})
}

// Run collects Tailscale users and optionally role-based groups and memberships
func (DirectorySync) Run(ctx context.Context, client *tsclient.Client, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
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

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaDirectoryAccount.Name,
			Envelopes: accountEnvelopes,
		},
	}

	if cfg.DisableGroupSync {
		logx.FromContext(ctx).Debug().Int("user_count", len(accountEnvelopes)).Msg("tailscale: collected users; group sync disabled")
		return payloadSets, nil
	}

	// Build role-based groups from the unique set of roles across all users
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
				return nil, err
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
			return nil, err
		}

		membershipEnvelopes = append(membershipEnvelopes, envelope)
	}

	// Also sync user-defined ACL groups from the policy file
	userByEmail := make(map[string]string, len(users))
	for _, u := range users {
		userByEmail[u.LoginName] = u.ID
	}

	acl, err := client.PolicyFile().Get(ctx)
	if err != nil {
		logx.FromContext(ctx).Warn().Err(err).Msg("tailscale: failed to fetch policy file; skipping user-defined groups")
	} else {
		for groupName, members := range acl.Groups {
			group := tailscaleGroupPayload{
				ID:   groupName,
				Name: groupName,
			}

			groupEnvelope, err := providerkit.MarshalEnvelope(groupName, group, ErrPayloadEncode)
			if err != nil {
				logx.FromContext(ctx).Error().Err(err).Str("group", groupName).Msg("tailscale: failed to marshal ACL group")
				return nil, err
			}

			groupEnvelopes = append(groupEnvelopes, groupEnvelope)

			for _, member := range members {
				userID, ok := userByEmail[member]
				if !ok {
					// skip group references, tags, autogroups, wildcards
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
					return nil, err
				}

				membershipEnvelopes = append(membershipEnvelopes, membershipEnvelope)
			}
		}
	}

	logx.FromContext(ctx).Debug().
		Int("user_count", len(accountEnvelopes)).
		Int("group_count", len(groupEnvelopes)).
		Int("membership_count", len(membershipEnvelopes)).
		Msg("tailscale: collected users, role groups, and memberships")

	payloadSets = append(payloadSets,
		types.IngestPayloadSet{
			Schema:    entityops.SchemaDirectoryGroup.Name,
			Envelopes: groupEnvelopes,
		},
		types.IngestPayloadSet{
			Schema:    entityops.SchemaDirectoryMembership.Name,
			Envelopes: membershipEnvelopes,
		},
	)

	return payloadSets, nil
}

// listTailscaleUsers fetches all users from the Tailscale API and maps them to payloads
func listTailscaleUsers(ctx context.Context, client *tsclient.Client) ([]tsclient.User, error) {
	users, err := client.Users().List(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUsersFetchFailed, err)
	}

	return users, nil
}
