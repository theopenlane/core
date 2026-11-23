package vulnintel

import (
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// DefinitionID is the stable identifier for the vulnerability intelligence definition
var DefinitionID = types.NewDefinitionRef("def_01K0VULNINTEL0000000000001")

// Builder returns the vulnerability intelligence definition, hidden and with no connections since it only listens for
// vulnerability ingests from other definitions and enriches them from the public OSV database
func Builder(cfg Config) registry.Builder {
	return registry.Builder(func() (types.Definition, error) {
		return types.Definition{
			DefinitionSpec: types.DefinitionSpec{
				ID:          DefinitionID.ID(),
				Family:      "intelligence",
				DisplayName: "Vulnerability Intelligence",
				Description: "Fill in vulnerability descriptions, CVSS vectors, fixed versions and references from the OSV database.",
				Category:    "security-posture",
				Tags:        []string{"vulnerabilities", "enrichment"},
				Active:      true,
				Visible:     false,
			},
			GalaListeners: []types.GalaListenerRegistration{
				enrichmentListener(cfg),
			},
		}, nil
	})
}
