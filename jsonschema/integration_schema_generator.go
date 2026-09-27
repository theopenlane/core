//go:build generate

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// ErrDestructiveSurfaceChange indicates a credential slot was removed without a slot declaring that it replaces it
var ErrDestructiveSurfaceChange = errors.New("integration schema: removed credential slot has no replacing slot")

const (
	// integrationSchemaDir is the directory holding one committed surface snapshot per definition
	integrationSchemaDir = "./jsonschema/integrations"
	// snapshotExtension is the file extension of every surface snapshot
	snapshotExtension = ".json"
	// snapshotFilePermission is the mode used for written snapshot files
	snapshotFilePermission = 0600
	// snapshotDirPermission is the mode used when creating the snapshot directory
	snapshotDirPermission = 0755
	// outcomeErrored is reported when the upgrade of an existing installation fails on its first use
	outcomeErrored = "existing installations are marked errored on first use until reconnected"
)

// main snapshots the installation-facing surface of every catalog definition, exiting non-zero on failure
func main() {
	if err := generateIntegrationSchemas(integrationSchemaDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// generateIntegrationSchemas writes one surface snapshot per registered definition, refusing only a removed slot that no slot declares it replaces
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
			if err := gateSurfaceChange(target, existing, next); err != nil {
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

// gateSurfaceChange prints each change from the committed snapshot with what the runtime does about it, failing only when a removed slot has no replacing slot
func gateSurfaceChange(target string, existing []byte, next registry.Surface) error {
	var old registry.Surface
	if err := json.Unmarshal(existing, &old); err != nil {
		return fmt.Errorf("decode %s: %w", target, err)
	}

	findings, err := classifySurfaceChange(old, next)

	for _, finding := range findings {
		fmt.Println(finding)
	}

	return err
}

// classifySurfaceChange returns one line per change between a committed surface and the next one, with the refusal when a removed slot has no replacing slot
func classifySurfaceChange(old, next registry.Surface) ([]string, error) {
	var (
		findings []string
		refused  error
	)

	for _, credential := range old.Credentials {
		nextCredential, kept := lo.Find(next.Credentials, func(candidate registry.SurfaceCredential) bool { return candidate.Ref == credential.Ref })
		if kept {
			slotFindings, err := classifySchemaChange(next.ID, credential, nextCredential)
			if err != nil {
				return findings, err
			}

			findings = append(findings, slotFindings...)

			continue
		}

		taker, replaced := lo.Find(next.Credentials, func(candidate registry.SurfaceCredential) bool {
			return lo.Contains(candidate.Replaces, credential.Ref)
		})
		if !replaced {
			refused = fmt.Errorf("%w: %s", ErrDestructiveSurfaceChange, next.ID)
		}

		findings = append(findings, fmt.Sprintf("%s: credential slot %s removed: %s", next.ID, credential.Ref,
			lo.Ternary(replaced, "stored payloads convert to "+taker.Ref, "existing installations cannot resolve it; declare types.Replacing on the slot that takes it over")))
	}

	for _, connection := range old.Connections {
		if !lo.Contains(next.Connections, connection) {
			findings = append(findings, fmt.Sprintf("%s: connection %s removed: %s", next.ID, connection, outcomeErrored))
		}
	}

	return findings, refused
}

// classifySchemaChange applies the property rules to one slot's committed and next schemas
func classifySchemaChange(definitionID string, old, next registry.SurfaceCredential) ([]string, error) {
	slot := fmt.Sprintf("%s: credential slot %s", definitionID, old.Ref)

	oldRoot, _, err := jsonx.SchemaRoot(old.Schema)
	if err != nil {
		return nil, fmt.Errorf("%s committed schema: %w", slot, err)
	}

	nextRoot, _, err := jsonx.SchemaRoot(next.Schema)
	if err != nil {
		return nil, fmt.Errorf("%s next schema: %w", slot, err)
	}

	if nextRoot.Properties == nil {
		nextRoot.Properties = jsonschema.NewProperties()
	}

	var findings []string

	for pair := oldRoot.Properties.Oldest(); pair != nil; pair = pair.Next() {
		nextProperty, kept := nextRoot.Properties.Get(pair.Key)
		if !kept {
			findings = append(findings, fmt.Sprintf("%s property %s removed: stored value is dropped on upgrade", slot, pair.Key))

			continue
		}

		if pair.Value.Type != nextProperty.Type {
			findings = append(findings, fmt.Sprintf("%s property %s type changed from %q to %q: %s", slot, pair.Key, pair.Value.Type, nextProperty.Type, outcomeErrored))
		}

		if len(pair.Value.Enum) > 0 && len(nextProperty.Enum) > 0 && !lo.Every(nextProperty.Enum, pair.Value.Enum) {
			findings = append(findings, fmt.Sprintf("%s property %s enum narrowed: %s", slot, pair.Key, outcomeErrored))
		}
	}

	for _, name := range nextRoot.Required {
		property, declared := nextRoot.Properties.Get(name)
		if lo.Contains(oldRoot.Required, name) || (declared && property.Default != nil) {
			continue
		}

		findings = append(findings, fmt.Sprintf("%s required property %s added without default: %s", slot, name, lo.Ternary(next.Backfill, "backfilled on upgrade", outcomeErrored)))
	}

	return findings, nil
}
