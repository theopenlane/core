//go:build cli

package actionPlan

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new actionPlan",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the action plan")
	createCmd.Flags().StringP("title", "t", "", "short title describing the action plan")
	createCmd.Flags().StringP("status", "s", "", "status of the action plan (e.g. DRAFT, PUBLISHED, ARCHIVED)")
	createCmd.Flags().StringP("priority", "p", "", "priority of the action plan (LOW, MEDIUM, HIGH, CRITICAL)")
	createCmd.Flags().StringP("description", "d", "", "detailed description of the action plan")
	createCmd.Flags().StringP("finding-ids", "f", "", "comma-separated list of finding IDs to associate with the action plan")
	createCmd.Flags().StringP("vulnerability-ids", "v", "", "comma-separated list of vulnerability IDs to associate with the action plan")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateActionPlanInput, err error) {
	input.Name = cmd.Config.String("name")
	if input.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	input.Title = cmd.Config.String("title")
	if input.Title == "" {
		return input, cmd.NewRequiredFieldMissingError("title")
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

	findingIDs := cmd.Config.String("finding-ids")
	if findingIDs != "" {
		input.FindingIDs = cmd.ParseIDList(findingIDs)
	}

	vulnerabilityIDs := cmd.Config.String("vulnerability-ids")
	if vulnerabilityIDs != "" {
		input.VulnerabilityIDs = cmd.ParseIDList(vulnerabilityIDs)
	}

	return input, nil
}

// create a new actionPlan
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

	o, err := client.CreateActionPlan(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
