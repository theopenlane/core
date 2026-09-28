package registry

import (
	"encoding/json"

	"github.com/theopenlane/core/common/helpers"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// computeVersion hashes the marshalled DefinitionSurface, so any change to its encoding moves the version: slot names, the reflected stored schemas including their $id and doc tags, replacement lists, backfill declarations, connection refs, the user input and metadata schemas, and operation, webhook, and webhook event names
func computeVersion(def types.Definition) (string, error) {
	encoded, err := json.Marshal(DefinitionSurface(def))
	if err != nil {
		return "", err
	}

	return helpers.NewHashBuilder().WriteStrings(string(encoded)).Hex(), nil
}
