package microsoftteams

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/v2/internal/integrations/templatekit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestResolveOperationTemplateNoop(t *testing.T) {
	t.Parallel()

	cfg := MessageSendOperation{
		TeamID:  "team-1",
		Channel: "channel-1",
		Text:    "hello",
	}

	err := cfg.ResolveTemplate(context.Background(), types.OperationRequest{}, &cfg)
	require.NoError(t, err)
	require.Equal(t, "team-1", cfg.TeamID)
	require.Equal(t, "channel-1", cfg.Channel)
	require.Equal(t, "hello", cfg.Text)
}

func TestResolveOperationTemplateBothRefsUseResolutionPath(t *testing.T) {
	t.Parallel()

	cfg := MessageSendOperation{
		TemplateID:  "some-id",
		TemplateKey: "some-key",
		TeamID:      "team-1",
		Channel:     "channel-1",
		Text:        "hello",
	}

	err := cfg.ResolveTemplate(context.Background(), types.OperationRequest{}, &cfg)
	require.ErrorIs(t, err, templatekit.ErrTemplateNotFound)
}
