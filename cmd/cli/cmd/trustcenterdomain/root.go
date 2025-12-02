//go:build cli

package trustcenterdomain

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

// command represents the base trustcenterdomain command when called without any subcommands
var command = &cobra.Command{
	Use:     "trust-center-domain",
	Aliases: []string{"trustcenterdomain"},
	Short:   "the subcommands for working with trust center domains",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the trust center domains in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the trust center domains and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.CreateTrustCenterDomain:
		nodes := []*openlaneclient.CreateTrustCenterDomain_CreateTrustCenterDomain_CustomDomain{
			v.CreateTrustCenterDomain.GetCustomDomain(),
		}

		e = nodes
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.CustomDomain

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.CustomDomain
		err = json.Unmarshal(s, &in)
		cobra.CheckErr(err)

		list = append(list, in)
	}

	tableOutput(list)

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the output in a table format
func tableOutput(out []openlaneclient.CustomDomain) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "CNAME", "Trust Center ID", "Verification ID", "Created At")
	for _, i := range out {
		verificationID := ""
		if i.DNSVerificationID != nil {
			verificationID = *i.DNSVerificationID
		}

		// For trust center domains, we don't have direct access to trust center ID from CustomDomain
		// This would need to be retrieved from the relationship or passed through
		trustCenterID := "N/A"

		writer.AddRow(i.ID, i.CnameRecord, trustCenterID, verificationID, i.CreatedAt)
	}

	writer.Render()
}
