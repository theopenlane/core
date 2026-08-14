package serveropts

import (
	"maps"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
)

// jobTopicRenames merges every owner's retired-topic mapping for the dispatch-time rename
// fallback; delete once the pre-namespace backlog has drained
func jobTopicRenames() map[gala.TopicName]gala.TopicName {
	renames := entityops.LegacyTopicRenames()
	maps.Copy(renames, operations.LegacyTopicRenames())

	return renames
}
