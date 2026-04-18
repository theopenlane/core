package slack

import (
	"context"
	"time"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

// slackUserPayload is the normalized payload for a Slack workspace user
type slackUserPayload struct {
	// ID is the Slack user identifier
	ID string `json:"id"`
	// TeamID is the Slack team/workspace identifier
	TeamID string `json:"team_id"`
	// Name is the Slack username (handle)
	Name string `json:"name,omitempty"`
	// RealName is the user's full display name
	RealName string `json:"real_name,omitempty"`
	// Email is the user's email address from their profile
	Email string `json:"email,omitempty"`
	// FirstName is the user's first name from their profile
	FirstName string `json:"first_name,omitempty"`
	// LastName is the user's last name from their profile
	LastName string `json:"last_name,omitempty"`
	// DisplayName is the user's chosen display name
	DisplayName string `json:"display_name,omitempty"`
	// Title is the user's job title
	Title string `json:"title,omitempty"`
	// AvatarURL is the user's avatar image URL
	AvatarURL string `json:"avatar_url,omitempty"`
	// Deleted reports whether the user account has been deactivated
	Deleted bool `json:"deleted"`
	// IsBot reports whether this user is a bot account
	IsBot bool `json:"is_bot"`
	// IsAdmin reports whether the user is a workspace admin
	IsAdmin bool `json:"is_admin"`
	// Has2FA reports whether 2FA is enabled for the user
	Has2FA bool `json:"has_2fa"`
	// IsRestricted reports whether this user is a restricted user (e.g. multi channel guest)
	IsRestricted bool `json:"is_restricted"`
	// IsUltraRestricted reports whether this user is a restricted user (e.g. single channel guest)
	IsUltraRestricted bool `json:"is_ultra_restricted"`
	// IsStranger reports whether this user is a external user (e.g. slack-connect)
	IsStranger bool `json:"is_stranger"`
	// IsExternal reports whether this user is a external user (e.g. slack-connect)
	IsExternal bool `json:"is_external"`
}

// IngestHandle adapts directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequest(slackClient, func(ctx context.Context, request types.OperationRequest, client *slackgo.Client) ([]types.IngestPayloadSet, error) {
		return d.Run(ctx, client, request.LastRunAt)
	})
}

// Run collects Slack workspace users and emits directory account ingest payloads
func (DirectorySync) Run(ctx context.Context, client *slackgo.Client, lastRunAt *time.Time) ([]types.IngestPayloadSet, error) {
	users, err := client.GetUsersContext(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("slack directory sync: failed to fetch users")
		return nil, ErrUsersFetchFailed
	}

	envelopes := make([]types.MappingEnvelope, 0, len(users))

	for _, user := range users {
		if user.IsBot {
			continue
		}

		if lastRunAt != nil && !time.Unix(int64(user.Updated), 0).After(*lastRunAt) {
			continue
		}

		payload := normalizeUser(user)

		resource := user.TeamID + "/" + user.ID

		envelope, err := providerkit.MarshalEnvelope(resource, payload, ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: envelopes,
		},
	}, nil
}

func normalizeUser(user slackgo.User) slackUserPayload {
	return slackUserPayload{
		ID:                user.ID,
		TeamID:            user.TeamID,
		Name:              user.Name,
		RealName:          user.RealName,
		Email:             user.Profile.Email,
		FirstName:         user.Profile.FirstName,
		LastName:          user.Profile.LastName,
		DisplayName:       user.Profile.DisplayName,
		Title:             user.Profile.Title,
		AvatarURL:         user.Profile.Image192,
		Deleted:           user.Deleted,
		IsBot:             user.IsBot,
		IsAdmin:           user.IsAdmin,
		Has2FA:            user.Has2FA,
		IsRestricted:      user.IsRestricted,
		IsUltraRestricted: user.IsUltraRestricted,
		IsStranger:        user.IsStranger,
		IsExternal:        user.IsRestricted || user.IsStranger,
	}
}
