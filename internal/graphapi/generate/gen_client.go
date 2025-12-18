//go:build ignore

package main

import (
	"context"
	"os"

	"github.com/Yamashou/gqlgenc/config"
	"github.com/Yamashou/gqlgenc/generator"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/core/internal/genhelpers"
)

const (
	graphapiGenDir = "internal/graphapi/generate/"
)

func main() {
	genhelpers.SetupLogging()

	genhelpers.ChangeToRootDir("../../../")

	cfg, err := config.LoadConfig(graphapiGenDir + ".gqlgenc_testclient.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
		os.Exit(2)
	}

	if err := generator.Generate(context.Background(), cfg); err != nil {
		log.Error().Err(err).Msg("Failed to generate gqlgenc client")
	}
}
