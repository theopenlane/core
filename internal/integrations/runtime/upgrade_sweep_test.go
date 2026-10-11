package runtime

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

func TestStaleInstallationsKeepsOnlyOlderVersions(t *testing.T) {
	reg, err := testint.VersionedRegistry(func() (types.Definition, error) {
		return types.Definition{DefinitionSpec: types.DefinitionSpec{ID: "sweep-def"}}, nil
	})
	assert.NilError(t, err)

	current := reg.Version("sweep-def")
	assert.Assert(t, current != "")

	installations := []*ent.Integration{
		{ID: "current", DefinitionID: "sweep-def", DefinitionVersion: current},
		{ID: "empty", DefinitionID: "sweep-def", DefinitionVersion: ""},
		{ID: "older", DefinitionID: "sweep-def", DefinitionVersion: "01A00000000000000000000000"},
		{ID: "newer", DefinitionID: "sweep-def", DefinitionVersion: "7ZZZZZZZZZZZZZZZZZZZZZZZZZ"},
	}

	stale := staleInstallations(reg, installations)

	assert.DeepEqual(t, lo.Map(stale, func(inst *ent.Integration, _ int) string { return inst.ID }), []string{"empty", "older"})
}
