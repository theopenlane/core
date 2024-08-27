//go:build ignore

package main

import (
	"log"
	"os"

	"github.com/vektah/gqlparser/v2/formatter"

	"github.com/theopenlane/core/internal/graphapi"
)

// read in schema from internal package and save it to the schema file
func main() {
	execSchema := graphapi.NewExecutableSchema(graphapi.Config{})
	schema := execSchema.Schema()

	f, err := os.Create("schema.graphql")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmtr := formatter.NewFormatter(f)

	fmtr.FormatSchema(schema)
}
