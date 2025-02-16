package orgsetting

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

// cmd represents the base org setting command when called without any subcommands
var command = &cobra.Command{
	Use:     "organization-setting",
	Aliases: []string{"org-setting"},
	Short:   "the subcommands for working with the organization settings",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the output in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the output and print them in a table format
	switch v := e.(type) {
	case *openlaneclient.GetOrganizationSettings:
		var nodes []*openlaneclient.GetOrganizationSettings_OrganizationSettings_Edges_Node

		for _, i := range v.OrganizationSettings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetAllOrganizationSettings:
		var nodes []*openlaneclient.GetAllOrganizationSettings_OrganizationSettings_Edges_Node

		for _, i := range v.OrganizationSettings.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *openlaneclient.GetOrganizationSettingByID:
		e = v.OrganizationSetting
	case *openlaneclient.UpdateOrganizationSetting:
		e = v.UpdateOrganizationSetting.OrganizationSetting
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []openlaneclient.OrganizationSetting

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in openlaneclient.OrganizationSetting
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
func tableOutput(out []openlaneclient.OrganizationSetting) {
	writer := tables.NewTableWriter(command.OutOrStdout(),
		"ID",
		"OrganizationName",
		"BillingContact",
		"BillingAddress",
		"BillingEmail",
		"BillingPhone",
		"GeoLocation",
		"TaxIdentifier",
		"Tags",
		"Domains",
		"AllowedEmailDomains",
	)
	for _, i := range out {
		writer.AddRow(i.ID,
			i.Organization.Name,
			*i.BillingContact,
			i.BillingAddress.String(),
			*i.BillingEmail,
			*i.BillingPhone,
			*i.GeoLocation,
			*i.TaxIdentifier,
			strings.Join(i.Tags, ", "),
			strings.Join(i.Domains, ", "),
			strings.Join(i.AllowedEmailDomains, ", "))
	}

	writer.Render()
}
