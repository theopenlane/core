package registry

import (
	"encoding/json"
	"slices"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/helpers"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// computeVersion hashes the marshalled DefinitionSurface
func computeVersion(def types.Definition) (string, error) {
	encoded, err := json.Marshal(DefinitionSurface(def))
	if err != nil {
		return "", err
	}

	return helpers.NewHashBuilder().WriteStrings(string(encoded)).Hex(), nil
}

// Fingerprint is the hash of the sorted versions of every registered definition
func (r *Registry) Fingerprint() string {
	versions := lo.Map(r.Definitions(), func(def types.Definition, _ int) string { return r.Version(def.ID) })

	slices.Sort(versions)

	return helpers.NewHashBuilder().WriteStrings(versions...).Hex()
}
