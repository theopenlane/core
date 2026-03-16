package scim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	scimresource "github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationscim "github.com/theopenlane/core/internal/integrations/definitions/scim"
	integrationops "github.com/theopenlane/core/internal/integrations/operations"
	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

const scimDeleteAction = "delete"

// ingestDirectoryPayloadSets routes SCIM directory payloads through the shared Runtime ingest path
func ingestDirectoryPayloadSets(ctx context.Context, client *generated.Client, ic *IntegrationContext, payloadSets []integrationtypes.IngestPayloadSet) error {
	if ic == nil || ic.Runtime == nil {
		return integrationsruntime.ErrInstallationRequired
	}

	installation, err := scimRuntimeInstallation(ic)
	if err != nil {
		return err
	}

	syncRunID, err := ensureScimSyncRun(ctx, client, installation.ID, installation.OwnerID)
	if err != nil {
		return err
	}

	return ic.Runtime.IngestPayloadSets(
		ctx,
		client,
		installation,
		integrationscim.DirectorySyncOperation.Name(),
		payloadSets,
		integrationsruntime.IngestOptions{
			DirectorySyncRunID:               syncRunID,
			SkipDirectorySyncRunFinalization: true,
		},
	)
}

// scimRuntimeInstallation normalizes legacy SCIM integrations so Runtime can resolve the definition
func scimRuntimeInstallation(ic *IntegrationContext) (*generated.Integration, error) {
	if ic == nil || ic.Installation == nil {
		return nil, integrationsruntime.ErrInstallationRequired
	}

	installation := *ic.Installation
	if installation.DefinitionID == "" {
		installation.DefinitionID = integrationscim.DefinitionID.ID()
	}

	if installation.DefinitionSlug == "" {
		installation.DefinitionSlug = integrationscim.Slug
	}

	return &installation, nil
}

// handleDirectoryIngestError maps shared ingest failures to SCIM-compatible errors
func handleDirectoryIngestError(err error, detail string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrInvalidAttributes):
		return scimerrors.ScimError{
			ScimType: scimerrors.ScimTypeInvalidValue,
			Detail:   detail,
			Status:   400,
		}
	case errors.Is(err, integrationops.ErrIngestMappedDocumentInvalid),
		errors.Is(err, integrationops.ErrIngestRequiredKeyMissing),
		errors.Is(err, integrationops.ErrIngestUpsertKeyMissing):
		return scimerrors.ScimError{
			ScimType: scimerrors.ScimTypeInvalidValue,
			Detail:   detail,
			Status:   400,
		}
	case errors.Is(err, integrationops.ErrIngestUpsertConflict):
		return scimerrors.ScimError{
			ScimType: scimerrors.ScimTypeUniqueness,
			Detail:   detail,
			Status:   409,
		}
	default:
		return fmt.Errorf("directory ingest failed: %w", err)
	}
}

// buildDirectoryAccountPayloadSet serializes one SCIM user payload into the shared directory-account ingest schema
func buildDirectoryAccountPayloadSet(attributes scimresource.ResourceAttributes, action string) (integrationtypes.IngestPayloadSet, error) {
	payload := cloneSCIMAttributes(attributes)
	externalID := directoryAccountExternalID(payload)
	if externalID == "" {
		return integrationtypes.IngestPayloadSet{}, ErrInvalidAttributes
	}

	payload["externalId"] = externalID

	return buildDirectoryPayloadSet(
		integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
		payload,
		externalID,
		action,
	)
}

// buildDirectoryGroupPayloadSets serializes one SCIM group payload and its memberships into shared ingest payload sets
func buildDirectoryGroupPayloadSets(attributes scimresource.ResourceAttributes, action string) ([]integrationtypes.IngestPayloadSet, error) {
	payload := cloneSCIMAttributes(attributes)
	externalID := directoryGroupExternalID(payload)
	if externalID == "" {
		return nil, ErrInvalidAttributes
	}

	payload["externalId"] = externalID

	groupPayloadSet, err := buildDirectoryPayloadSet(
		integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
		payload,
		externalID,
		action,
	)
	if err != nil {
		return nil, err
	}

	if action == scimDeleteAction {
		return []integrationtypes.IngestPayloadSet{groupPayloadSet}, nil
	}

	memberIDs := extractMemberIDsFromValue(payload["members"])
	if len(memberIDs) == 0 {
		return []integrationtypes.IngestPayloadSet{groupPayloadSet}, nil
	}

	membershipPayloads := make([]any, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		membershipPayloads = append(membershipPayloads, map[string]any{
			"group": map[string]any{
				"externalId": externalID,
			},
			"member": map[string]any{
				"value": memberID,
			},
		})
	}

	membershipPayloadSet, err := buildDirectoryPayloadSet(
		integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
		membershipPayloads,
		externalID,
		action,
	)
	if err != nil {
		return nil, err
	}

	return []integrationtypes.IngestPayloadSet{groupPayloadSet, membershipPayloadSet}, nil
}

