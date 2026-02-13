package hooks

import (
	"strings"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/events"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/subscriber"
	"github.com/theopenlane/core/internal/ent/generated/user"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/slacktemplates"
)

// RegisterGalaSlackListeners registers Gala mutation listeners that emit Slack notifications.
func RegisterGalaSlackListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterDurableListeners(registry, gala.QueueClassGeneral,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: gala.Topic[eventqueue.MutationGalaPayload]{
				Name: gala.TopicName(entgen.TypeSubscriber),
			},
			Name:   "slack.subscriber",
			Handle: handleSubscriberMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: gala.Topic[eventqueue.MutationGalaPayload]{
				Name: gala.TopicName(entgen.TypeUser),
			},
			Name:   "slack.user",
			Handle: handleUserMutationGala,
		},
	)
}

// handleSubscriberMutationGala sends a Slack notification for subscriber create mutations.
func handleSubscriberMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if payload.Operation != ent.OpCreate.String() {
		return nil
	}

	return sendSlackNotificationWithEmail(
		ctx.Context,
		mutationEmailFromGala(payload, ctx.Envelope.Headers.Properties, subscriber.FieldEmail),
		galaSubscriberTemplateOverride(),
		slacktemplates.GalaSubscriberTemplateName,
	)
}

// handleUserMutationGala sends a Slack notification for user create mutations.
func handleUserMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if payload.Operation != ent.OpCreate.String() {
		return nil
	}

	return sendSlackNotificationWithEmail(
		ctx.Context,
		mutationEmailFromGala(payload, ctx.Envelope.Headers.Properties, user.FieldEmail),
		galaUserTemplateOverride(),
		slacktemplates.GalaUserTemplateName,
	)
}

// mutationEmailFromGala resolves an email field from proposed changes with header fallback.
func mutationEmailFromGala(payload eventqueue.MutationGalaPayload, properties map[string]string, fieldName string) string {
	fieldName = strings.TrimSpace(fieldName)
	if fieldName == "" {
		return ""
	}

	rawProposedEmail, found := payload.ProposedChanges[fieldName]
	if found {
		proposedEmail, ok := events.ValueAsString(rawProposedEmail)
		if !ok {
			return ""
		}

		return strings.TrimSpace(proposedEmail)
	}

	return strings.TrimSpace(properties[fieldName])
}
