//go:build ignore

package main

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/vektah/gqlparser/v2/formatter"

	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	"github.com/theopenlane/core/pkg/genhelpers"
)

// read in schema from internal package and save it to the schema file
func main() {
	genhelpers.SetupLogging()

	genhelpers.ChangeToRootDir("../../../")

	_, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get current directory")
	}

	log.Info().Msg("Generating schema for client")

	execSchema := gqlgenerated.NewExecutableSchema(gqlgenerated.Config{})
	schema := execSchema.Schema()

	f, err := os.Create("internal/graphapi/clientschema/schema.graphql")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create schema file")
	}

	defer f.Close()

	fmtr := formatter.NewFormatter(f)

	log.Info().Msg("writing schema.graphl to file")

	fmtr.FormatSchema(schema)
}
