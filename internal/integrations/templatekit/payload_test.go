package templatekit

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func TestBuildDispatchPayload_EmptyDefaultsSingleOverlay(t *testing.T) {
	t.Parallel()

	type overlay struct {
		Name string `json:"name"`
	}

	result, err := BuildDispatchPayload(nil, overlay{Name: "test"})
	assert.NilError(t, err)

	var out map[string]any
	assert.NilError(t, json.Unmarshal(result, &out))
	assert.Assert(t, is.Equal("test", out["name"]))
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
	assert.NilError(t, err)

	var out map[string]any
	assert.NilError(t, json.Unmarshal(result, &out))
	assert.Assert(t, is.Equal("override", out["subject"]))
	assert.Assert(t, is.Equal("default body", out["body"]))
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
	assert.NilError(t, err)

	var out map[string]any
	assert.NilError(t, json.Unmarshal(result, &out))
	assert.Assert(t, is.Equal("second", out["key"]))
	assert.Assert(t, is.Equal("added", out["extra"]))
}

func TestBuildDispatchPayload_NilDefaults(t *testing.T) {
	t.Parallel()

	result, err := BuildDispatchPayload(nil)
	assert.NilError(t, err)
	assert.Assert(t, is.DeepEqual(json.RawMessage(`{}`), result))
}

func TestBuildDispatchPayload_EmptyDefaults(t *testing.T) {
	t.Parallel()

	result, err := BuildDispatchPayload(map[string]any{})
	assert.NilError(t, err)
	assert.Assert(t, result != nil)
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
	assert.NilError(t, err)
	assert.Assert(t, is.Equal("C-ORIGINAL", c.Channel))
}

func TestResolveOperationTemplate_BothIDAndKeyUsesResolutionPath(t *testing.T) {
	t.Parallel()

	type cfg struct {
		TemplateID  string `json:"templateId,omitempty"`
		TemplateKey string `json:"templateKey,omitempty"`
	}

	c := cfg{}
	err := ResolveOperationTemplate(t.Context(), types.OperationRequest{}, "id", "key", &c)
	assert.ErrorIs(t, err, ErrTemplateNotFound)
}

func TestOperationOwnerID_FromRequestIntegration(t *testing.T) {
	t.Parallel()

	ownerID := operationOwnerID(context.Background(), types.OperationRequest{
		Integration: &generated.Integration{OwnerID: "org-request"},
	})
	assert.Assert(t, is.Equal("org-request", ownerID))
}

func TestOperationOwnerID_FromOperationContext(t *testing.T) {
	t.Parallel()

	ctx := gala.WithOperationContext(context.Background(), gala.OperationContext{OwnerID: "org-meta"})
	ownerID := operationOwnerID(ctx, types.OperationRequest{})
	assert.Assert(t, is.Equal("org-meta", ownerID))
}
