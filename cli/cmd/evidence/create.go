//go:build cli

package evidence

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/cobra"

	"github.com/theopenlane/cli/cmd"
	"github.com/theopenlane/core/pkg/objects/storage"
	openlane "github.com/theopenlane/go-client"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new evidence",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	// command line flags for the create command
	createCmd.Flags().StringP("name", "n", "", "name of the evidence")
	createCmd.Flags().StringP("description", "d", "", "description of the evidence")
	createCmd.Flags().StringArrayP("programs", "p", []string{}, "program of the evidence")
	createCmd.Flags().StringArray("controls", []string{}, "ids of controls to link to the evidence")
	createCmd.Flags().StringArray("subcontrols", []string{}, "ids of subcontrols to link to the evidence")
	createCmd.Flags().StringArray("controlRefs", []string{}, "standard short name and ref code, e.g. (SOC2:CC1.1) of controls to link to the evidence")
	createCmd.Flags().StringArray("subcontrolRefs", []string{}, "standard short name and ref code, e.g. (SOC2:CC1.POF1) of subcontrols to link to the evidence")
	createCmd.Flags().StringArray("control-objectives", []string{}, "ids of control objectives to link to the evidence")
	createCmd.Flags().StringP("collection-procedure", "c", "", "steps taken to collect the evidence")
	createCmd.Flags().StringP("source", "s", "", "system source of the evidence")
	createCmd.Flags().BoolP("is-automated", "a", false, "whether the evidence was collected automatically")
	createCmd.Flags().StringP("url", "u", "", "remote url of the evidence, used in place of uploading files")
	createCmd.Flags().StringArrayP("files", "f", []string{}, "files to upload as evidence")
}

// createValidation validates the required fields for the command
func createValidation(ctx context.Context, client *openlane.Client) (input graphclient.CreateEvidenceInput, uploads []*graphql.Upload, err error) {
	// validation of required fields for the create command
	// output the input struct with the required fields and optional fields based on the command line flags
	input.Name = cmd.Config.String("name")
	description := cmd.Config.String("description")
	programs := cmd.Config.Strings("programs")
	controls := cmd.Config.Strings("controls")
	subcontrols := cmd.Config.Strings("subcontrols")
	controlRefs := cmd.Config.Strings("controlRefs")
	subcontrolRefs := cmd.Config.Strings("subcontrolRefs")
	controlObjectives := cmd.Config.Strings("control-objectives")
	collectionProcedure := cmd.Config.String("collection-procedure")
	source := cmd.Config.String("source")
	isAutomated := cmd.Config.Bool("is-automated")
	url := cmd.Config.String("url")
	files := cmd.Config.Strings("files")

	if input.Name == "" {
		return input, nil, cmd.NewRequiredFieldMissingError("evidence name")
	}

	if description != "" {
		input.Description = &description
	}

	if len(programs) > 0 {
		input.ProgramIDs = programs
	}

	if len(controls) > 0 {
		input.ControlIDs = controls
	} else if len(controlRefs) > 0 {
		for _, ref := range controlRefs {
			parts := strings.Split(ref, ":")
			if len(parts) != 2 {
				cmd.NewInvalidFieldError("control ref", ref)
			}

			pageSize := int64(100)
			var after *string

			for {
				control, err := client.GetControls(ctx, &pageSize, nil, after, nil, &graphclient.ControlWhereInput{
					ReferenceFramework: &parts[0],
					RefCode:            &parts[1],
				}, nil)
				cobra.CheckErr(err)

				if len(control.Controls.Edges) == 0 {
					return input, nil, cmd.NewInvalidFieldError("control ref", ref)
				}

				input.ControlIDs = append(input.ControlIDs, control.Controls.Edges[0].Node.ID)

				if !control.Controls.PageInfo.HasNextPage {
					break
				}

				after = control.Controls.PageInfo.EndCursor
			}
		}
	}

	if len(subcontrols) > 0 {
		input.SubcontrolIDs = subcontrols
	} else if len(subcontrolRefs) > 0 {
		for _, ref := range subcontrolRefs {
			parts := strings.Split(ref, ":")
			if len(parts) != 2 {
				cmd.NewInvalidFieldError("subcontrol ref", ref)
			}

			pageSize := int64(100)
			var after *string

			for {
				control, err := client.GetSubcontrols(ctx, &pageSize, nil, after, nil, &graphclient.SubcontrolWhereInput{
					ReferenceFramework: &parts[0],
					RefCode:            &parts[1],
				}, nil)
				cobra.CheckErr(err)

				if len(control.Subcontrols.Edges) == 0 {
					return input, nil, cmd.NewInvalidFieldError("control ref", ref)
				}

				input.SubcontrolIDs = append(input.SubcontrolIDs, control.Subcontrols.Edges[0].Node.ID)

				if !control.Subcontrols.PageInfo.HasNextPage {
					break
				}

				after = control.Subcontrols.PageInfo.EndCursor
			}
		}
	}

	if len(controlObjectives) > 0 {
		input.ControlObjectiveIDs = controlObjectives
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
			return input, nil, err
		}

		uploads = append(uploads, &graphql.Upload{
			File:        u.RawFile,
			Filename:    u.OriginalName,
			Size:        u.Size,
			ContentType: u.ContentType,
		})
	}

	return input, uploads, nil
}

// create a new evidence
func create(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, uploads, err := createValidation(ctx, client)
	cobra.CheckErr(err)

	o, err := client.CreateEvidence(ctx, input, uploads)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
