package keystore

import (
	"github.com/invopop/jsonschema"

	"github.com/theopenlane/common/integrations/config"
)

// Schema returns the JSON schema for integration provider specifications.
func Schema() *jsonschema.Schema {
	return jsonschema.Reflect(&config.ProviderSpec{})
}
