package microsoftteams

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/templatekit"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestResolveOperationTemplateNoop(t *testing.T) {
	t.Parallel()

	cfg := MessageSendOperation{
		TeamID:    "team-1",
		ChannelID: "channel-1",
		Body:      "hello",
	}

	err := templatekit.ResolveOperationTemplate(context.Background(), types.OperationRequest{}, cfg.TemplateID, cfg.TemplateKey, &cfg)
	require.NoError(t, err)
	require.Equal(t, "team-1", cfg.TeamID)
	require.Equal(t, "channel-1", cfg.ChannelID)
	require.Equal(t, "hello", cfg.Body)
}

func TestResolveOperationTemplateBothRefsUseResolutionPath(t *testing.T) {
	t.Parallel()

	cfg := MessageSendOperation{
		TemplateID:  "some-id",
		TemplateKey: "some-key",
		TeamID:      "team-1",
		ChannelID:   "channel-1",
		Body:        "hello",
	}

	err := templatekit.ResolveOperationTemplate(context.Background(), types.OperationRequest{}, cfg.TemplateID, cfg.TemplateKey, &cfg)
	require.ErrorIs(t, err, templatekit.ErrTemplateNotFound)
}
