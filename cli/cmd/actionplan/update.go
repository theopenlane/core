//go:build cli

package actionPlan

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing actionPlan",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "actionplan id to update")
	updateCmd.Flags().StringP("name", "n", "", "name of the action plan")
	updateCmd.Flags().StringP("title", "t", "", "short title describing the action plan")
	updateCmd.Flags().StringP("status", "s", "", "status of the action plan (e.g. DRAFT, PUBLISHED, ARCHIVED)")
	updateCmd.Flags().StringP("priority", "p", "", "priority of the action plan (LOW, MEDIUM, HIGH, CRITICAL)")
	updateCmd.Flags().StringP("description", "d", "", "detailed description of the action plan")
	updateCmd.Flags().StringP("add-finding-ids", "", "", "comma-separated list of finding IDs to add")
	updateCmd.Flags().StringP("add-vulnerability-ids", "", "", "comma-separated list of vulnerability IDs to add")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateActionPlanInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("actionplan id")
	}

	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	title := cmd.Config.String("title")
	if title != "" {
		input.Title = &title
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToDocumentStatus(status)
	}

	priority := cmd.Config.String("priority")
	if priority != "" {
		input.Priority = enums.ToPriority(priority)
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	addFindingIDs := cmd.Config.String("add-finding-ids")
	if addFindingIDs != "" {
		input.AddFindingIDs = cmd.ParseIDList(addFindingIDs)
	}

	addVulnerabilityIDs := cmd.Config.String("add-vulnerability-ids")
	if addVulnerabilityIDs != "" {
		input.AddVulnerabilityIDs = cmd.ParseIDList(addVulnerabilityIDs)
	}

	return id, input, nil
}

// update an existing actionPlan in the platform
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

	o, err := client.UpdateActionPlan(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
