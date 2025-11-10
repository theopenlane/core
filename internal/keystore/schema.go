package keystore

import "github.com/invopop/jsonschema"

func Schema() *jsonschema.Schema {
	return jsonschema.Reflect(&ProviderSpec{})
}
