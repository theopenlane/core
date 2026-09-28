//go:build generate

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// ErrDestructiveSurfaceChange indicates a credential slot, operation, or webhook was removed without a registration declaring that it replaces it
var ErrDestructiveSurfaceChange = errors.New("integration schema: removed credential slot, operation, or webhook has no replacing registration")

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

// generateIntegrationSchemas writes one surface snapshot per registered definition, refusing only a removed slot, operation, or webhook that no registration declares it replaces
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

// gateSurfaceChange prints each change from the committed snapshot with what the runtime does about it, failing only when a removed slot, operation, or webhook has no replacing registration
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

// classifySurfaceChange returns one line per change between a committed surface and the next one, with the refusal when a removed slot, operation, or webhook has no replacing registration
func classifySurfaceChange(old, next registry.Surface) ([]string, error) {
	var (
		findings []string
		refused  error
	)

	for _, credential := range old.Credentials {
		nextCredential, kept := lo.Find(next.Credentials, func(candidate registry.SurfaceCredential) bool { return candidate.Ref == credential.Ref })
		if kept {
			slotFindings, err := classifySchemaChange(fmt.Sprintf("%s: credential slot %s", next.ID, credential.Ref), credential.SurfaceSchema, nextCredential.SurfaceSchema)
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
			lo.Ternary(replaced, "stored payloads convert to "+taker.Ref, "existing installations cannot resolve it; declare .Replacing on the slot ref that takes it over")))
	}

	inputFindings, err := classifyUserInputChange(next.ID, old.UserInput, next.UserInput)
	if err != nil {
		return findings, err
	}

	findings = append(findings, inputFindings...)

	for _, connection := range old.Connections {
		if !lo.Contains(next.Connections, connection) {
			findings = append(findings, fmt.Sprintf("%s: connection %s removed: %s", next.ID, connection, outcomeErrored))
		}
	}

	findings = append(findings, classifyMetadataChange(next.ID, old.Metadata, next.Metadata)...)

	for _, operation := range old.Operations {
		if lo.ContainsBy(next.Operations, func(candidate registry.SurfaceNamed) bool { return candidate.Name == operation.Name }) {
			continue
		}

		taker, replaced := lo.Find(next.Operations, func(candidate registry.SurfaceNamed) bool {
			return lo.Contains(candidate.Replaces, operation.Name)
		})
		if !replaced {
			refused = fmt.Errorf("%w: %s", ErrDestructiveSurfaceChange, next.ID)
		}

		findings = append(findings, fmt.Sprintf("%s: operation %s removed: %s", next.ID, operation.Name,
			lo.Ternary(replaced, "run history and health move to "+taker.Name, "its run history and health are orphaned; declare .Replacing on the operation ref that takes it over")))
	}

	for _, webhook := range old.Webhooks {
		nextWebhook, kept := lo.Find(next.Webhooks, func(candidate registry.SurfaceWebhook) bool { return candidate.Name == webhook.Name })
		if kept {
			findings = append(findings, classifyWebhookEventChange(next.ID, webhook, nextWebhook)...)

			continue
		}

		taker, replaced := lo.Find(next.Webhooks, func(candidate registry.SurfaceWebhook) bool {
			return lo.Contains(candidate.Replaces, webhook.Name)
		})
		if !replaced {
			refused = fmt.Errorf("%w: %s", ErrDestructiveSurfaceChange, next.ID)
		}

		findings = append(findings, fmt.Sprintf("%s: webhook %s removed: %s", next.ID, webhook.Name,
			lo.Ternary(replaced, "endpoint row renamed to "+taker.Name, "its endpoint row is deleted as stale and its endpoint id lost; declare .Replacing on the webhook ref that takes it over")))
	}

	return findings, refused
}

// classifyUserInputChange applies the property rules to the committed and next user input schemas, treating an added user input as a change from an empty schema
func classifyUserInputChange(definitionID string, old, next *registry.SurfaceSchema) ([]string, error) {
	label := definitionID + ": user input"

	switch {
	case old == nil && next == nil:
		return nil, nil
	case next == nil:
		return []string{label + " removed: stored config is left untouched"}, nil
	case old == nil:
		findings, err := classifySchemaChange(label, registry.SurfaceSchema{}, *next)

		return append([]string{label + " added: stored config is conformed on upgrade"}, findings...), err
	default:
		return classifySchemaChange(label, *old, *next)
	}
}

