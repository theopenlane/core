//go:build cli

package standard

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing standard",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "standard id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the standard")
	updateCmd.Flags().StringP("description", "d", "", "description of the standard")
	updateCmd.Flags().StringP("version", "v", "", "version of the standard")
	updateCmd.Flags().StringSliceP("domains", "m", []string{}, "domain included in the standard")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "tags of the standard")
	updateCmd.Flags().StringP("category", "c", "", "category of the standard")
	updateCmd.Flags().StringP("status", "s", "", "status of the standard")
	updateCmd.Flags().StringP("link", "l", "", "link to the governing body of the standard")
	updateCmd.Flags().StringP("framework", "f", "", "framework name for the standard")
	updateCmd.Flags().StringP("governing-body", "g", "", "governing body of the standard")
	updateCmd.Flags().StringP("governing-body-log-url", "u", "", "governing body URL for the logo")
	updateCmd.Flags().StringP("bump-revision", "b", "", "bump major, minor, or patch version of the standard")
	updateCmd.Flags().StringP("revision", "r", "", "revision of the standard, used over bump-revision")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateStandardInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("standard id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
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

	revision := cmd.Config.String("revision")
	if revision != "" {
		input.Revision = &revision
	} else {
		bumpRevision := cmd.Config.String("bump-revision")
		if bumpRevision != "" {
			input.RevisionBump = models.ToVersionBump(bumpRevision)
		}
	}

	return id, input, nil
}

// update an existing standard in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateStandard(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
