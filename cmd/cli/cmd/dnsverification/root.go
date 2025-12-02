//go:build cli

package dnsverification

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client"
)

// command represents the base dnsverification command when called without any subcommands
var command = &cobra.Command{
	Use:     "dns-verification",
	Aliases: []string{"dnsverification"},
	Short:   "the subcommands for working with DNS verification records",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the DNS verifications in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the DNS verifications and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.UpdateDNSVerification:
		nodes := []*openlaneclient.UpdateDNSVerification_UpdateDNSVerification_DNSVerification{
			&v.UpdateDNSVerification.DNSVerification,
		}

		e = nodes

	case *openlaneclient.CreateDNSVerification:
		nodes := []*openlaneclient.CreateDNSVerification_CreateDNSVerification_DNSVerification{
			&v.CreateDNSVerification.DNSVerification,
		}

		e = nodes
	case *openlaneclient.GetDNSVerificationByID:
		nodes := []*openlaneclient.GetDNSVerificationByID_DNSVerification{
			&v.DNSVerification,
		}

		e = nodes
	case *openlaneclient.GetAllDNSVerifications:
		var nodes []*openlaneclient.GetAllDNSVerifications_DNSVerifications_Edges_Node

		for _, i := range v.DNSVerifications.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetDNSVerifications:
		var nodes []*openlaneclient.GetDNSVerifications_DNSVerifications_Edges_Node
		for _, i := range v.DNSVerifications.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.DeleteDNSVerification:
		// For delete operations, just print the deleted ID
		return cmd.JSONPrint([]byte(`{"deletedID": "` + v.DeleteDNSVerification.DeletedID + `"}`))
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.DNSVerification

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.DNSVerification
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
func tableOutput(out []openlaneclient.DNSVerification) {
	// create a table writer
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Cloudflare ID", "DNS Status", "SSL Status", "Created At")
	for _, i := range out {
		writer.AddRow(i.ID, i.CloudflareHostnameID, i.DNSVerificationStatus, i.AcmeChallengeStatus, i.CreatedAt)
	}

	writer.Render()
}
