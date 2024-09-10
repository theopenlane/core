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

	if err := api.Generate(cfg,
		api.ReplacePlugin(resolvergen.New()), // replace the resolvergen plugin
		api.AddPlugin(bulkgen.New()),         // add the bulkgen plugin
		api.AddPlugin(searchgen.New()),       // add the search plugin
	); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
}
