package hooks

import (
	"bytes"
	"html/template"
	"os"

	"entgo.io/ent"
	"github.com/rs/zerolog"

	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/core/pkg/slacktemplates"
	"github.com/theopenlane/utils/slack"
)

type SlackConfig struct {
	WebhookURL               string
	NewSubscriberMessageFile string
	NewUserMessageFile       string
}

var slackCfg SlackConfig

func SetSlackConfig(cfg SlackConfig) {
	slackCfg = cfg
}

func handleSubscriberMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
	if payload.Operation != ent.OpCreate.String() {
		return nil
	}

	return handleSubscriberCreate(ctx, payload)
}

func handleUserMutation(ctx *soiree.EventContext, payload *MutationPayload) error {
	if payload.Operation != ent.OpCreate.String() {
		return nil
	}

	return handleUserCreate(ctx, payload)
}

func handleSubscriberCreate(ctx *soiree.EventContext, _ *MutationPayload) error {
	if slackCfg.WebhookURL == "" {
		return nil
	}

	email, _ := ctx.Properties().String("email")

	var (
		t   *template.Template
		err error
		msg string
	)

	if slackCfg.NewSubscriberMessageFile != "" {
		b, err := os.ReadFile(slackCfg.NewSubscriberMessageFile)
		if err != nil {
			zerolog.Ctx(ctx.Context()).Debug().Msg("failed to read slack template")

			return err
		}

		t, err = template.New("slack").Parse(string(b))
		if err != nil {
			zerolog.Ctx(ctx.Context()).Debug().Msg("failed to parse slack template")

			return err
		}
	} else {
		t, err = template.ParseFS(slacktemplates.Templates, slacktemplates.SubscriberTemplateName)
		if err != nil {
			zerolog.Ctx(ctx.Context()).Debug().Msg("failed to parse embedded slack template")

			return err
		}
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, struct{ Email string }{Email: email}); err != nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("failed to execute slack template")

		return err
	}

	msg = buf.String()

	client := slack.New(slackCfg.WebhookURL)

	payload := &slack.Payload{Text: msg}

	return client.Post(ctx.Context(), payload)
}

func handleUserCreate(ctx *soiree.EventContext, _ *MutationPayload) error {
	if slackCfg.WebhookURL == "" {
		return nil
	}

	email, _ := ctx.Properties().String("email")

	var (
		t   *template.Template
		err error
		msg string
	)

	if slackCfg.NewUserMessageFile != "" {
		b, err := os.ReadFile(slackCfg.NewUserMessageFile)
		if err != nil {
			zerolog.Ctx(ctx.Context()).Debug().Msg("failed to read slack template")

			return err
		}

		t, err = template.New("slack").Parse(string(b))
		if err != nil {
			zerolog.Ctx(ctx.Context()).Debug().Msg("failed to parse slack template")

			return err
		}
	} else {
		t, err = template.ParseFS(slacktemplates.Templates, slacktemplates.UserTemplateName)
		if err != nil {
			zerolog.Ctx(ctx.Context()).Debug().Msg("failed to parse embedded slack template")

			return err
		}
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, struct{ Email string }{Email: email}); err != nil {
		zerolog.Ctx(ctx.Context()).Debug().Msg("failed to execute slack template")

		return err
	}

	msg = buf.String()

	client := slack.New(slackCfg.WebhookURL)

	payload := &slack.Payload{Text: msg}

	return client.Post(ctx.Context(), payload)
}
