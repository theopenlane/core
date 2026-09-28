package registry

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// outcomeErrored is reported when the upgrade of an existing installation fails on its first use
const outcomeErrored = "existing installations are marked errored on first use until reconnected"

// GateSurfaceChange prints each change from the committed snapshot with what the runtime does about it, failing only when a removed slot, operation, or webhook has no replacing registration
func GateSurfaceChange(target string, existing []byte, next Surface) error {
	var old Surface
	if err := json.Unmarshal(existing, &old); err != nil {
		return fmt.Errorf("decode %s: %w", target, err)
	}

	findings, err := ClassifySurfaceChange(old, next)

	for _, finding := range findings {
		fmt.Println(finding)
	}

	return err
}

// ClassifySurfaceChange returns one line per change between a committed surface and the next one, with the refusal when a removed slot, operation, or webhook has no replacing registration
func ClassifySurfaceChange(old, next Surface) ([]string, error) {
	var (
		findings []string
		refused  error
	)

	for _, credential := range old.Credentials {
		nextCredential, kept := lo.Find(next.Credentials, func(candidate SurfaceCredential) bool { return candidate.Ref == credential.Ref })
		if kept {
			slotFindings, err := classifySchemaChange(fmt.Sprintf("%s: credential slot %s", next.ID, credential.Ref), credential.SurfaceSchema, nextCredential.SurfaceSchema)
			if err != nil {
				return findings, err
			}

			findings = append(findings, slotFindings...)

			continue
		}

		taker, replaced := lo.Find(next.Credentials, func(candidate SurfaceCredential) bool {
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

	findings = append(findings, classifyInstallationChange(next.ID, old.Installation, next.Installation)...)

	for _, operation := range old.Operations {
		nextOperation, kept := lo.Find(next.Operations, func(candidate SurfaceOperation) bool { return candidate.Name == operation.Name })
		if kept {
			configFindings, err := classifySchemaChange(fmt.Sprintf("%s: operation %s config", next.ID, operation.Name), SurfaceSchema{Schema: operation.Schema}, SurfaceSchema{Schema: nextOperation.Schema})
			if err != nil {
				return findings, err
			}

			findings = append(findings, configFindings...)

			continue
		}

		taker, replaced := lo.Find(next.Operations, func(candidate SurfaceOperation) bool {
			return lo.Contains(candidate.Replaces, operation.Name)
		})
		if !replaced {
			refused = fmt.Errorf("%w: %s", ErrDestructiveSurfaceChange, next.ID)
		}

		findings = append(findings, fmt.Sprintf("%s: operation %s removed: %s", next.ID, operation.Name,
			lo.Ternary(replaced, "run history and health move to "+taker.Name, "its run history and health are orphaned; declare .Replacing on the operation ref that takes it over")))
	}

	for _, webhook := range old.Webhooks {
		nextWebhook, kept := lo.Find(next.Webhooks, func(candidate SurfaceWebhook) bool { return candidate.Name == webhook.Name })
		if kept {
			findings = append(findings, classifyWebhookEventChange(next.ID, webhook, nextWebhook)...)

			continue
		}

		taker, replaced := lo.Find(next.Webhooks, func(candidate SurfaceWebhook) bool {
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
func classifyUserInputChange(definitionID string, old, next *SurfaceSchema) ([]string, error) {
	label := definitionID + ": user input"

	switch {
	case old == nil && next == nil:
		return nil, nil
	case next == nil:
		return []string{label + " removed: stored config is left untouched"}, nil
	case old == nil:
		findings, err := classifySchemaChange(label, SurfaceSchema{}, *next)

		return append([]string{label + " added: stored config is conformed on upgrade"}, findings...), err
	default:
		return classifySchemaChange(label, *old, *next)
	}
}

// classifyInstallationChange reports a definition whose derived installation metadata schema was added, removed, or changed
func classifyInstallationChange(definitionID string, old, next *SurfaceSchema) []string {
	label := definitionID + ": installation metadata schema"

	switch {
	case old == nil && next == nil:
		return nil
	case old == nil:
		return []string{label + " added: re-derived on upgrade"}
	case next == nil:
		return []string{label + " removed: re-derived on upgrade"}
	case !sameJSON(old.Schema, next.Schema):
		return []string{label + " changed: re-derived on upgrade"}
	default:
		return nil
	}
}

// classifyWebhookEventChange reports each event removed from a kept webhook contract; only the endpoint row's allowed events bind to event names
func classifyWebhookEventChange(definitionID string, old, next SurfaceWebhook) []string {
	var findings []string

	for _, event := range old.Events {
		if lo.ContainsBy(next.Events, func(candidate SurfaceNamed) bool { return candidate.Name == event.Name }) {
			continue
		}

		taker, replaced := lo.Find(next.Events, func(candidate SurfaceNamed) bool {
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
func classifySchemaChange(label string, old, next SurfaceSchema) ([]string, error) {
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
