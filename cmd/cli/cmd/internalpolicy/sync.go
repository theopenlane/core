//go:build cli

package internalpolicy

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cmd/cli/cmd"
	openlaneclient "github.com/theopenlane/go-client/genclient"
	"github.com/theopenlane/shared/objects/storage"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync an internal policy from a local file",
	Run: func(cmd *cobra.Command, args []string) {
		err := sync(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(syncCmd)

	syncCmd.Flags().StringP("id", "i", "", "policy id to update, if not provided will attempt to read from frontmatter and will create a new policy if not found")
	syncCmd.Flags().StringP("file", "f", "", "local path to file to upload as the procedure details")
}

// syncValidation validates the required fields for the command
func syncValidation() (id string, detailsFile *graphql.Upload, document *storage.ParsedDocument, err error) {
	detailsFileLoc := cmd.Config.String("file")
	if detailsFileLoc == "" {
		return id, nil, nil, cmd.NewRequiredFieldMissingError("file")
	}

	file, err := storage.NewUploadFile(detailsFileLoc)
	if err != nil {
		return id, nil, nil, err
	}

	detailsFile = &graphql.Upload{
		File:        file.RawFile,
		Filename:    file.OriginalName,
		Size:        file.Size,
		ContentType: file.ContentType,
	}

	id = cmd.Config.String("id")
	if id == "" {
		// get the id from the file frontmatter if present
		document, err = storage.ParseDocument(file.RawFile, file.ContentType)
		if err != nil {
			return id, nil, document, err
		}

		if document.Frontmatter != nil && document.Frontmatter.OpenlaneID != "" {
			id = document.Frontmatter.OpenlaneID
		}
	}

	// re-read to the original file as the raw file was already read during parsing
	file, err = storage.NewUploadFile(detailsFileLoc)
	if err != nil {
		return id, nil, nil, err
	}

	detailsFile = &graphql.Upload{
		File:        file.RawFile,
		Filename:    file.OriginalName,
		Size:        file.Size,
		ContentType: file.ContentType,
	}

	return id, detailsFile, document, nil
}

// sync an existing internal policy in the platform
func sync(ctx context.Context) error {
	// attempt to setup with token, otherwise fall back to JWT with session
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		// setup http client
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, detailsFile, document, err := syncValidation()
	cobra.CheckErr(err)

	if id == "" {
		fmt.Println("→ no id found in request, creating new policy")

		o, err := client.CreateUploadInternalPolicy(ctx, *detailsFile, nil)
		cobra.CheckErr(err)

		if document == nil {
			document = &storage.ParsedDocument{
				Frontmatter: &storage.Frontmatter{},
				Data:        []byte{},
			}
		} else if document.Frontmatter == nil {
			document.Frontmatter = &storage.Frontmatter{}
		}

		document.Frontmatter.OpenlaneID = o.CreateUploadInternalPolicy.InternalPolicy.ID
		document.Frontmatter.Title = o.CreateUploadInternalPolicy.InternalPolicy.Name
		document.Frontmatter.Status = o.CreateUploadInternalPolicy.InternalPolicy.Status.String()

		fmt.Println("→ updating document with frontmatter")

		if err := writeFrontmatter(cmd.Config.String("file"), document); err != nil {
			return err
		}

		return consoleOutput(o)
	}

	fmt.Println("→ updating existing policy with id:", id)

	o, err := client.UpdateInternalPolicyWithFile(ctx, id, *detailsFile, openlaneclient.UpdateInternalPolicyInput{})
	cobra.CheckErr(err)

	document.Frontmatter.Title = o.UpdateInternalPolicy.InternalPolicy.Name
	document.Frontmatter.Status = o.UpdateInternalPolicy.InternalPolicy.Status.String()

	fmt.Println("→ updating document with frontmatter")

	if err := writeFrontmatter(cmd.Config.String("file"), document); err != nil {
		return err
	}

	return consoleOutput(o)
}

func writeFrontmatter(path string, doc *storage.ParsedDocument) error {
	var b strings.Builder
	b.WriteString("---\n")

	y, err := yaml.Marshal(doc.Frontmatter)
	if err != nil {
		return err
	}

	b.Write(y)
	b.WriteString("---\n\n")

	if doc.Data != nil {
		switch v := doc.Data.(type) {
		case []byte:
			b.Write(v)
		case string:
			b.WriteString(v)
		}
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}
