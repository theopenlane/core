package templatekit

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestBuildDispatchPayload_EmptyDefaultsSingleOverlay(t *testing.T) {
	t.Parallel()

	type overlay struct {
		Name string `json:"name"`
	}

	result, err := BuildDispatchPayload(nil, overlay{Name: "test"})
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(result, &out))
	require.Equal(t, "test", out["name"])
}

func TestBuildDispatchPayload_DefaultsWithOverlayMerge(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{
		"subject": "default subject",
		"body":    "default body",
	}

	type overlay struct {
		Subject string `json:"subject"`
	}

	result, err := BuildDispatchPayload(defaults, overlay{Subject: "override"})
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(result, &out))
	require.Equal(t, "override", out["subject"])
	require.Equal(t, "default body", out["body"])
}

func TestBuildDispatchPayload_MultipleOverlaysInOrder(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{"key": "base"}

	type first struct {
		Key   string `json:"key"`
		Extra string `json:"extra"`
	}
	type second struct {
		Key string `json:"key"`
	}

	result, err := BuildDispatchPayload(defaults, first{Key: "first", Extra: "added"}, second{Key: "second"})
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(result, &out))
	require.Equal(t, "second", out["key"])
	require.Equal(t, "added", out["extra"])
}

func TestBuildDispatchPayload_NilDefaults(t *testing.T) {
	t.Parallel()

	result, err := BuildDispatchPayload(nil)
	require.NoError(t, err)
	require.Equal(t, json.RawMessage(`{}`), result)
}

func TestBuildDispatchPayload_EmptyDefaults(t *testing.T) {
	t.Parallel()

	result, err := BuildDispatchPayload(map[string]any{})
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestResolveOperationTemplate_NoopWhenEmpty(t *testing.T) {
	t.Parallel()

	type cfg struct {
		TemplateID  string `json:"templateId,omitempty"`
		TemplateKey string `json:"templateKey,omitempty"`
		Channel     string `json:"channel,omitempty"`
	}

	c := cfg{Channel: "C-ORIGINAL"}
	err := ResolveOperationTemplate(t.Context(), types.OperationRequest{}, "", "", &c)
	require.NoError(t, err)
	require.Equal(t, "C-ORIGINAL", c.Channel)
}

func TestResolveOperationTemplate_BothIDAndKeyUsesResolutionPath(t *testing.T) {
	t.Parallel()

	type cfg struct {
		TemplateID  string `json:"templateId,omitempty"`
		TemplateKey string `json:"templateKey,omitempty"`
	}

	c := cfg{}
	err := ResolveOperationTemplate(t.Context(), types.OperationRequest{}, "id", "key", &c)
	require.ErrorIs(t, err, ErrTemplateNotFound)
}

func TestOperationOwnerID_FromRequestIntegration(t *testing.T) {
	t.Parallel()

	ownerID := operationOwnerID(context.Background(), types.OperationRequest{
		Integration: &generated.Integration{OwnerID: "org-request"},
	})
	require.Equal(t, "org-request", ownerID)
}

func TestOperationOwnerID_FromExecutionMetadata(t *testing.T) {
	t.Parallel()

	ctx := types.WithExecutionMetadata(context.Background(), types.ExecutionMetadata{OwnerID: "org-meta"})
	ownerID := operationOwnerID(ctx, types.OperationRequest{})
	require.Equal(t, "org-meta", ownerID)
}
