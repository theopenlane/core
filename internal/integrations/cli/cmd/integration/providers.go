//go:build examples

package integration

import (
	"context"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/internal/integrations/cli/cmd"
)

// providerDefinition is the minimal projection of the definition response
// used to render a provider list in table form
type providerDefinition struct {
	Spec struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Category    string `json:"category"`
		Active      bool   `json:"active"`
	} `json:"spec"`
}

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "list available integration provider definitions",
	RunE: func(c *cobra.Command, _ []string) error {
		return listProviders(c.Context())
	},
}

func init() {
	command.AddCommand(providersCmd)
}

// listProviders fetches all provider definitions and renders them
func listProviders(ctx context.Context) error {
	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	resp, err := client.ListIntegrationProviders(ctx)
	if err != nil {
		return err
	}

	if cmd.OutputFormat() == cmd.JSONOutput {
		return cmd.RenderJSON(resp)
	}

	var providers []providerDefinition
	if err := json.Unmarshal(resp.Providers, &providers); err != nil {
		return err
	}

	rows := make([][]string, 0, len(providers))
	for _, p := range providers {
		active := "false"
		if p.Spec.Active {
			active = "true"
		}

		rows = append(rows, []string{p.Spec.ID, p.Spec.DisplayName, p.Spec.Category, active})
	}

	return cmd.RenderTable(resp, []string{"ID", "DisplayName", "Category", "Active"}, rows)
}
