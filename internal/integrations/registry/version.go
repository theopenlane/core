package registry

import (
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/helpers"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// SurfaceHash hashes the marshalled DefinitionSurface
func SurfaceHash(def types.Definition) (string, error) {
	encoded, err := json.Marshal(DefinitionSurface(def))
	if err != nil {
		return "", err
	}

	return helpers.NewHashBuilder().WriteStrings(string(encoded)).Hex(), nil
}

// Fingerprint is the hash of every registered definition's id and version, in definition id order
func (r *Registry) Fingerprint() string {
	versions := lo.FlatMap(r.Definitions(), func(def types.Definition, _ int) []string { return []string{def.ID, r.Version(def.ID)} })

	return helpers.NewHashBuilder().WriteStrings(versions...).Hex()
}
