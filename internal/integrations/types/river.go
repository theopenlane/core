package types //nolint:revive

import (
	"github.com/theopenlane/core/pkg/gala"
)

// GetTagsForExecutionMetadata safely translate execution metadata to river-safe tags
func GetTagsForExecutionMetadata(meta ExecutionMetadata) []string {
	tags := []string{}
	if meta.DefinitionID != "" {
		tags = append(tags, gala.SanitizeTag(meta.DefinitionID))
	}

	if meta.Operation != "" {
		tags = append(tags, gala.SanitizeTag(meta.Operation))
	}

	if meta.RunType != "" {
		tags = append(tags, gala.SanitizeTag(meta.RunType.String()))
	}

	if meta.Webhook != "" {
		tags = append(tags, gala.SanitizeTag(meta.Webhook))
	}

	return tags
}
