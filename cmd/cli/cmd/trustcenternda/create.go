//go:build cli

package trustcenternda

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/objects/storage"
	openlaneclient "github.com/theopenlane/go-client"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center NDA",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("trust-center-id", "t", "", "trust center id for the NDA")
	createCmd.Flags().StringP("nda-file", "f", "", "NDA file")
}

// createValidation validates the required fields for the command
func createValidation() (openlaneclient.CreateTrustCenterNDAInput, *graphql.Upload, error) {
	input := openlaneclient.CreateTrustCenterNDAInput{}

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = trustCenterID
	} else {
		return input, nil, cmd.NewRequiredFieldMissingError("trust center id")
	}

	ndaFile := cmd.Config.String("nda-file")
	if ndaFile == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("nda file")
	}
	ndaUploadFile, err := storage.NewUploadFile(ndaFile)
	if err != nil {
		return input, nil, err
	}
	ndaUpload := &graphql.Upload{
		File:        ndaUploadFile.RawFile,
		Filename:    ndaUploadFile.OriginalName,
		Size:        ndaUploadFile.Size,
		ContentType: ndaUploadFile.ContentType,
	}

	return input, ndaUpload, nil
}

// create a new trust center NDA
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, uploadFile, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateTrustCenterNda(ctx, input, []*graphql.Upload{uploadFile})
	cobra.CheckErr(err)

	return consoleOutput(o)
}
