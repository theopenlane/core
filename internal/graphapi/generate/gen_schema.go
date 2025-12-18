//go:build ignore

package main

import (
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"

	"github.com/vektah/gqlparser/v2/formatter"

	"github.com/theopenlane/core/internal/genhelpers"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	gqlhistorygenerated "github.com/theopenlane/core/internal/graphapi/historygenerated"
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

	generateClientSchema(execSchema, "internal/graphapi/clientschema/schema.graphql")

	log.Info().Msg("Generating history schema for client")
	execHistorySchema := gqlhistorygenerated.NewExecutableSchema(gqlhistorygenerated.Config{})
	generateClientSchema(execHistorySchema, "internal/graphapi/historyschema/schema.graphql")
}

func generateClientSchema(execSchema graphql.ExecutableSchema, path string) {
	log.Info().Msg("Generating schema for client")

	schema := execSchema.Schema()

	f, err := os.Create(path)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create schema file")
	}

	defer f.Close()

	fmtr := formatter.NewFormatter(f)

	log.Info().Msg("writing schema.graphl to file")

	fmtr.FormatSchema(schema)
}
