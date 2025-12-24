//go:build cli

package subcontrol

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get an existing subcontrol",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)
	getCmd.Flags().StringP("id", "i", "", "subcontrol id to query")
	getCmd.Flags().StringP("ref-code", "r", "", "the unique reference code of the subcontrol")
	getCmd.Flags().StringP("control-id", "c", "", "control id to filter the query")
	getCmd.Flags().StringP("control", "o", "", "control reference code to filter the query")
}

// get an existing subcontrol in the platform
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
	refCode := cmd.Config.String("ref-code")
	controlRefCode := cmd.Config.String("control")
	controlID := cmd.Config.String("control")

	// if an subcontrol ID is provided, filter on that subcontrol
	if id != "" {
		o, err := client.GetSubcontrolByID(ctx, id)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	order := &graphclient.SubcontrolOrder{}
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.SubcontrolOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.SubcontrolOrderField(*cmd.OrderBy),
		}
	}

	// if a ref code is provided, filter on that control
	if refCode != "" {
		o, err := client.GetSubcontrols(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &graphclient.SubcontrolWhereInput{
			RefCode: &refCode,
		}, []*graphclient.SubcontrolOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// if a control ID is provided, filter on that control
	if controlID != "" {
		o, err := client.GetSubcontrols(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &graphclient.SubcontrolWhereInput{
			ControlID: &controlID,
		}, []*graphclient.SubcontrolOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// if a control ref code is provided, filter on that control
	if controlRefCode != "" {
		where := &graphclient.ControlWhereInput{
			RefCode: &controlRefCode,
		}

		o, err := client.GetSubcontrols(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &graphclient.SubcontrolWhereInput{
			HasControlWith: []*graphclient.ControlWhereInput{where},
		}, []*graphclient.SubcontrolOrder{order})
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	// get all will be filtered for the authorized organization(s)
	o, err := client.GetAllSubcontrols(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.SubcontrolOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
