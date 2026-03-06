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
			Name:       "slack.onboarding",
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

// handleDemoRequestMutationGala sends a Slack notification when an onboarding is created.
func handleDemoRequestMutationGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	if !SlackNotificationsEnabled() {
		return nil
	}

	companyName := eventqueue.MutationStringValuePreferPayload(payload, ctx.Envelope.Headers.Properties, onboarding.FieldCompanyName)
	domains := eventqueue.MutationStringSliceValue(payload, onboarding.FieldDomains)

	companyDetails, _ := mutationMapValue(payload, onboarding.FieldCompanyDetails)
	userDetails, _ := mutationMapValue(payload, onboarding.FieldUserDetails)
	compliance, _ := mutationMapValue(payload, onboarding.FieldCompliance)

	demoRequested, _ := eventqueue.MutationValue(payload, onboarding.FieldDemoRequested)

	var email string

	caller, ok := auth.CallerFromContext(ctx.Context)
	if ok && caller != nil {
		email = caller.SubjectEmail
	}

	tmpl, err := loadSlackTemplate(ctx.Context, "", slacktemplates.DemoRequestName)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, struct {
		CompanyName    string
		Email          string
		Domains        []string
		CompanyDetails map[string]any
		UserDetails    map[string]any
		Compliance     map[string]any
		DemoRequested  bool
	}{
		CompanyName:    companyName,
		Email:          email,
		Domains:        domains,
		CompanyDetails: companyDetails,
		UserDetails:    userDetails,
		Compliance:     compliance,
		DemoRequested:  demoRequested == true,
	}); err != nil {
		return err
	}

	return SendSlackNotification(ctx.Context, buf.String())
}

func mutationMapValue(payload eventqueue.MutationGalaPayload, field string) (map[string]any, bool) {
	raw, ok := eventqueue.MutationValue(payload, field)
	if !ok || raw == nil {
		return nil, false
	}

	m, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}

	return m, true
}
