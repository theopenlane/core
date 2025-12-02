package main

import (
	"context"
	"os"

	"github.com/Yamashou/gqlgenc/config"
	"github.com/Yamashou/gqlgenc/generator"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/shared/logx"
)

const (
	graphapiGenDir = "./"
)

func main() {
	logx.Configure(logx.LoggerConfig{
		Level:         zerolog.DebugLevel,
		Pretty:        true,
		Writer:        os.Stderr,
		IncludeCaller: true,
		SetGlobal:     true,
	})

	cfg, err := config.LoadConfig(graphapiGenDir + ".gqlgenc.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
		os.Exit(2)
	}

	if err := generator.Generate(context.Background(), cfg); err != nil {
		log.Error().Err(err).Msg("Failed to generate gqlgenc client")
	}
}