// buildDirectoryPayloadSet wraps one or more SCIM payload documents in the shared ingest envelope shape
func buildDirectoryPayloadSet(schema string, payload any, resource string, action string) (integrationtypes.IngestPayloadSet, error) {
	switch value := payload.(type) {
	case []any:
		envelopes := make([]integrationtypes.MappingEnvelope, 0, len(value))
		for _, item := range value {
			payloadRaw, err := json.Marshal(item)
			if err != nil {
				return integrationtypes.IngestPayloadSet{}, fmt.Errorf("failed to encode scim ingest payload: %w", err)
			}

			envelopes = append(envelopes, integrationtypes.MappingEnvelope{
				Resource: resource,
				Action:   action,
				Payload:  payloadRaw,
			})
		}

		return integrationtypes.IngestPayloadSet{Schema: schema, Envelopes: envelopes}, nil
	default:
		payloadRaw, err := json.Marshal(payload)
		if err != nil {
			return integrationtypes.IngestPayloadSet{}, fmt.Errorf("failed to encode scim ingest payload: %w", err)
		}

		return integrationtypes.IngestPayloadSet{
			Schema: schema,
			Envelopes: []integrationtypes.MappingEnvelope{
				{
					Resource: resource,
					Action:   action,
					Payload:  payloadRaw,
				},
			},
		}, nil
	}
}

// cloneSCIMAttributes copies SCIM attributes so handler patching does not mutate shared maps
func cloneSCIMAttributes(attributes scimresource.ResourceAttributes) scimresource.ResourceAttributes {
	cloned := make(scimresource.ResourceAttributes, len(attributes))
	for key, value := range attributes {
		cloned[key] = cloneSCIMValue(value)
	}

	return cloned
}

// cloneSCIMValue deep-copies SCIM JSON-ish values used in attributes and patch payloads
func cloneSCIMValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, child := range typed {
			cloned[key] = cloneSCIMValue(child)
		}

		return cloned
	case scimresource.ResourceAttributes:
		return cloneSCIMAttributes(typed)
	case []any:
		cloned := make([]any, len(typed))
		for i, child := range typed {
			cloned[i] = cloneSCIMValue(child)
		}

		return cloned
	case []map[string]any:
		cloned := make([]any, 0, len(typed))
		for _, child := range typed {
			cloned = append(cloned, cloneSCIMValue(child))
		}

		return cloned
	default:
		return value
	}
}

// directoryAccountExternalID derives a stable account key from SCIM payload attributes
func directoryAccountExternalID(attributes scimresource.ResourceAttributes) string {
	if externalID, _ := attributes["externalId"].(string); strings.TrimSpace(externalID) != "" {
		return strings.TrimSpace(externalID)
	}

	if userName, _ := attributes["userName"].(string); strings.TrimSpace(userName) != "" {
		return strings.TrimSpace(userName)
	}

	if emails, ok := attributes["emails"].([]any); ok {
		for _, item := range emails {
			emailMap, ok := item.(map[string]any)
			if !ok {
				continue
			}

			value, _ := emailMap["value"].(string)
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}

	return ""
}

// directoryGroupExternalID derives a stable group key from SCIM payload attributes
func directoryGroupExternalID(attributes scimresource.ResourceAttributes) string {
	if externalID, _ := attributes["externalId"].(string); strings.TrimSpace(externalID) != "" {
		return strings.TrimSpace(externalID)
	}

	if displayName, _ := attributes["displayName"].(string); strings.TrimSpace(displayName) != "" {
		return strings.TrimSpace(displayName)
	}

	return ""
}

// mergeSCIMMap deep-merges SCIM patch values into a target attribute map
func mergeSCIMMap(target map[string]any, patch map[string]any) {
	for key, value := range patch {
		switch typed := value.(type) {
		case map[string]any:
			child := ensureSCIMMap(target, key)
			mergeSCIMMap(child, typed)
		default:
			target[key] = cloneSCIMValue(value)
		}
	}
}

// ensureSCIMMap returns a mutable nested SCIM map at key, creating one when needed
func ensureSCIMMap(target map[string]any, key string) map[string]any {
	if current, ok := target[key].(map[string]any); ok {
		return current
	}

	child := map[string]any{}
	target[key] = child

	return child
}

// uniqueStrings deduplicates string slices while preserving the first-seen order
func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))

	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}

		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}
