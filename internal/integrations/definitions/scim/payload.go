package scim

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elimity-com/scim"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

// DeleteAction is the action string used to indicate a SCIM delete operation
const DeleteAction = "delete"

// BuildDirectoryAccountPayloadSet constructs an ingest payload set for a directory account
func BuildDirectoryAccountPayloadSet(attributes scim.ResourceAttributes, action string) (integrationtypes.IngestPayloadSet, error) {
	payload := CloneSCIMAttributes(attributes)

	externalID := DirectoryAccountExternalID(payload)
	if externalID == "" {
		return integrationtypes.IngestPayloadSet{}, ErrInvalidAttributes
	}

	payload["externalId"] = externalID

	return BuildDirectoryPayloadSet(
		integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
		payload, externalID, action,
	)
}

// BuildDirectoryGroupPayloadSets constructs ingest payload sets for a directory group and its memberships
func BuildDirectoryGroupPayloadSets(attributes scim.ResourceAttributes, action string) ([]integrationtypes.IngestPayloadSet, error) {
	payload := CloneSCIMAttributes(attributes)

	externalID := DirectoryGroupExternalID(payload)
	if externalID == "" {
		return nil, ErrInvalidAttributes
	}

	payload["externalId"] = externalID

	groupPayloadSet, err := BuildDirectoryPayloadSet(
		integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
		payload, externalID, action,
	)
	if err != nil {
		return nil, err
	}

	if action == DeleteAction {
		return []integrationtypes.IngestPayloadSet{groupPayloadSet}, nil
	}

	memberIDs := ExtractMemberIDsFromValue(payload["members"])
	if len(memberIDs) == 0 {
		return []integrationtypes.IngestPayloadSet{groupPayloadSet}, nil
	}

	membershipPayloads := make([]any, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		membershipPayloads = append(membershipPayloads, map[string]any{
			"group":  map[string]any{"externalId": externalID},
			"member": map[string]any{"value": memberID},
		})
	}

	membershipPayloadSet, err := BuildDirectoryPayloadSet(
		integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
		membershipPayloads, externalID, action,
	)
	if err != nil {
		return nil, err
	}

	return []integrationtypes.IngestPayloadSet{groupPayloadSet, membershipPayloadSet}, nil
}

// BuildDirectoryPayloadSet constructs an ingest payload set from a schema name, payload, resource ID, and action
func BuildDirectoryPayloadSet(schema string, payload any, resource string, action string) (integrationtypes.IngestPayloadSet, error) {
	items, ok := payload.([]any)
	if !ok {
		items = []any{payload}
	}

	envelopes := make([]integrationtypes.MappingEnvelope, 0, len(items))

	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			return integrationtypes.IngestPayloadSet{}, fmt.Errorf("failed to encode scim ingest payload: %w", err)
		}

		envelopes = append(envelopes, integrationtypes.MappingEnvelope{
			Resource: resource, Action: action, Payload: raw,
		})
	}

	return integrationtypes.IngestPayloadSet{Schema: schema, Envelopes: envelopes}, nil
}

// CloneSCIMAttributes performs a deep copy of SCIM resource attributes
func CloneSCIMAttributes(attributes scim.ResourceAttributes) scim.ResourceAttributes {
	cloned := make(scim.ResourceAttributes, len(attributes))
	for key, value := range attributes {
		cloned[key] = CloneSCIMValue(value)
	}

	return cloned
}

// CloneSCIMValue performs a deep copy of a SCIM attribute value
func CloneSCIMValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, child := range typed {
			cloned[key] = CloneSCIMValue(child)
		}

		return cloned
	case scim.ResourceAttributes:
		return CloneSCIMAttributes(typed)
	case []any:
		cloned := make([]any, len(typed))
		for i, child := range typed {
			cloned[i] = CloneSCIMValue(child)
		}

		return cloned
	case []map[string]any:
		cloned := make([]any, 0, len(typed))
		for _, child := range typed {
			cloned = append(cloned, CloneSCIMValue(child))
		}

		return cloned
	default:
		return value
	}
}

// DirectoryAccountExternalID resolves the external ID for a directory account
// by checking externalId, userName, and emails in fallback order
func DirectoryAccountExternalID(attributes scim.ResourceAttributes) string {
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

// DirectoryGroupExternalID resolves the external ID for a directory group
// by checking externalId and displayName in fallback order
func DirectoryGroupExternalID(attributes scim.ResourceAttributes) string {
	if externalID, _ := attributes["externalId"].(string); strings.TrimSpace(externalID) != "" {
		return strings.TrimSpace(externalID)
	}

	if displayName, _ := attributes["displayName"].(string); strings.TrimSpace(displayName) != "" {
		return strings.TrimSpace(displayName)
	}

	return ""
}

// MergeSCIMMap recursively merges patch values into a target map
func MergeSCIMMap(target map[string]any, patch map[string]any) {
	for key, value := range patch {
		switch typed := value.(type) {
		case map[string]any:
			child := EnsureSCIMMap(target, key)
			MergeSCIMMap(child, typed)
		default:
			target[key] = CloneSCIMValue(value)
		}
	}
}

// EnsureSCIMMap returns the nested map at key, creating it if absent
func EnsureSCIMMap(target map[string]any, key string) map[string]any {
	if current, ok := target[key].(map[string]any); ok {
		return current
	}

	child := map[string]any{}
	target[key] = child

	return child
}

