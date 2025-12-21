//go:build ignore

package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/theopenlane/core/pkg/catalog"
)

func main() {
	// nothing fancy here, just reflect the Catalog struct
	r := jsonschema.Reflect(&catalog.Catalog{})
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Fatal("failed to marshal catalog schema:", err)
	}

	if err := os.WriteFile("genjsonschema/catalog.schema.json", data, 0600); err != nil { //nolint:mnd
		log.Fatal("failed to write catalog schema:", err)
	}
}
