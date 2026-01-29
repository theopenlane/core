//go:build ignore

package main

import (
	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/schema"
	"github.com/theopenlane/core/internal/genhelpers"
	"github.com/theopenlane/gqlgen-plugins/bulkgen"
	"github.com/theopenlane/gqlgen-plugins/resolvergen"
	"github.com/theopenlane/gqlgen-plugins/searchgen"
)

const (
	// graphapiGenDir is the directory where the configuration file for gqlgen is located
	graphapiGenDir = "internal/graphapi/generate/"
	// csvDir is the directory where the CSV files will be stored for example bulk operations
	csvDir = "internal/httpserve/handlers/csv"
	// graphqlImport that includes the transaction wrappers and other common graphql helpers
	graphqlImport = "github.com/theopenlane/core/internal/graphapi/common"

	// checksum files to track schema changes
	schemaChecksumFile  = "./internal/graphapi/checksum/.schema_checksum"
	historyChecksumFile = "./internal/graphapi/checksum/.history_schema_checksum"
)

var (
	// changes to these paths should trigger full schema generation
	mainInputPaths = []string{
		"internal/graphapi/schema",
		"internal/ent/generated",
		graphapiGenDir,
	}

	// changes to these paths should trigger history schema generation
	historyInputPaths = []string{
		"internal/graphapi/historyschema",
		"internal/ent/historygenerated",
		graphapiGenDir,
	}
)

func main() {
	genhelpers.SetupLogging()

	// change to the root of the repo so that the config hierarchy is correct
	genhelpers.ChangeToRootDir("../../../")

	gqlGenerate()
	gqlHistoryGenerate()
}

func gqlGenerate() {
	// check if there were schema changes before running full codegen
	hasChanges, err := genhelpers.HasSchemaChanges(schemaChecksumFile, mainInputPaths...)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for schema changes, running history generation anyway")
		hasChanges = true
	}

	if !hasChanges {
		log.Info().Msg("no schema changes detected, skipping gqlgen server generation")
		return
	}

	cfg, err := config.LoadConfig(graphapiGenDir + ".gqlgen.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	modelImport := "github.com/theopenlane/core/internal/graphapi/model"
	entPackage := "github.com/theopenlane/core/internal/ent/generated"
	csvGeneratedPackage := "github.com/theopenlane/core/internal/ent/csvgenerated"
	rulePackage := "github.com/theopenlane/core/internal/ent/privacy/rule"

	if err := api.Generate(cfg,
		api.ReplacePlugin(resolvergen.NewWithOptions(
			resolvergen.WithEntGeneratedPackage(entPackage),
			resolvergen.WithArchivableSchemas([]string{schema.Program{}.Name()}),
			resolvergen.WithGraphQLImport(graphqlImport),
			resolvergen.WithCSVGeneratedPackage(csvGeneratedPackage),
			resolvergen.WithForceRegenerateBulkResolvers(false),
		)), // replace the resolvergen plugin
		api.AddPlugin(bulkgen.NewWithOptions(
			bulkgen.WithModelPackage(modelImport),
			bulkgen.WithEntGeneratedPackage(entPackage),
			bulkgen.WithCSVOutputPath(csvDir),
			bulkgen.WithGraphQLImport(graphqlImport),
			bulkgen.WithCSVGeneratedPackage(csvGeneratedPackage),
		)), // add the bulkgen plugin
		api.AddPlugin(searchgen.NewWithOptions(
			searchgen.WithEntGeneratedPackage(entPackage),
			searchgen.WithModelPackage(modelImport),
			searchgen.WithRulePackage(rulePackage),
			searchgen.WithIncludeAdminSearch(false),
			searchgen.WithGraphQLImport(graphqlImport),
		)), // add the search plugin
	); err != nil {
		log.Fatal().Err(err).Msg("failed to generate gqlgen server")
	}

	// update checksum file
	log.Info().Msg("updating schema checksum file after successful generation")
	genhelpers.SetSchemaChecksum(schemaChecksumFile, mainInputPaths...)
}

func gqlHistoryGenerate() {
	// check if there were schema changes before running full codegen
	hasChanges, err := genhelpers.HasSchemaChanges(historyChecksumFile, historyInputPaths...)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for history schema changes, running history generation anyway")
		hasChanges = true
	}

	if !hasChanges {
		log.Info().Msg("no history schema changes detected, skipping gqlgen history server generation")
		return
	}

	cfg, err := config.LoadConfig(graphapiGenDir + ".gqlgen_history.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	entPackage := "github.com/theopenlane/core/internal/ent/historygenerated"

	if err := api.Generate(cfg,
		api.ReplacePlugin(resolvergen.NewWithOptions(
			resolvergen.WithEntGeneratedPackage(entPackage),
			resolvergen.WithGraphQLImport(graphqlImport),
		)),
	); err != nil {
		log.Fatal().Err(err).Msg("failed to generate gqlgen history server")
	}

	// update checksum file
	log.Info().Msg("updating history schema checksum file after successful generation")
	genhelpers.SetSchemaChecksum(historyChecksumFile, historyInputPaths...)
}
