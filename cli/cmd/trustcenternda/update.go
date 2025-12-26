//go:build cli

package trustcenternda

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/pkg/objects/storage"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing trust center NDA",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	// command line flags for the update command
	updateCmd.Flags().StringP("id", "i", "", "trust center id for the NDA to update")
	updateCmd.Flags().StringP("nda-file", "f", "", "NDA file (optional - if not provided, existing NDA will be kept)")
}

// updateValidation validates the required fields for the command
func updateValidation() (string, *graphql.Upload, error) {
	id := cmd.Config.String("id")
	if id == "" {
		return "", nil, cmd.NewRequiredFieldMissingError("id")
	}

	ndaFile := cmd.Config.String("nda-file")
	if ndaFile == "" {
		// No file provided, return nil upload (this is allowed for update)
		return id, nil, nil
	}

	ndaUploadFile, err := storage.NewUploadFile(ndaFile)
	if err != nil {
		return "", nil, err
	}
	ndaUpload := &graphql.Upload{
		File:        ndaUploadFile.RawFile,
		Filename:    ndaUploadFile.OriginalName,
		Size:        ndaUploadFile.Size,
		ContentType: ndaUploadFile.ContentType,
	}

	return id, ndaUpload, nil
}

// update an existing trust center NDA
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, uploadFile, err := updateValidation()
	cobra.CheckErr(err)

	var templateFiles []*graphql.Upload
	if uploadFile != nil {
		templateFiles = []*graphql.Upload{uploadFile}
	}

	o, err := client.UpdateTrustCenterNda(ctx, id, templateFiles)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
