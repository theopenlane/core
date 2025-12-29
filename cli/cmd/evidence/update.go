//go:build cli

package evidence

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing evidence",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "evidence id to update")

	// command line flags for the create command
	updateCmd.Flags().StringP("name", "n", "", "name of the evidence")
	updateCmd.Flags().StringP("description", "d", "", "description of the evidence")
	updateCmd.Flags().StringArrayP("add-programs", "p", []string{}, "program of the evidence")
	updateCmd.Flags().StringArray("add-controls", []string{}, "ids of controls to link to the evidence")
	updateCmd.Flags().StringArray("add-subcontrols", []string{}, "ids of subcontrols to link to the evidence")
	updateCmd.Flags().StringArray("add-control-objectives", []string{}, "ids of control objectives to link to the evidence")
	updateCmd.Flags().StringP("collection-procedure", "c", "", "steps taken to collect the evidence")
	updateCmd.Flags().StringP("source", "s", "", "system source of the evidence")
	updateCmd.Flags().BoolP("is-automated", "a", false, "whether the evidence was collected automatically")
	updateCmd.Flags().StringP("url", "u", "", "remote url of the evidence, used in place of uploading files")
	updateCmd.Flags().StringArrayP("files", "f", []string{}, "files to upload as evidence")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateEvidenceInput, uploads []*graphql.Upload, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, uploads, cmd.NewRequiredFieldMissingError("evidence id")
	}

	// validation of required fields for the update command
	// output the input struct with the required fields and optional fields based on the command line flags
	name := cmd.Config.String("name")
	description := cmd.Config.String("description")
	programs := cmd.Config.Strings("add-programs")
	controls := cmd.Config.Strings("add-controls")
	subcontrols := cmd.Config.Strings("add-subcontrols")
	controlObjectives := cmd.Config.Strings("add-control-objectives")
	collectionProcedure := cmd.Config.String("collection-procedure")
	source := cmd.Config.String("source")
	isAutomated := cmd.Config.Bool("is-automated")
	url := cmd.Config.String("url")
	files := cmd.Config.Strings("files")

	if name != "" {
		input.Name = &name
	}

	if description != "" {
		input.Description = &description
	}

	if len(programs) > 0 {
		input.AddProgramIDs = programs
	}

	if len(controls) > 0 {
		input.AddControlIDs = controls
	}

	if len(subcontrols) > 0 {
		input.AddSubcontrolIDs = subcontrols
	}

	if len(controlObjectives) > 0 {
		input.AddControlObjectiveIDs = controlObjectives
	}

	if collectionProcedure != "" {
		input.CollectionProcedure = &collectionProcedure
	}

	if source != "" {
		input.Source = &source
	}

	input.IsAutomated = &isAutomated

	if url != "" {
		input.URL = &url
	}

	// parse the files to upload
	for _, file := range files {
		u, err := storage.NewUploadFile(file)
		if err != nil {
			return id, input, uploads, err
		}

		uploads = append(uploads, &graphql.Upload{
			File:        u.RawFile,
			Filename:    u.OriginalName,
			Size:        u.Size,
			ContentType: u.ContentType,
		})
	}

	return
}

// update an existing evidence in the platform
func update(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, uploads, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateEvidence(ctx, id, input, uploads)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
