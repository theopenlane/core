//go:build cli

package trustcenternda

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

// command represents the base trust center NDA command when called without any subcommands
var command = &cobra.Command{
	Use:   "trust-center-nda",
	Short: "the subcommands for working with trust center ANDAs",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON
	if cmd.OutputFormat == cmd.JSONOutput {
		return jsonOutput(e)
	}

	// check the type of the output and print accordingly
	switch v := e.(type) {
	case *graphclient.CreateTrustCenterNda:
		return jsonOutput(v)
	case *graphclient.UpdateTrustCenterNda:
		return jsonOutput(v)
	default:
		return jsonOutput(v)
	}
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		return err
	}

	return cmd.JSONPrint(s)
}
