//go:build cli

package events

import (
	"context"
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new event",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("type", "t", "", "type of the event")
	createCmd.Flags().StringP("metadata", "m", "", "metadata for the event")
	createCmd.Flags().StringSliceP("user-ids", "u", []string{}, "user id associated with the event")
	createCmd.Flags().StringP("event-json", "j", "", "json payload for the template")
}

// createValidation validates the required fields for the command
func createValidation() (input openlaneclient.CreateEventInput, err error) {
	input.EventType = cmd.Config.String("type")
	if input.EventType == "" {
		return input, cmd.NewRequiredFieldMissingError("type")
	}

	userIDs := cmd.Config.Strings("user-ids")
	if userIDs != nil {
		input.UserIDs = userIDs
	}

	eventJSON := cmd.Config.String("event-json")
	if eventJSON != "" {
		var data []byte

		if data, err = os.ReadFile(eventJSON); err != nil {
			cobra.CheckErr(err)
		}

		parsedMessage, err := cmd.ParseBytes(data)
		cobra.CheckErr(err)

		input.Metadata = parsedMessage
	}

	metadata := cmd.Config.String("metadata")
	if metadata != "" {
		err := json.Unmarshal([]byte(metadata), &input.Metadata)
		cobra.CheckErr(err)
	}

	return input, nil
}

// create a new event
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateEvent(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
