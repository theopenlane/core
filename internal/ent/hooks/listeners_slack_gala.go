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
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: gala.Topic[eventqueue.MutationGalaPayload]{
				Name: gala.TopicName(entgen.TypeSubscriber),
			},
			Name:       "slack.subscriber",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleSubscriberMutationGala,
		},
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: gala.Topic[eventqueue.MutationGalaPayload]{
				Name: gala.TopicName(entgen.TypeUser),
			},
			Name:       "slack.user",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleUserMutationGala,
		},
	)
}

// handleSubscriberMutationGala sends a Slack notification for subscriber create mutations.
func handleSubscriberMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	return sendSlackNotificationWithEmail(
		ctx.Context,
		mutationEmailFromGala(payload, ctx.Envelope.Headers.Properties, subscriber.FieldEmail),
		subscriberTemplateOverride(),
		slacktemplates.SubscriberTemplateName,
	)
}

// handleUserMutationGala sends a Slack notification for user create mutations.
func handleUserMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	return sendSlackNotificationWithEmail(
		ctx.Context,
		mutationEmailFromGala(payload, ctx.Envelope.Headers.Properties, user.FieldEmail),
		userTemplateOverride(),
		slacktemplates.UserTemplateName,
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
