//go:build cli

package integrations

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	api "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/utils/cli/tables"

	openlane "github.com/theopenlane/go-client"
	"github.com/theopenlane/go-client/graphclient"
)

// cmd represents the base integration command when called without any subcommands
var command = &cobra.Command{
	Use:   "integration",
	Short: "the subcommands for working with integrations",
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
	case *graphclient.GetAllIntegrations:
		var nodes []*graphclient.GetAllIntegrations_Integrations_Edges_Node

		for _, i := range v.Integrations.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetIntegrationByID:
		e = v.Integration
	case *graphclient.DeleteIntegration:
		deletedTableOutput(v)
		return nil
	case *openlane.IntegrationProvidersResponse:
		providersTableOutput(v)
		return nil
	case *api.ConfigureIntegrationResponse:
		configureTableOutput(v)
		return nil
	case *api.DeleteIntegrationResponse:
		disconnectTableOutput(v)
		return nil
	case *api.RunIntegrationOperationResponse:
		operationTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Integration

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Integration
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
func tableOutput(out []graphclient.Integration) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Name", "Description", "Kind")
	for _, i := range out {
		writer.AddRow(i.ID, i.Name, *i.Description, i.Kind)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteIntegration) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteIntegration.DeletedID)

	writer.Render()
}

// providerDefinition is a lightweight projection of the definition response used for table rendering
type providerDefinition struct {
	Spec struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Category    string `json:"category"`
		Active      bool   `json:"active"`
	} `json:"spec"`
}

// providersTableOutput prints integration providers in a table format
func providersTableOutput(e *openlane.IntegrationProvidersResponse) {
	var providers []providerDefinition

	if err := json.Unmarshal(e.Providers, &providers); err != nil {
		cobra.CheckErr(err)
	}

	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "DisplayName", "Category", "Active")
	for _, p := range providers {
		writer.AddRow(p.Spec.ID, p.Spec.DisplayName, p.Spec.Category, p.Spec.Active)
	}

	writer.Render()
}

// configureTableOutput prints the configure integration response in a table format
func configureTableOutput(e *api.ConfigureIntegrationResponse) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "Provider", "IntegrationID", "HealthStatus", "HealthSummary", "WebhookEndpointURL")
	writer.AddRow(e.Provider, e.IntegrationID, e.HealthStatus, e.HealthSummary, e.WebhookEndpointURL)

	writer.Render()
}

// disconnectTableOutput prints the disconnect integration response in a table format
func disconnectTableOutput(e *api.DeleteIntegrationResponse) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID", "Message")
	writer.AddRow(e.DeletedID, e.Message)

	writer.Render()
}

// operationTableOutput prints the run operation response in a table format
func operationTableOutput(e *api.RunIntegrationOperationResponse) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "Provider", "Operation", "Status", "Summary")
	writer.AddRow(e.Provider, e.Operation, e.Status, e.Summary)

	writer.Render()
}
