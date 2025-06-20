package main

import (
	"encoding/json"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/theopenlane/core/pkg/catalog"
)

func main() {
	r := jsonschema.Reflect(&catalog.Catalog{})
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("./pkg/catalog/catalog.schema.json", data, 0600); err != nil {
		panic(err)
	}
}
