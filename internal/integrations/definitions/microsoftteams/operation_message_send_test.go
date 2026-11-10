package microsoftteams

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

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
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("team-1", cfg.TeamID))
	assert.Assert(t, is.Equal("channel-1", cfg.Channel))
	assert.Assert(t, is.Equal("hello", cfg.Text))
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
	assert.ErrorIs(t, err, templatekit.ErrTemplateNotFound)
}
