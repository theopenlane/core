//go:build cli

package workflowAssignment

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing workflowAssignment",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "workflowassignment id to query")
	getCmd.Flags().BoolP("show-targets", "t", false, "show assignment target details")
}

// get an existing workflowAssignment in the platform
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
	showTargets := cmd.Config.Bool("show-targets")

	// if an workflowassignment ID is provided, filter on that workflowAssignment, otherwise get all
	if id != "" {
		o, err := client.GetWorkflowAssignmentByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.WorkflowAssignmentOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.WorkflowAssignmentOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.WorkflowAssignmentOrderField(*cmd.OrderBy),
		}
	}

	// use GetWorkflowAssignments when showing targets (includes edge data)
	if showTargets {
		o, err := client.GetWorkflowAssignments(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, nil, []*graphclient.WorkflowAssignmentOrder{order})
		cobra.CheckErr(err)

		return consoleOutputWithTargets(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllWorkflowAssignments(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.WorkflowAssignmentOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
