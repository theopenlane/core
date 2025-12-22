//go:build ignore

package main

import (
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"

	"github.com/vektah/gqlparser/v2/formatter"

	"github.com/theopenlane/core/internal/genhelpers"
	gqlgenerated "github.com/theopenlane/core/internal/graphapi/generated"
	gqlhistorygenerated "github.com/theopenlane/core/internal/graphapi/generated"
)

const (
	graphapiGenDir = "internal/graphapi/generate/"

	// checksum files to track schema changes
	schemaChecksumFile  = "./internal/graphapi/clientschema/checksum/.schema_checksum"
	historyChecksumFile = "./internal/graphapi/historyschema/checksum/.history_schema_checksum"
)

var (
	// changes to these paths should trigger full schema generation
	mainInputPaths = []string{
		"internal/graphapi/generated",
		"internal/graphapi/model",
		graphapiGenDir,
	}

	// changes to these paths should trigger history schema generation
	historyInputPaths = []string{
		"internal/graphapi/historygenerated",
		"internal/graphapi/historymodel",
		graphapiGenDir,
	}
)

// read in schema from internal package and save it to the schema file
func main() {
	genhelpers.SetupLogging()

	genhelpers.ChangeToRootDir("../../../")

	_, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get current directory")
	}

	// check if there were schema changes before running full codegen
	hasChanges, err := genhelpers.HasSchemaChanges(schemaChecksumFile, mainInputPaths...)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for schema changes, running history generation anyway")
		hasChanges = true
	}

	if hasChanges {
		log.Info().Msg("Generating schema for client")
		execSchema := gqlgenerated.NewExecutableSchema(gqlgenerated.Config{})

		generateClientSchema(execSchema, "internal/graphapi/clientschema/schema.graphql")
	} else {
		log.Info().Msg("no schema changes detected, skipping gqlgen schema generation")
	}

	// check if there were schema changes before running full codegen
	hasHistoryChanges, err := genhelpers.HasSchemaChanges(historyChecksumFile, historyInputPaths...)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for history schema changes, running history generation anyway")
		hasHistoryChanges = true
	}

	if hasHistoryChanges {
		log.Info().Msg("Generating history schema for client")
		execHistorySchema := gqlhistorygenerated.NewExecutableSchema(gqlhistorygenerated.Config{})
		generateClientSchema(execHistorySchema, "internal/graphapi/historyschema/schema.graphql")
	} else {
		log.Info().Msg("no history schema changes detected, skipping gqlgen history schema generation")
	}

	log.Info().Msg("Finished schema generation, saving updated checksums")

	// update checksums if there were changes
	if hasChanges {
		genhelpers.SetSchemaChecksum(schemaChecksumFile, mainInputPaths...)
	}

	if hasHistoryChanges {
		genhelpers.SetSchemaChecksum(historyChecksumFile, historyInputPaths...)
	}
}

func generateClientSchema(execSchema graphql.ExecutableSchema, path string) {
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
