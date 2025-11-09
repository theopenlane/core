package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/theopenlane/core/internal/httpserve/server"
)

const (
	specDir        = "internal/httpserve/specs"
	jsonFilename   = "openlane.openapi.json"
	yamlFilename   = "openlane.openapi.yaml"
	jsonOutputPath = specDir + "/" + jsonFilename
	yamlOutputPath = specDir + "/" + yamlFilename

	specFilePerm fs.FileMode = 0o600
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("failed to generate OpenAPI specifications")
	}
}

func run() error {
	spec, err := server.GenerateOpenAPISpecDocument()
	if err != nil {
		return fmt.Errorf("build OpenAPI spec: %w", err)
	}

	jsonData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal spec to json: %w", err)
	}

	if err := os.WriteFile(jsonOutputPath, jsonData, specFilePerm); err != nil {
		return fmt.Errorf("write json spec: %w", err)
	}

	var intermediate any
	if err := json.Unmarshal(jsonData, &intermediate); err != nil {
		return fmt.Errorf("prepare data for yaml: %w", err)
	}

	yamlData, err := yaml.Marshal(intermediate)
	if err != nil {
		return fmt.Errorf("marshal spec to yaml: %w", err)
	}

	if err := os.WriteFile(yamlOutputPath, yamlData, specFilePerm); err != nil {
		return fmt.Errorf("write yaml spec: %w", err)
	}

	return nil
}
