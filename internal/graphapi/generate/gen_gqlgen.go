// go:build ignore

package main

import (
	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/gqlgen-plugins/bulkgen"
	"github.com/theopenlane/gqlgen-plugins/fieldgen"
	"github.com/theopenlane/gqlgen-plugins/resolvergen"
	"github.com/theopenlane/gqlgen-plugins/searchgen"

	"github.com/theopenlane/core/internal/genhelpers"
)

const (
	graphapiGenDir = "internal/graphapi/generate/"
)

// extraFields is a list of fields to add to the schema
// dynamically. This is useful for adding fields that are
// not part of the ent db schema but are needed in the graphql
// schema (used in conjunction with the additional fields in ent)
var extraFields = []fieldgen.AdditionalField{
	{
		Name:                         "createdBy",
		CustomType:                   "Actor",
		NonNull:                      true,
		Description:                  "The user or service who created the object",
		AddToSchemaWithExistingField: "createdByID",
	},
	{
		Name:                         "updatedBy",
		CustomType:                   "Actor",
		NonNull:                      true,
		Description:                  "The user or service who last updated the object",
		AddToSchemaWithExistingField: "updatedByID",
	},
}

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
		api.AddPlugin(fieldgen.NewExtraFieldsGen(extraFields)), // add the fieldgen plugin
		api.ReplacePlugin(resolvergen.New()),                   // replace the resolvergen plugin
		api.AddPlugin(bulkgen.NewWithOptions(
			bulkgen.WithModelPackage(modelImport),
			bulkgen.WithEntGeneratedPackage(entPackage),
		)), // add the bulkgen plugin
		api.AddPlugin(searchgen.NewWithOptions(
			searchgen.WithEntGeneratedPackage(entPackage),
			searchgen.WithModelPackage(modelImport),
		)), // add the search plugin
	); err != nil {
		log.Fatal().Err(err).Msg("failed to generate gqlgen server")
	}
}
