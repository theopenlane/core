//go:build cli

package notificationtemplate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new notification template",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("key", "k", "", "stable identifier for the template")
	createCmd.Flags().StringP("name", "n", "", "display name for the template")
	createCmd.Flags().StringP("description", "d", "", "description of the template")
	createCmd.Flags().String("channel", "", "channel this template is intended for (e.g. EMAIL, SLACK)")
	createCmd.Flags().String("topic-pattern", "", "topic name or wildcard pattern this template targets")
	createCmd.Flags().String("title-template", "", "title template for external channel messages")
	createCmd.Flags().String("subject-template", "", "subject template for email notifications")
	createCmd.Flags().String("body-template", "", "body template for the notification")
	createCmd.Flags().String("template-format", "", "template format for rendering (TEXT, MARKDOWN, HTML)")
	createCmd.Flags().String("locale", "", "locale for the template, e.g. en-US")
	createCmd.Flags().String("email-template-id", "", "email template used for branded email delivery")
	createCmd.Flags().Bool("active", true, "whether the template is active")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateNotificationTemplateInput, err error) {
	key := cmd.Config.String("key")
	if key == "" {
		return input, cmd.NewRequiredFieldMissingError("key")
	}

	input.Key = key

	name := cmd.Config.String("name")
	if name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.Name = name

	channel := cmd.Config.String("channel")
	if channel == "" {
		return input, cmd.NewRequiredFieldMissingError("channel")
	}

	input.Channel = enums.Channel(channel)

	topicPattern := cmd.Config.String("topic-pattern")
	if topicPattern == "" {
		return input, cmd.NewRequiredFieldMissingError("topic pattern")
	}

	input.TopicPattern = topicPattern

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	titleTemplate := cmd.Config.String("title-template")
	if titleTemplate != "" {
		input.TitleTemplate = &titleTemplate
	}

	subjectTemplate := cmd.Config.String("subject-template")
	if subjectTemplate != "" {
		input.SubjectTemplate = &subjectTemplate
	}

	bodyTemplate := cmd.Config.String("body-template")
	if bodyTemplate != "" {
		input.BodyTemplate = &bodyTemplate
	}

	templateFormat := cmd.Config.String("template-format")
	if templateFormat != "" {
		f := enums.NotificationTemplateFormat(templateFormat)
		input.Format = &f
	}

	locale := cmd.Config.String("locale")
	if locale != "" {
		input.Locale = &locale
	}

	emailTemplateID := cmd.Config.String("email-template-id")
	if emailTemplateID != "" {
		input.EmailTemplateID = &emailTemplateID
	}

	active := cmd.Config.Bool("active")
	input.Active = &active

	return input, nil
}

// create a new notification template
func create(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateNotificationTemplate(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
