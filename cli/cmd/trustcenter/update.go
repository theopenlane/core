//go:build cli

package trustcenter

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing trustcenter",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "trustcenter id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("custom-domain-id", "d", "", "custom domain id for the trustcenter")
	updateCmd.Flags().StringSliceP("tags", "t", []string{}, "tags associated with the trustcenter")
	updateCmd.Flags().StringSliceP("append-tags", "a", []string{}, "append tags to the trustcenter")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateTrustCenterInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("id")
	}

	customDomainID := cmd.Config.String("custom-domain-id")
	if customDomainID != "" {
		input.CustomDomainID = &customDomainID
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	appendTags := cmd.Config.Strings("append-tags")
	if len(appendTags) > 0 {
		input.AppendTags = appendTags
	}

	return id, input, nil
}

// update an existing trustcenter in the platform
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

	o, err := client.UpdateTrustCenter(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