// classifyMetadataChange reports each connection whose derived metadata schema was added, removed, or changed
func classifyMetadataChange(definitionID string, old, next []registry.SurfaceConnectionMetadata) []string {
	var findings []string

	for _, metadata := range old {
		nextMetadata, kept := lo.Find(next, func(candidate registry.SurfaceConnectionMetadata) bool {
			return candidate.Connection == metadata.Connection
		})

		switch {
		case !kept:
			findings = append(findings, fmt.Sprintf("%s: connection %s metadata schema removed: re-derived on upgrade", definitionID, metadata.Connection))
		case !sameJSON(metadata.Schema, nextMetadata.Schema):
			findings = append(findings, fmt.Sprintf("%s: connection %s metadata schema changed: re-derived on upgrade", definitionID, metadata.Connection))
		}
	}

	for _, metadata := range next {
		if !lo.ContainsBy(old, func(candidate registry.SurfaceConnectionMetadata) bool {
			return candidate.Connection == metadata.Connection
		}) {
			findings = append(findings, fmt.Sprintf("%s: connection %s metadata schema added: re-derived on upgrade", definitionID, metadata.Connection))
		}
	}

	return findings
}

// classifyWebhookEventChange reports each event removed from a kept webhook contract; only the endpoint row's allowed events bind to event names
func classifyWebhookEventChange(definitionID string, old, next registry.SurfaceWebhook) []string {
	var findings []string

	for _, event := range old.Events {
		if lo.ContainsBy(next.Events, func(candidate registry.SurfaceNamed) bool { return candidate.Name == event.Name }) {
			continue
		}

		taker, replaced := lo.Find(next.Events, func(candidate registry.SurfaceNamed) bool {
			return lo.Contains(candidate.Replaces, event.Name)
		})

		findings = append(findings, fmt.Sprintf("%s: webhook %s event %s removed: %s", definitionID, old.Name, event.Name,
			lo.Ternary(replaced, "replaced by "+taker.Name+"; allowed events refreshed on upgrade", "allowed events refreshed on upgrade")))
	}

	return findings
}

// sameJSON reports whether two raw documents decode to the same value regardless of encoding
func sameJSON(a, b json.RawMessage) bool {
	var left, right any

	if err := json.Unmarshal(a, &left); err != nil {
		return false
	}

	if err := json.Unmarshal(b, &right); err != nil {
		return false
	}

	return reflect.DeepEqual(left, right)
}

// classifySchemaChange applies the property rules to one kind's committed and next schemas under the given finding label
func classifySchemaChange(label string, old, next registry.SurfaceSchema) ([]string, error) {
	oldRoot, _, err := jsonx.SchemaRoot(old.Schema)
	if err != nil {
		return nil, fmt.Errorf("%s committed schema: %w", label, err)
	}

	nextRoot, _, err := jsonx.SchemaRoot(next.Schema)
	if err != nil {
		return nil, fmt.Errorf("%s next schema: %w", label, err)
	}

	if oldRoot.Properties == nil {
		oldRoot.Properties = jsonschema.NewProperties()
	}

	if nextRoot.Properties == nil {
		nextRoot.Properties = jsonschema.NewProperties()
	}

	var findings []string

	for pair := oldRoot.Properties.Oldest(); pair != nil; pair = pair.Next() {
		nextProperty, kept := nextRoot.Properties.Get(pair.Key)
		if !kept {
			findings = append(findings, fmt.Sprintf("%s property %s removed: stored value is dropped on upgrade", label, pair.Key))

			continue
		}

		if pair.Value.Type != nextProperty.Type {
			findings = append(findings, fmt.Sprintf("%s property %s type changed from %q to %q: %s", label, pair.Key, pair.Value.Type, nextProperty.Type, outcomeErrored))
		}

		if len(pair.Value.Enum) > 0 && len(nextProperty.Enum) > 0 && !lo.Every(nextProperty.Enum, pair.Value.Enum) {
			findings = append(findings, fmt.Sprintf("%s property %s enum narrowed: %s", label, pair.Key, outcomeErrored))
		}
	}

	for _, name := range nextRoot.Required {
		property, declared := nextRoot.Properties.Get(name)
		if lo.Contains(oldRoot.Required, name) || (declared && property.Default != nil) {
			continue
		}

		findings = append(findings, fmt.Sprintf("%s required property %s added without default: %s", label, name, lo.Ternary(next.Backfill, "backfilled on upgrade", outcomeErrored)))
	}

	return findings, nil
}
