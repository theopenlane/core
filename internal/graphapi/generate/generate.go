// go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
	"github.com/theopenlane/gqlgen-plugins/bulkgen"
	"github.com/theopenlane/gqlgen-plugins/resolvergen"
	"github.com/theopenlane/gqlgen-plugins/searchgen"
)

func main() {
	cfg, err := config.LoadConfigFromDefaultLocations()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config", err.Error())
		os.Exit(2)
	}

	modelImport := "github.com/theopenlane/core/internal/graphapi/model"
	entPackage := "github.com/theopenlane/core/internal/ent/generated"

	if err := api.Generate(cfg,
		api.ReplacePlugin(resolvergen.New()), // replace the resolvergen plugin
		api.AddPlugin(bulkgen.NewWithOptions(
			bulkgen.WithModelPackage(modelImport),
			bulkgen.WithEntGeneratedPackage(entPackage),
		)), // add the bulkgen plugin
		api.AddPlugin(searchgen.NewWithOptions(
			searchgen.WithEntGeneratedPackage(entPackage),
			searchgen.WithModelPackage(modelImport),
		)), // add the search plugin
	); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
}
