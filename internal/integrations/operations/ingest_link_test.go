package operations

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

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

func TestLinkSpecs_KeyMatchRule(t *testing.T) {
	t.Parallel()

	rules := []types.LinkRule{
		{TargetSchema: "Control", Edge: "controls", TargetField: "ref_code", SourceField: "control_ref", SourceList: "control_refs"},
	}

	specs, err := linkSpecs(entityops.SchemaFinding, rules)
	assert.NilError(t, err)
	assert.Equal(t, len(specs), 1)
	assert.Equal(t, specs[0].Edge, "controls")
	assert.Assert(t, specs[0].Target.KeyMatch != nil)
	assert.Equal(t, specs[0].Target.KeyMatch.TargetField, "ref_code")
	assert.Equal(t, specs[0].Target.KeyMatch.SourceField, "control_ref")
	assert.Equal(t, specs[0].Target.KeyMatch.SourceList, "control_refs")
	assert.Equal(t, specs[0].Target.Expression, "")
}

func TestLinkSpecs_ExpressionRule(t *testing.T) {
	t.Parallel()

	rules := []types.LinkRule{
		{TargetSchema: "Control", Edge: "controls", Expression: `target.ref_code == source.control_ref`},
	}

	specs, err := linkSpecs(entityops.SchemaFinding, rules)
	assert.NilError(t, err)
	assert.Equal(t, len(specs), 1)
	assert.Equal(t, specs[0].Edge, "controls")
	assert.Assert(t, specs[0].Target.KeyMatch == nil)
	assert.Equal(t, specs[0].Target.Expression, `target.ref_code == source.control_ref`)
}

func TestLinkSpecs_MultipleRulesPreserveOrder(t *testing.T) {
	t.Parallel()

	rules := []types.LinkRule{
		{TargetSchema: "Control", Edge: "controls", TargetField: "ref_code", SourceField: "control_ref"},
		{TargetSchema: "Risk", Edge: "risks", TargetField: "ref_code", SourceField: "risk_ref"},
	}

	specs, err := linkSpecs(entityops.SchemaFinding, rules)
	assert.NilError(t, err)
	assert.Equal(t, len(specs), 2)
	assert.Equal(t, specs[0].Edge, "controls")
	assert.Equal(t, specs[1].Edge, "risks")
}

func TestLinkSpecs_UnresolvableEdgeWrapsErrLinkFailed(t *testing.T) {
	t.Parallel()

	// Asset declares multiple edges targeting Asset, so a rule addressing the type alone is ambiguous
	_, err := linkSpecs(entityops.SchemaAsset, []types.LinkRule{{TargetSchema: "Asset"}})
	assert.ErrorIs(t, err, ErrLinkFailed)
	assert.ErrorIs(t, err, registry.ErrLinkEdgeAmbiguous)
}

func TestLinkSpecs_EmptyRulesReturnsEmptySpecs(t *testing.T) {
	t.Parallel()

	specs, err := linkSpecs(entityops.SchemaFinding, nil)
	assert.NilError(t, err)
	assert.Equal(t, len(specs), 0)
}
