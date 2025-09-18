//go:build cli

package trustcenterdoc

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new trust center document",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("trust-center-id", "t", "", "trust center id for the document")
	createCmd.Flags().StringP("title", "n", "", "title of the document")
	createCmd.Flags().StringP("category", "c", "", "category of the document")
	createCmd.Flags().StringP("visibility", "v", "NOT_VISIBLE", "visibility of the document (NOT_VISIBLE, PROTECTED, PUBLICLY_VISIBLE)")
	createCmd.Flags().StringSliceP("tags", "g", []string{}, "tags associated with the document")
	createCmd.Flags().StringP("file", "f", "", "file to upload as the trust center document (required)")
}

// createValidation validates the required fields for the command
func createValidation() (openlaneclient.CreateTrustCenterDocInput, *graphql.Upload, error) {
	input := openlaneclient.CreateTrustCenterDocInput{}

	title := cmd.Config.String("title")
	if title == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("title")
	}

	category := cmd.Config.String("category")
	if category == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("category")
	}

	filePath := cmd.Config.String("file")
	if filePath == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("file")
	}

	input.Title = title
	input.Category = category

	trustCenterID := cmd.Config.String("trust-center-id")
	if trustCenterID != "" {
		input.TrustCenterID = &trustCenterID
	}

	visibility := cmd.Config.String("visibility")
	if visibility != "" {
		visibilityEnum := enums.ToTrustCenterDocumentVisibility(visibility)
		if visibilityEnum != nil {
			input.Visibility = visibilityEnum
		}
	}

	tags := cmd.Config.Strings("tags")
	if len(tags) > 0 {
		input.Tags = tags
	}

	// Handle file upload
	u, err := objects.NewUploadFile(filePath)
	if err != nil {
		return input, nil, err
	}

	fileUpload := &graphql.Upload{
		File:        u.File,
		Filename:    u.Filename,
		Size:        u.Size,
		ContentType: u.ContentType,
	}

	return input, fileUpload, nil
}

// create a new trust center document
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, fileUpload, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateTrustCenterDoc(ctx, input, *fileUpload)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
