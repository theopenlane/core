//go:build cli

package review

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new review",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("title", "t", "", "title of the review")
	createCmd.Flags().StringP("state", "s", "", "state of the review")
	createCmd.Flags().StringP("category", "c", "", "category for the review")
	createCmd.Flags().StringP("summary", "m", "", "summary text for the review")
	createCmd.Flags().StringP("finding-ids", "f", "", "comma-separated list of finding IDs to associate with the review")
	createCmd.Flags().StringP("vulnerability-ids", "v", "", "comma-separated list of vulnerability IDs to associate with the review")
	createCmd.Flags().StringP("asset-ids", "", "", "comma-separated list of asset IDs to associate with the review")
	createCmd.Flags().StringP("entity-ids", "", "", "comma-separated list of entity IDs to associate with the review")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateReviewInput, err error) {
	input.Title = cmd.Config.String("title")
	if input.Title == "" {
		return input, cmd.NewRequiredFieldMissingError("title")
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

	findingIDs := cmd.Config.String("finding-ids")
	if findingIDs != "" {
		input.FindingIDs = cmd.ParseIDList(findingIDs)
	}

	vulnerabilityIDs := cmd.Config.String("vulnerability-ids")
	if vulnerabilityIDs != "" {
		input.VulnerabilityIDs = cmd.ParseIDList(vulnerabilityIDs)
	}

	assetIDs := cmd.Config.String("asset-ids")
	if assetIDs != "" {
		input.AssetIDs = cmd.ParseIDList(assetIDs)
	}

	entityIDs := cmd.Config.String("entity-ids")
	if entityIDs != "" {
		input.EntityIDs = cmd.ParseIDList(entityIDs)
	}

	return input, nil
}

// create a new review
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

	o, err := client.CreateReview(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
