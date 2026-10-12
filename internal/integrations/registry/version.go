package registry

import (
	"encoding/json"

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
