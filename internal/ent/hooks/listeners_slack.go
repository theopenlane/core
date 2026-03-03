package hooks

import (
	"bytes"

	"entgo.io/ent"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/onboarding"
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
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic: gala.Topic[eventqueue.MutationGalaPayload]{
				Name: gala.TopicName(entgen.TypeOnboarding),
			},
			Name:       "slack.demo_request",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleDemoRequestMutationGala,
		},
	)
}

// handleSubscriberMutationGala sends a Slack notification for subscriber create mutations.
func handleSubscriberMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	return sendSlackNotificationWithEmail(
		ctx.Context,
		eventqueue.MutationStringValuePreferPayload(payload, ctx.Envelope.Headers.Properties, subscriber.FieldEmail),
		subscriberTemplateOverride(),
		slacktemplates.SubscriberTemplateName,
	)
}

// handleUserMutationGala sends a Slack notification for user create mutations.
func handleUserMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	return sendSlackNotificationWithEmail(
		ctx.Context,
		eventqueue.MutationStringValuePreferPayload(payload, ctx.Envelope.Headers.Properties, user.FieldEmail),
		userTemplateOverride(),
		slacktemplates.UserTemplateName,
	)
}

// handleDemoRequestMutationGala sends a Slack notification when an onboarding is created with demo_requested set to true.
func handleDemoRequestMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !SlackNotificationsEnabled() {
		return nil
	}

	raw, ok := eventqueue.MutationValue(payload, onboarding.FieldDemoRequested)
	if !ok || raw != true {
		return nil
	}

	companyName := eventqueue.MutationStringValuePreferPayload(payload, ctx.Envelope.Headers.Properties, onboarding.FieldCompanyName)

	var email string

	if au, err := auth.GetAuthenticatedUserFromContext(ctx.Context); err == nil && au != nil {
		email = au.SubjectEmail
	}

	tmpl, err := loadSlackTemplate(ctx.Context, "", slacktemplates.DemoRequestName)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, struct {
		CompanyName string
		Email       string
	}{
		CompanyName: companyName,
		Email:       email,
	}); err != nil {
		return err
	}

	return SendSlackNotification(ctx.Context, buf.String())
}
