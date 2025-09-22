//go:build cli

package trustcenterdoc

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing trust center document",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "trust center document id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("trust-center-id", "t", "", "trust center id for the document")
	updateCmd.Flags().StringP("title", "n", "", "title of the document")
	updateCmd.Flags().StringP("category", "c", "", "category of the document")
	updateCmd.Flags().StringP("visibility", "v", "", "visibility of the document (NOT_VISIBLE, PROTECTED, PUBLICLY_VISIBLE)")
	updateCmd.Flags().StringSliceP("tags", "g", []string{}, "tags associated with the document")
	updateCmd.Flags().StringSliceP("append-tags", "a", []string{}, "append tags to the document")
}

// updateValidation validates the required fields for the command
func updateValidation() (string, openlaneclient.UpdateTrustCenterDocInput, error) {
	input := openlaneclient.UpdateTrustCenterDocInput{}

	id := cmd.Config.String("id")
	if id == "" {
		return "", input, cmd.NewRequiredFieldMissingError("id")
	}

	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	category := cmd.Config.String("category")
	if category != "" {
		input.Category = &category
	}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = &trustCenterID
	}

	visibility := cmd.Config.String("visibility")
	if visibility != "" {
		visibilityEnum := enums.ToTrustCenterDocumentVisibility(visibility)
		if visibilityEnum != nil {
			input.Visibility = visibilityEnum
		}
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

// update an existing trust center document in the platform
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

	o, err := client.UpdateTrustCenterDoc(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
