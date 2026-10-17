package operations

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// ListTeamsFunc lists the teams, or workspace, the connection can post to with the channels in each
type ListTeamsFunc[C any] func(ctx context.Context, client C) ([]Team, error)

// ChannelsListOperation holds per-invocation parameters for the channels list operation
type ChannelsListOperation struct{}

// ChannelsList lists the teams the connection can post to and the channels in each
type ChannelsList struct {
	// Teams are the teams with their channels
	Teams []Team `json:"teams"`
}

// Team is one team or workspace
type Team struct {
	// ID is the team identifier
	ID string `json:"id"`
	// DisplayName is the team name
	DisplayName string `json:"displayName"`
	// Channels are the channels in the team
	Channels []Channel `json:"channels"`
}

// Channel is one channel within a team
type Channel struct {
	// ID is the channel identifier
	ID string `json:"id"`
	// DisplayName is the channel name
	DisplayName string `json:"displayName"`
}

// NewChannelsListOperation builds the inline operation that lists message destinations through list
func NewChannelsListOperation[C any](list ListTeamsFunc[C], encodeErr error) types.OperationRef[ChannelsListOperation] {
	return types.OperationPayloadOf[ChannelsListOperation]().
		Handles(channelsList[C]{list: list, encodeErr: encodeErr}.run).
		Policy(types.ExecutionPolicy{Inline: true})
}

// channelsList binds a provider's team listing to the shared result encoding
type channelsList[C any] struct {
	// list lists the provider's teams and channels
	list ListTeamsFunc[C]
	// encodeErr is the provider error returned when the result cannot be serialized
	encodeErr error
}

// run lists the teams and channels for choosing a message destination
func (l channelsList[C]) run(ctx context.Context, _ types.OperationRequest, client C, _ ChannelsListOperation) (json.RawMessage, error) {
	teams, err := l.list(ctx, client)
	if err != nil {
		return nil, err
	}

	return providerkit.EncodeResult(ChannelsList{Teams: teams}, l.encodeErr)
}
