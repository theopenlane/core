package serveropts

import (
	"context"
	"maps"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/integrations/operations"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

// jobTopicRenames merges every owner's retired-topic mapping
func jobTopicRenames() map[gala.TopicName]gala.TopicName {
	renames := entityops.LegacyTopicRenames()
	maps.Copy(renames, operations.LegacyTopicRenames())

	return renames
}

// jobTransition applies retired-topic renames; kind resolution belongs to the runtime
func jobTransition(renames map[gala.TopicName]gala.TopicName) func(string, gala.Envelope) (gala.Envelope, bool) {
	return func(_ string, envelope gala.Envelope) (gala.Envelope, bool) {
		if renamed, ok := renames[envelope.Topic]; ok {
			envelope.Topic = renamed
		}

		return envelope, true
	}
}

// jobTransitionRequest is the payload for a job transition migration run
type jobTransitionRequest struct{}

// jobTransitionTopic is the topic job transition migration runs are submitted on
var jobTransitionTopic = gala.NamespacedTopic[jobTransitionRequest](gala.SystemTopics, "job_transition")

// jobTransitionUniqueKey collapses concurrent pod startups to one live migration run
const jobTransitionUniqueKey = "job-transition-migration"

// submitJobTransitionMigration registers the migration listener and submits the run-once
// job; insert-time uniqueness elects exactly one runner across concurrently starting pods
func submitJobTransitionMigration(ctx context.Context, galaApp *gala.Gala) {
	if _, err := gala.Register(galaApp, gala.Definition[jobTransitionRequest]{
		Topic: jobTransitionTopic,
		Handle: func(handlerCtx gala.HandlerContext, _ jobTransitionRequest) error {
			migrated, err := galaApp.MigrateJobs(handlerCtx.Context, jobTransition(jobTopicRenames()))
			if err != nil {
				logx.FromContext(handlerCtx.Context).Error().Err(err).Int("migrated", migrated).Msg("job transition migration incomplete")

				return err
			}

			if migrated > 0 {
				logx.FromContext(handlerCtx.Context).Info().Int("migrated", migrated).Msg("queued jobs migrated to designated topics and kinds")
			}

			return nil
		},
	}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to register job transition migration listener")

		return
	}

	if _, err := galaApp.EmitWithHeaders(ctx, jobTransitionTopic.Name, jobTransitionRequest{}, gala.Headers{
		UniqueKey: jobTransitionUniqueKey,
	}); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to submit job transition migration")
	}
}
