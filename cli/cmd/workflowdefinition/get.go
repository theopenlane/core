//go:build cli

package workflowdefinition

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"list", "ls"},
	Short:   "get workflow definitions",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("id", "i", "", "workflow definition id to query")
	getCmd.Flags().StringP("name", "n", "", "filter by workflow definition name (contains, case-insensitive)")
	getCmd.Flags().String("schema-type", "", "filter by workflow schema type (e.g. Control)")
	getCmd.Flags().String("workflow-kind", "", "filter by workflow kind (APPROVAL, LIFECYCLE, NOTIFICATION)")
	getCmd.Flags().String("display-id", "", "filter by workflow definition display id")
}

// get workflow definitions in the platform
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
	name := cmd.Config.String("name")
	schemaType := cmd.Config.String("schema-type")
	workflowKind := cmd.Config.String("workflow-kind")
	displayID := cmd.Config.String("display-id")

	// if an ID is provided, filter on that workflow definition
	if id != "" {
		o, err := client.GetWorkflowDefinitionByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	var order *graphclient.WorkflowDefinitionOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.WorkflowDefinitionOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.WorkflowDefinitionOrderField(*cmd.OrderBy),
		}
	}

	where := graphclient.WorkflowDefinitionWhereInput{}
	hasWhere := false

	if name != "" {
		where.NameContainsFold = &name
		hasWhere = true
	}

	if schemaType != "" {
		where.SchemaTypeEqualFold = &schemaType
		hasWhere = true
	}

	if displayID != "" {
		where.DisplayID = &displayID
		hasWhere = true
	}

	if workflowKind != "" {
		parsed := enums.ToWorkflowKind(workflowKind)
		if parsed == nil {
			return fmt.Errorf("invalid workflow kind %q (expected: APPROVAL, LIFECYCLE, NOTIFICATION)", workflowKind)
		}
		where.WorkflowKind = parsed
		hasWhere = true
	}

	if hasWhere {
		o, err := client.GetWorkflowDefinitions(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &where, []*graphclient.WorkflowDefinitionOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllWorkflowDefinitions(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.WorkflowDefinitionOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
