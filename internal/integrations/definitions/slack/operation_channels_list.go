package slack

import (
	"context"
	"encoding/json"

	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// ChannelsListOperationInput holds per-invocation parameters for channel discovery
type ChannelsListOperationInput struct {
	// Cursor is the Slack pagination cursor from a previous channels.list response
	Cursor string `json:"cursor,omitempty" jsonschema:"title=Cursor"`
	// Limit is the maximum number of channels to return
	Limit int `json:"limit,omitempty" jsonschema:"title=Limit,description=Maximum number of channels to return (defaults to 100)."`
	// IncludeArchived controls whether archived channels are included
	IncludeArchived bool `json:"include_archived,omitempty" jsonschema:"title=Include Archived Channels"`
	// Types filters the conversation types returned by Slack
	Types []string `json:"types,omitempty" jsonschema:"title=Conversation Types,description=Optional Slack conversation types such as public_channel or private_channel."`
}

// ChannelSummary captures one Slack conversation suitable for notification delivery
type ChannelSummary struct {
	// ID is the Slack conversation identifier
	ID string `json:"id"`
	// Name is the channel display name
	Name string `json:"name"`
	// IsPrivate indicates the channel is private
	IsPrivate bool `json:"isPrivate"`
	// IsArchived indicates the channel is archived
	IsArchived bool `json:"isArchived"`
	// IsMember indicates the app token is currently a member of the channel
	IsMember bool `json:"isMember"`
}

// ChannelsList returns available Slack channels and pagination metadata
type ChannelsList struct {
	// Channels holds the returned Slack conversations
	Channels []ChannelSummary `json:"channels"`
	// NextCursor holds the Slack pagination cursor for the next page
	NextCursor string `json:"nextCursor,omitempty"`
}

// Handle adapts channels list to the generic operation registration boundary
func (l ChannelsList) Handle() types.OperationHandler {
	return providerkit.OperationWithClientConfig(SlackClient, ChannelsListOperation, ErrOperationConfigInvalid, l.Run)
}

// Run lists Slack conversations that can be used as message destinations
func (ChannelsList) Run(ctx context.Context, c *slackgo.Client, cfg ChannelsListOperationInput) (json.RawMessage, error) {
	limit := cfg.Limit
	if limit <= 0 {
		limit = 100
	}

	types := cfg.Types
	if len(types) == 0 {
		types = []string{"public_channel", "private_channel"}
	}

	channels, nextCursor, err := c.GetConversationsContext(ctx, &slackgo.GetConversationsParameters{
		Cursor:          cfg.Cursor,
		ExcludeArchived: !cfg.IncludeArchived,
		Limit:           limit,
		Types:           types,
	})
	if err != nil {
		return nil, ErrConversationsListFailed
	}

	items := make([]ChannelSummary, 0, len(channels))
	for _, channel := range channels {
		items = append(items, ChannelSummary{
			ID:         channel.ID,
			Name:       channel.Name,
			IsPrivate:  channel.IsPrivate,
			IsArchived: channel.IsArchived,
			IsMember:   channel.IsMember,
		})
	}

	return providerkit.EncodeResult(ChannelsList{
		Channels:   items,
		NextCursor: nextCursor,
	}, ErrResultEncode)
}
