//go:build cli

package entity

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing entity",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "entity id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the entity")
	updateCmd.Flags().StringP("display-name", "s", "", "human friendly name of the entity")
	updateCmd.Flags().StringP("type", "t", "", "type of the entity")
	updateCmd.Flags().StringP("description", "d", "", "description of the entity")
	updateCmd.Flags().StringSliceP("contacts", "c", []string{}, "contact IDs to associate with the entity")
	updateCmd.Flags().StringSlice("domains", []string{}, "domains associated with the entity")
	updateCmd.Flags().String("note", "", "add note about the entity")
	updateCmd.Flags().String("status", "", "status of the entity")
}

// updateValidation validates the required fields for the command
func updateValidation(ctx context.Context) (id string, input openlaneclient.UpdateEntityInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("entity id")
	}

	// validation of required fields for the update command
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	entityType := cmd.Config.String("type")
	if entityType != "" {
		id, err := getEntityTypeID(ctx, entityType)
		cobra.CheckErr(err)

		input.EntityTypeID = &id
	}

	displayName := cmd.Config.String("display-name")
	if displayName != "" {
		input.DisplayName = &displayName
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	contacts := cmd.Config.Strings("contacts")
	if len(contacts) > 0 {
		input.AddContactIDs = contacts
	}

	domains := cmd.Config.Strings("domains")
	if len(domains) > 0 {
		input.AppendDomains = domains
	}

	note := cmd.Config.String("note")
	if note != "" {
		input.Note = &openlaneclient.CreateNoteInput{
			Text: note,
		}
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = &status
	}

	return id, input, nil
}

// update an existing entity in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation(ctx)
	cobra.CheckErr(err)

	o, err := client.UpdateEntity(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
