//go:build cli

package internalpolicy

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing internal policy",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "policy id to update")

	// command line flags for the update command
	updateCmd.Flags().StringP("name", "n", "", "name of the policy")
	updateCmd.Flags().StringP("details", "d", "", "description of the policy")
	updateCmd.Flags().StringP("status", "s", "", "status of the policy")
	updateCmd.Flags().StringP("type", "t", "", "type of the policy")
	updateCmd.Flags().StringP("revision", "v", "v0.1", "revision of the policy")
	updateCmd.Flags().StringP("file", "f", "", "local path to file to upload as the procedure details")
	updateCmd.Flags().StringP("editor-group-id", "g", "", "editor group id")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateInternalPolicyInput, detailsFile *graphql.Upload, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, nil, cmd.NewRequiredFieldMissingError("internal policy id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	name := cmd.Config.String("name")
	if name != "" {
		input.Name = &name
	}

	details := cmd.Config.String("details")
	if details != "" {
		input.Details = &details
	}

	status := cmd.Config.String("status")
	if status != "" {
		input.Status = enums.ToDocumentStatus(status)
	}

	policyType := cmd.Config.String("type")
	if policyType != "" {
		input.InternalPolicyKindName = &policyType
	}

	revision := cmd.Config.String("revision")
	if revision != "" {
		input.Revision = &revision
	}

	editorGroupID := cmd.Config.String("editor-group-id")
	if editorGroupID != "" {
		input.AddEditorIDs = []string{editorGroupID}
	}

	detailsFileLoc := cmd.Config.String("file")
	if detailsFileLoc != "" {
		file, err := storage.NewUploadFile(detailsFileLoc)
		if err != nil {
			return id, input, nil, err
		}

		detailsFile = &graphql.Upload{
			File:        file.RawFile,
			Filename:    file.OriginalName,
			Size:        file.Size,
			ContentType: file.ContentType,
		}
	}

	return id, input, detailsFile, nil
}

// update an existing internal policy in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, detailsFile, err := updateValidation()
	cobra.CheckErr(err)

	if detailsFile != nil {
		o, err := client.UpdateInternalPolicyWithFile(ctx, id, *detailsFile, input)
		cobra.CheckErr(err)

		return consoleOutput(o)
	}

	o, err := client.UpdateInternalPolicy(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
