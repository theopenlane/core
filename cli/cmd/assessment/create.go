//go:build cli

package assessment

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new assessment",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the assessment")
	createCmd.Flags().StringP("type", "y", "", "type of assessment (e.g. INTERNAL, EXTERNAL)")
	createCmd.Flags().StringP("template-id", "t", "", "template ID to use for the assessment")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateAssessmentInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	assessmentType := cmd.Config.String("type")
	if assessmentType != "" {
		input.AssessmentType = enums.ToAssessmentType(assessmentType)
	}

	templateID := cmd.Config.String("template-id")
	if templateID != "" {
		input.TemplateID = &templateID
	}

	return input, nil
}

// create a new assessment
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateAssessment(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
