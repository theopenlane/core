package engine

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/workflows"
)

func TestSplitNotificationTargetsSeparatesUserAndChannelTargets(t *testing.T) {
	targets := []workflows.TargetConfig{
		{
			Type: enums.WorkflowTargetTypeResolver,
		},
		{
			Type:        enums.WorkflowTargetTypeChannel,
			Channel:     enums.ChannelSlack,
			Destination: "C12345",
			Config: map[string]any{
				"thread_ts": "123.456",
			},
		},
	}

	userTargets, channelTargets, err := splitNotificationTargets(targets)
	require.NoError(t, err)
	require.Len(t, userTargets, 1)
	require.Len(t, channelTargets, 1)
	require.Equal(t, enums.ChannelSlack, channelTargets[0].Channel)
	require.Equal(t, "C12345", channelTargets[0].Destination)
	require.Equal(t, "123.456", channelTargets[0].Config["thread_ts"])
}

func TestSplitNotificationTargetsRequiresChannel(t *testing.T) {
	targets := []workflows.TargetConfig{
		{
			Type:        enums.WorkflowTargetTypeChannel,
			Destination: "C12345",
		},
	}

	_, _, err := splitNotificationTargets(targets)
	require.ErrorIs(t, err, ErrNotificationChannelTargetChannelRequired)
}

func TestSplitNotificationTargetsRequiresDestination(t *testing.T) {
	targets := []workflows.TargetConfig{
		{
			Type:    enums.WorkflowTargetTypeChannel,
			Channel: enums.ChannelSlack,
		},
	}

	_, _, err := splitNotificationTargets(targets)
	require.ErrorIs(t, err, ErrNotificationChannelTargetDestinationRequired)
}

func TestSplitNotificationTargetsRejectsInAppChannelTarget(t *testing.T) {
	targets := []workflows.TargetConfig{
		{
			Type:        enums.WorkflowTargetTypeChannel,
			Channel:     enums.ChannelInApp,
			Destination: "any",
		},
	}

	_, _, err := splitNotificationTargets(targets)
	require.ErrorIs(t, err, ErrNotificationChannelUnsupported)
}
