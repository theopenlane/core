//go:build cli

package trustcentercontrol

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing trust center control",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	// command line flags for the update command
	updateCmd.Flags().StringP("id", "i", "", "trust center control id to update")
	updateCmd.Flags().StringP("control-id", "c", "", "control id to associate with the trust center")
	updateCmd.Flags().StringP("trust-center-id", "t", "", "trust center id to associate with the control")
	updateCmd.Flags().StringSliceP("tags", "", []string{}, "tags associated with the trust center control")
	updateCmd.Flags().StringSliceP("append-tags", "", []string{}, "append tags to the trust center control")
	updateCmd.Flags().Bool("clear-tags", false, "clear all tags from the trust center control")
	updateCmd.Flags().Bool("clear-trust-center", false, "clear the trust center association")
}

// updateValidation validates the required fields for the command
func updateValidation() (string, openlaneclient.UpdateTrustCenterControlInput, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", openlaneclient.UpdateTrustCenterControlInput{}, cmd.NewRequiredFieldMissingError("id")
	}

	input := openlaneclient.UpdateTrustCenterControlInput{}

	controlID := cmd.Config.String("control-id")
	if controlID != "" {
		input.ControlID = &controlID
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = &trustCenterID
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	appendTags := cmd.Config.Strings("append-tags")
	if len(appendTags) > 0 {
		input.AppendTags = appendTags
	}

	clearTags := cmd.Config.Bool("clear-tags")
	if clearTags {
		clearTagsPtr := true
		input.ClearTags = &clearTagsPtr
	}

	clearTrustCenter := cmd.Config.Bool("clear-trust-center")
	if clearTrustCenter {
		clearTrustCenterPtr := true
		input.ClearTrustCenter = &clearTrustCenterPtr
	}

	return id, input, nil
}

// update an existing trust center control in the platform
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

	o, err := client.UpdateTrustCenterControl(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
