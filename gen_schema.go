//go:build ignore

package main

import (
	"log"
	"os"

	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/vektah/gqlparser/v2/formatter"
)

// read in schema from internal package and save it to the schema file
func main() {
	execSchema := gqlgenerated.NewExecutableSchema(gqlgenerated.Config{})
	schema := execSchema.Schema()

	f, err := os.Create("schema.graphql")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	fmtr := formatter.NewFormatter(f)

	fmtr.FormatSchema(schema)
}
