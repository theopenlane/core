//go:build generate

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
)

const (
	// integrationSchemaDir is the directory holding one committed surface snapshot per definition
	integrationSchemaDir = "./jsonschema/integrations"
	// snapshotExtension is the file extension of every surface snapshot
	snapshotExtension = ".json"
	// snapshotFilePermission is the mode used for written snapshot files
	snapshotFilePermission = 0600
	// snapshotDirPermission is the mode used when creating the snapshot directory
	snapshotDirPermission = 0755
)

// main snapshots the installation-facing surface of every catalog definition, exiting non-zero on failure
func main() {
	if err := generateIntegrationSchemas(integrationSchemaDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generateIntegrationSchemas writes one surface snapshot per registered definition, refusing a removed slot, operation, webhook, or webhook event that no registration declares it replaces
func generateIntegrationSchemas(dir string) error {
	reg := registry.New()
	if err := reg.RegisterAll(catalog.Builders(catalog.Config{}, "", false)...); err != nil {
		return err
	}

	if err := os.MkdirAll(dir, snapshotDirPermission); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	for _, def := range reg.Definitions() {
		next := registry.DefinitionSurface(def)
		target := filepath.Join(dir, def.ID+snapshotExtension)

		existing, err := os.ReadFile(target)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("read %s: %w", target, err)
		}

		if err == nil {
			if err := registry.GateSurfaceChange(target, existing, next); err != nil {
				return err
			}
		}

		data, err := json.MarshalIndent(next, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s surface: %w", def.ID, err)
		}

		if err := os.WriteFile(target, append(data, '\n'), snapshotFilePermission); err != nil {
			return fmt.Errorf("write %s: %w", target, err)
		}
	}

	return nil
}
