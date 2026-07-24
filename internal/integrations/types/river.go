package types //nolint:revive

import (
	"github.com/theopenlane/core/pkg/gala"
)

// GetTagsForOperationContext safely translates an operation context to river-safe tags
func GetTagsForOperationContext(oc gala.OperationContext) []string {
	src := IntegrationSourceFrom(oc)

	tags := []string{}
	if src.DefinitionID != "" {
		tags = append(tags, gala.SanitizeTag(src.DefinitionID))
	}

	if oc.Operation != "" {
		tags = append(tags, gala.SanitizeTag(oc.Operation))
	}

	if src.RunType != "" {
		tags = append(tags, gala.SanitizeTag(src.RunType.String()))
	}

	if src.Webhook != "" {
		tags = append(tags, gala.SanitizeTag(src.Webhook))
	}

	return tags
}

// GetPropertiesForOperationContext returns header properties for an operation context,
// merging source provenance fields used for active-job metadata matching
func GetPropertiesForOperationContext(oc gala.OperationContext) map[string]string {
	src := IntegrationSourceFrom(oc)
	props := oc.Properties()

	if src.DefinitionID != "" {
		props["definitionId"] = src.DefinitionID
	}

	if src.RunType != "" {
		props["runType"] = src.RunType.String()
	}

	return props
}
