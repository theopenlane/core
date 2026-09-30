package registry

import (
	"encoding/json"

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
