//go:build cli

package review

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing review",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "review id to update")
	updateCmd.Flags().StringP("title", "t", "", "title of the review")
	updateCmd.Flags().StringP("state", "s", "", "state of the review")
	updateCmd.Flags().StringP("category", "c", "", "category for the review")
	updateCmd.Flags().StringP("summary", "m", "", "summary text for the review")
	updateCmd.Flags().StringP("add-finding-ids", "", "", "comma-separated list of finding IDs to add")
	updateCmd.Flags().StringP("add-vulnerability-ids", "", "", "comma-separated list of vulnerability IDs to add")
	updateCmd.Flags().StringP("add-asset-ids", "", "", "comma-separated list of asset IDs to add")
	updateCmd.Flags().StringP("add-entity-ids", "", "", "comma-separated list of entity IDs to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateReviewInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("review id")
	}

	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	state := cmd.Config.String("state")
	if state != "" {
		input.State = &state
	}

	category := cmd.Config.String("category")
	if category != "" {
		input.Category = &category
	}

	summary := cmd.Config.String("summary")
	if summary != "" {
		input.Summary = &summary
	}

	addFindingIDs := cmd.Config.String("add-finding-ids")
	if addFindingIDs != "" {
		input.AddFindingIDs = cmd.ParseIDList(addFindingIDs)
	}

	addVulnerabilityIDs := cmd.Config.String("add-vulnerability-ids")
	if addVulnerabilityIDs != "" {
		input.AddVulnerabilityIDs = cmd.ParseIDList(addVulnerabilityIDs)
	}

	addAssetIDs := cmd.Config.String("add-asset-ids")
	if addAssetIDs != "" {
		input.AddAssetIDs = cmd.ParseIDList(addAssetIDs)
	}

	addEntityIDs := cmd.Config.String("add-entity-ids")
	if addEntityIDs != "" {
		input.AddEntityIDs = cmd.ParseIDList(addEntityIDs)
	}

	return id, input, nil
}

// update an existing review in the platform
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

	o, err := client.UpdateReview(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
