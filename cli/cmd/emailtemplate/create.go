//go:build cli

package emailtemplate

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new email template",
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
	createCmd.Flags().String("template-context", "", "runtime data context (e.g. compliance, marketing)")
	createCmd.Flags().String("subject-template", "", "subject template for email notifications")
	createCmd.Flags().String("body-template", "", "body template for the email")
	createCmd.Flags().String("text-template", "", "plain text fallback template")
	createCmd.Flags().String("preheader-template", "", "preheader/preview text template")
	createCmd.Flags().String("template-format", "", "template format for rendering (TEXT, MARKDOWN, HTML)")
	createCmd.Flags().String("locale", "", "locale for the template, e.g. en-US")
	createCmd.Flags().Bool("active", true, "whether the template is active")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateEmailTemplateInput, err error) {
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

	templateContext := cmd.Config.String("template-context")
	if templateContext == "" {
		return input, cmd.NewRequiredFieldMissingError("template context")
	}

	input.TemplateContext = enums.TemplateContext(templateContext)

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	subjectTemplate := cmd.Config.String("subject-template")
	if subjectTemplate != "" {
		input.SubjectTemplate = &subjectTemplate
	}

	bodyTemplate := cmd.Config.String("body-template")
	if bodyTemplate != "" {
		input.BodyTemplate = &bodyTemplate
	}

	textTemplate := cmd.Config.String("text-template")
	if textTemplate != "" {
		input.TextTemplate = &textTemplate
	}

	preheaderTemplate := cmd.Config.String("preheader-template")
	if preheaderTemplate != "" {
		input.PreheaderTemplate = &preheaderTemplate
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

	active := cmd.Config.Bool("active")
	input.Active = &active

	return input, nil
}

// create a new email template
func create(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateEmailTemplate(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
