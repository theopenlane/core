package engine

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestBuildRenderedTemplateConfig verifies config assembly from rendered templates
func TestBuildRenderedTemplateConfig(t *testing.T) {
	t.Parallel()

	rendered := &renderedNotificationTemplate{
		Title:   "Title",
		Subject: "Subject",
		Body:    "Body text",
		Blocks:  []map[string]any{{"type": "section"}},
		Data:    map[string]any{"key": "value"},
	}

	config := buildRenderedTemplateConfig(rendered)
	require.Equal(t, "Title", config["title"])
	require.Equal(t, "Subject", config["subject"])
	require.Equal(t, "Body text", config["body"])
	require.NotNil(t, config["blocks"])
	require.NotNil(t, config["data"])
}
