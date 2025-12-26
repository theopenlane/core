//go:build cli

package programmembers

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get existing members of a program",
	Run: func(cmd *cobra.Command, args []string) {
		err := get(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(getCmd)

	getCmd.Flags().StringP("program-id", "p", "", "program id to query")
}

// get existing members of a program
func get(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	var order *graphclient.ProgramMembershipOrder
	if cmd.OrderBy != nil && cmd.OrderDirection != nil {
		order = &graphclient.ProgramMembershipOrder{
			Direction: graphclient.OrderDirection(strings.ToUpper(*cmd.OrderDirection)),
			Field:     graphclient.ProgramMembershipOrderField(*cmd.OrderBy),
		}
	}
	// filter options
	id := cmd.Config.String("program-id")
	if id == "" {
		o, err := client.GetAllProgramMemberships(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, []*graphclient.ProgramMembershipOrder{order})
		cobra.CheckErr(err)
		return consoleOutput(o)
	}

	where := graphclient.ProgramMembershipWhereInput{
		ProgramID: &id,
	}

	o, err := client.GetProgramMemberships(ctx, cmd.First, cmd.Last, cmd.After, cmd.Before, &where, []*graphclient.ProgramMembershipOrder{order})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
