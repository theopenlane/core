//go:build cli

package workflowInstance

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing workflowInstance",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "workflowinstance id to query")
	getCmd.Flags().BoolP("show-assignments", "a", false, "show workflow assignment details with targets")
}

// get an existing workflowInstance in the platform
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}
	// filter options
	id := cmd.Config.String("id")
	showAssignments := cmd.Config.Bool("show-assignments")

	// if an workflowinstance ID is provided, filter on that workflowInstance, otherwise get all
	if id != "" {
		o, err := client.GetWorkflowInstanceByID(ctx, id)
		cobra.CheckErr(err)

		if showAssignments {
			return consoleOutputWithAssignments(o)
		}

		return consoleOutput(o)
	}

	var order *graphclient.WorkflowInstanceOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.WorkflowInstanceOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.WorkflowInstanceOrderField(*cmd.OrderBy),
		}
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllWorkflowInstances(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.WorkflowInstanceOrder{order})
	cobra.CheckErr(err)

	if showAssignments {
		return consoleOutputWithAssignmentsAll(o)
	}

	return consoleOutput(o)
}
