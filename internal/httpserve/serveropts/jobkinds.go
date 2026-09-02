package serveropts

import (
	"maps"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/definitions/catalog"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// jobTopicRenames merges every owner's retired-topic mapping for the dispatch-time rename
// fallback; delete once the pre-namespace backlog has drained
func jobTopicRenames(so *ServerOptions) map[gala.TopicName]gala.TopicName {
	renames := entityops.LegacyTopicRenames()
	maps.Copy(renames, operations.LegacyTopicRenames())
	maps.Copy(renames, definitionTopicRenames(so))

	return renames
}

// jobOperationRenames maps the retired soft-delete operation string to its designated operation
func jobOperationRenames() map[string]string {
	return map[string]string{"SoftDeleteOne": entityops.OpSoftDelete}
}

// definitionTopicRenames maps every definition's retired integration.<id>.* operation and
// webhook topics to their kind-prefixed replacements
func definitionTopicRenames(so *ServerOptions) map[gala.TopicName]gala.TopicName {
	renames := map[gala.TopicName]gala.TopicName{}

	for _, builder := range catalog.Builders(so.Config.Settings.Integrations, so.Config.Settings.Auth.Token.Issuer, so.Config.Settings.Server.Dev) {
		def, err := builder()
		if err != nil {
			// runtime construction registers the same builder and surfaces the failure there
			continue
		}

		for _, operation := range def.Operations {
			renames[gala.TopicName("integration."+def.ID+"."+operation.Name)] = operation.Topic
		}

		for _, webhook := range def.Webhooks {
			for _, event := range webhook.Events {
				renames[gala.TopicName("integration."+def.ID+".webhook."+event.Name)] = event.Topic
			}
		}
	}

	return renames
}
