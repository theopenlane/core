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

	// checksum files to track schema changes
	clientChecksumFile = "./internal/graphapi/testclient/.client_checksum"
)

var (
	// changes to these paths should trigger full client generation
	inputPaths = []string{
		"internal/graphapi/clientschema",
		"internal/graphapi/historyschema",
		graphapiGenDir,
	}
)

func main() {
	genhelpers.SetupLogging()

	genhelpers.ChangeToRootDir("../../../")

	hasChanges, err := genhelpers.HasSchemaChanges(clientChecksumFile, inputPaths...)
	if err != nil {
		log.Warn().Err(err).Msg("failed to check for schema changes, running history generation anyway")
		hasChanges = true
	}

	if !hasChanges {
		log.Info().Msg("no schema changes detected, skipping gqlgen server generation")
		return
	}

	cfg, err := config.LoadConfig(graphapiGenDir + ".gqlgenc_testclient.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
		os.Exit(2)
	}

	if err := generator.Generate(context.Background(), cfg); err != nil {
		log.Error().Err(err).Msg("Failed to generate gqlgenc client")
	}

	// update checksum file
	genhelpers.SetSchemaChecksum(clientChecksumFile, inputPaths...)
}
