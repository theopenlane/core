package operations

import (
	"context"
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestInjectLinks_NoRules(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{"external_id":"f-1","category":"S3.8"}`)

	result, err := injectLinks(context.Background(), nil, "org-1", nil, entityops.SchemaFinding, payload)
	assert.NilError(t, err)
	assert.Equal(t, string(result), string(payload))
}

func TestInjectLinks_UnknownTargetSchema(t *testing.T) {
	t.Parallel()

	_, err := injectLinks(context.Background(), nil, "org-1", []types.LinkRule{{TargetSchema: "NotASchema", TargetField: "ref_code", SourceField: "category"}}, entityops.SchemaFinding, json.RawMessage(`{}`))
	assert.ErrorIs(t, err, ErrLinkFailed)
	assert.ErrorIs(t, err, registry.ErrLinkEdgeNotFound)
}

func TestInjectLinks_UnknownEdge(t *testing.T) {
	t.Parallel()

	_, err := injectLinks(context.Background(), nil, "org-1", []types.LinkRule{{TargetSchema: "Control", Edge: "not_an_edge", TargetField: "ref_code", SourceField: "category"}}, entityops.SchemaFinding, json.RawMessage(`{}`))
	assert.ErrorIs(t, err, ErrLinkFailed)
	assert.ErrorIs(t, err, registry.ErrLinkEdgeNotFound)
}

func TestResolveLinkEdge_AmbiguousTargetRequiresEdge(t *testing.T) {
	t.Parallel()

	// Asset declares multiple edges targeting Asset, so a rule addressing the type alone must fail
	_, err := registry.ResolveLinkEdge(entityops.SchemaAsset, types.LinkRule{TargetSchema: "Asset"})
	assert.ErrorIs(t, err, registry.ErrLinkEdgeAmbiguous)
}

func TestResolveLinkEdge_ExplicitEdgeTargetMismatch(t *testing.T) {
	t.Parallel()

	edge, err := registry.ResolveLinkEdge(entityops.SchemaFinding, types.LinkRule{TargetSchema: "Control", Edge: "controls"})
	assert.NilError(t, err)
	assert.Equal(t, edge.Name, "controls")

	_, err = registry.ResolveLinkEdge(entityops.SchemaFinding, types.LinkRule{TargetSchema: "Risk", Edge: "controls"})
	assert.ErrorIs(t, err, registry.ErrLinkEdgeNotFound)
}
