// go:build ignore

package main

import (
	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/gqlgen-plugins/bulkgen"
	"github.com/theopenlane/gqlgen-plugins/resolvergen"
	"github.com/theopenlane/gqlgen-plugins/searchgen"

	"github.com/theopenlane/core/internal/genhelpers"
)

const (
	graphapiGenDir = "internal/graphapi/generate/"
	csvDir         = "internal/httpserve/handlers/csv"
)

func main() {
	genhelpers.SetupLogging()

	// change to the root of the repo so that the config hierarchy is correct
	genhelpers.ChangeToRootDir("../../../")

	cfg, err := config.LoadConfig(graphapiGenDir + ".gqlgen.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	modelImport := "github.com/theopenlane/core/internal/graphapi/model"
	entPackage := "github.com/theopenlane/core/internal/ent/generated"

	if err := api.Generate(cfg,
		api.ReplacePlugin(resolvergen.NewWithOptions(resolvergen.WithEntGeneratedPackage(
			entPackage,
		))), // replace the resolvergen plugin
		api.AddPlugin(bulkgen.NewWithOptions(
			bulkgen.WithModelPackage(modelImport),
			bulkgen.WithEntGeneratedPackage(entPackage),
			bulkgen.WithCSVOutputPath(csvDir),
		)), // add the bulkgen plugin
		api.AddPlugin(searchgen.NewWithOptions(
			searchgen.WithEntGeneratedPackage(entPackage),
			searchgen.WithModelPackage(modelImport),
		)), // add the search plugin
	); err != nil {
		log.Fatal().Err(err).Msg("failed to generate gqlgen server")
	}
}
