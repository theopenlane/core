//go:build cli

package templates

import (
	"context"
	"encoding/json"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new template",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the template")
	createCmd.Flags().StringP("description", "d", "", "description of the template")
	createCmd.Flags().StringP("json-config", "j", "", "json payload for the template")
	createCmd.Flags().StringP("ui-schema", "u", "", "ui schema for the template")
	createCmd.Flags().StringP("type", "t", "DOCUMENT", "type of the template")
	createCmd.Flags().StringP("kind", "k", "QUESTIONNAIRE", "kind of the template")
	createCmd.Flags().StringP("trust-center-id", "c", "", "trust center id to associate with the template")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateTemplateInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("template name")
	}

	jsonConfig := cmd.Config.String("json-config")
	if jsonConfig == "" {
		return input, cmd.NewRequiredFieldMissingError("json config")
	}

	data, err := os.ReadFile(jsonConfig)
	cobra.CheckErr(err)

	// Unmarshal the JSON data into the map
	var jsonData map[string]any
	err = json.Unmarshal(data, &jsonData)
	cobra.CheckErr(err)

	input.Jsonconfig = jsonData

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	uiSchema := cmd.Config.String("ui-schema")
	if uiSchema != "" {
		data, err = os.ReadFile(uiSchema)
		cobra.CheckErr(err)

		// Unmarshal the JSON data into the map
		var jsonData map[string]any
		err = json.Unmarshal(data, &jsonData)
		cobra.CheckErr(err)

		input.Uischema = jsonData
	}

	templateType := cmd.Config.String("type")
	if templateType != "" {
		input.TemplateType = enums.ToDocumentType(templateType)
	}

	templateKind := cmd.Config.String("kind")
	if templateKind != "" {
		input.Kind = enums.ToTemplateKind(templateKind)
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = &trustCenterID
	}

	return input, nil
}

// create a new template
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

	o, err := client.CreateTemplate(ctx, input, []*graphql.Upload{})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
