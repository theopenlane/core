package runtime

import (
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestStaleInstallationsTreatsEmptyAndMismatchedVersionsAlike(t *testing.T) {
	reg := registry.New()
	assert.NilError(t, reg.Register(types.Definition{DefinitionSpec: types.DefinitionSpec{ID: "sweep-def"}}))

	current := reg.Version("sweep-def")
	assert.Assert(t, current != "")

	installations := []*ent.Integration{
		{ID: "current", DefinitionID: "sweep-def", DefinitionVersion: current},
		{ID: "empty", DefinitionID: "sweep-def", DefinitionVersion: ""},
		{ID: "stale", DefinitionID: "sweep-def", DefinitionVersion: "stale-hash"},
	}

	stale := staleInstallations(reg, installations)

	assert.DeepEqual(t, lo.Map(stale, func(inst *ent.Integration, _ int) string { return inst.ID }), []string{"empty", "stale"})
}
