//go:build cli

package entity

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	openlane "github.com/theopenlane/go-client"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new entity",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the entity")
	createCmd.Flags().StringP("display-name", "s", "", "human friendly name of the entity")
	createCmd.Flags().StringP("type", "t", "", "type of the entity")
	createCmd.Flags().StringP("description", "d", "", "description of the entity")
	createCmd.Flags().String("status", "", "status of the entity")
	createCmd.Flags().StringSlice("domains", []string{}, "domains associated with the entity")
	createCmd.Flags().String("note", "", "note about the entity")
	createCmd.Flags().StringSlice("tags", []string{}, "tags associated with the entity")
}

// createValidation validates the required fields for the command
func createValidation(ctx context.Context) (input graphclient.CreateEntityInput, err error) {
	// validation of required fields for the create command
	name := cmd.Config.String("name")
	displayName := cmd.Config.String("display-name")

	if name == "" && displayName == "" {
		return input, cmd.NewRequiredFieldMissingError("entity name or display name")
	}

	if name != "" {
		input.Name = &name
	}

	if displayName != "" {
		input.DisplayName = &displayName
	}

	entityType := cmd.Config.String("type")
	if entityType != "" {
		// get the entity type id
		id, err := getEntityTypeID(ctx, entityType)
		cobra.CheckErr(err)

		fmt.Println("Entity Type ID: ", id)

		input.EntityTypeID = &id
	}

	description := cmd.Config.String("description")
	if description != "" {
		input.Description = &description
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = &status
	}

	domains := cmd.Config.Strings("domains")
	if len(domains) > 0 {
		input.Domains = domains
	}

	note := cmd.Config.String("note")
	if note != "" {
		input.Note = &graphclient.CreateNoteInput{
			Text: note,
		}
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	return input, nil
}

// create a new entity
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation(ctx)
	cobra.CheckErr(err)

	o, err := client.CreateEntity(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}

func getEntityTypeID(ctx context.Context, name string) (string, error) {
	client, err := cmd.SetupClientWithAuth(ctx)
	cobra.CheckErr(err)

	where := &openlane.EntityTypeWhereInput{
		Name: &name,
	}

	o, err := client.GetEntityTypes(ctx, cmd.First, cmd.Last, where)
	cobra.CheckErr(err)

	if len(o.EntityTypes.Edges) == 0 || len(o.EntityTypes.Edges) > 1 {
		return "", fmt.Errorf("%w: entity type '%s' not found", cmd.ErrNotFound, name)
	}

	return o.EntityTypes.Edges[0].Node.ID, nil
}
