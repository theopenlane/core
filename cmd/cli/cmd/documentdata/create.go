//go:build cli

package documentdata

import (
	"context"
	"encoding/json"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new document data",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("template-id", "t", "", "template id to use for the document data")
	createCmd.Flags().StringP("data", "d", "", "json data to use for the document data")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateDocumentDataInput, err error) {
	templateID := cmd.Config.String("template-id")
	if templateID == "" {
		return input, cmd.NewRequiredFieldMissingError("template id")
	}

	input.TemplateID = lo.ToPtr(templateID)

	dataFile := cmd.Config.String("data")
	if dataFile == "" {
		return input, cmd.NewRequiredFieldMissingError("data")
	}

	dataStr, err := os.ReadFile(dataFile)
	if err != nil {
		return input, err
	}

	var jsonData map[string]any
	if err = json.Unmarshal(dataStr, &jsonData); err != nil {
		return input, err
	}
	input.Data = jsonData

	return input, nil
}

// create a new document data
func create(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateDocumentData(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
