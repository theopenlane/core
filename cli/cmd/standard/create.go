//go:build cli

package standard

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new standard",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the standard")
	createCmd.Flags().StringP("description", "d", "", "description of the standard")
	createCmd.Flags().StringP("version", "v", "", "version of the standard")
	createCmd.Flags().StringSliceP("domains", "m", []string{}, "domain included in the standard")
	createCmd.Flags().StringSliceP("tags", "t", []string{}, "tags of the standard")
	createCmd.Flags().StringP("category", "c", "", "category of the standard")
	createCmd.Flags().StringP("status", "s", "", "status of the standard")
	createCmd.Flags().StringP("link", "l", "", "link to the governing body of the standard")
	createCmd.Flags().StringP("framework", "f", "", "framework name for the standard")
	createCmd.Flags().StringP("governing-body", "g", "", "governing body of the standard")
	createCmd.Flags().StringP("governing-body-log-url", "u", "", "governing body URL for the logo")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateStandardInput, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	version := cmd.Config.String("version")
	if version != "" {
		input.Version = &version
	}

	domains := cmd.Config.Strings("domains")
	if len(domains) > 0 {
		input.Domains = domains
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	category := cmd.Config.String("category")
	if category != "" {
		input.StandardType = &category
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToStandardStatus(status)
	}

	link := cmd.Config.String("link")
	if link != "" {
		input.Link = &link
	}

	framework := cmd.Config.String("framework")
	if framework != "" {
		input.Framework = &framework
	}

	governingBody := cmd.Config.String("governing-body")
	if governingBody != "" {
		input.GoverningBody = &governingBody
	}

	governingBodyLogoURL := cmd.Config.String("governing-body-log-url")
	if governingBodyLogoURL != "" {
		input.GoverningBodyLogoURL = &governingBodyLogoURL
	}

	return input, nil
}

// create a new standard
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

	o, err := client.CreateStandard(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
