package hooks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"entgo.io/ent"

	"github.com/microcosm-cc/bluemonday"
	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

type detailsMutation interface {
	SetSummary(string)
	Details() (string, bool)
	Client() *generated.Client
}

type importSchemaMutation interface {
	utils.GenericMutation
	detailsMutation

	SetName(string)
	SetDetails(string)
	SetFileID(string)
	FileID() (string, bool)
	SetURL(string)
	URL() (string, bool)
}

// HookSummarizeDetails summarizes the policy and produces a short human readable copy
func HookSummarizeDetails() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut := m.(detailsMutation)

			details, ok := mut.Details()
			if !ok || details == "" {
				return next.Mutate(ctx, m)
			}

			summarizer := mut.Client().Summarizer
			if summarizer == nil {
				return nil, errors.New("summarizer client not found") //nolint:err113
			}

			summary, err := summarizer.Summarize(ctx, details)
			if err != nil {
				return nil, err
			}

			mut.SetSummary(summary)

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}

// HookImportDocument  checks to see if we have an uploaded file.
// If we do, use that as the details of the document and also
// use the name of the file
func HookImportDocument() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(importSchemaMutation)
			if !ok {
				log.Info().Msg("import document hook used on unsupported mutation type")

				return next.Mutate(ctx, m)
			}

			_, exists := mut.URL()

			switch exists {
			case true:
				if err := importURLToSchema(mut); err != nil {
					return nil, err
				}

			default:
				ctx, err := checkDocumentFile(ctx, mut)
				if err != nil {
					return nil, err
				}

				if err := importFileToSchema(ctx, mut); err != nil {
					return nil, err
				}

			}

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}

func mutationToFileKey(m importSchemaMutation) string {
	return fmt.Sprintf("%sFile", strcase.LowerCamelCase(m.Type()))
}

func checkDocumentFile[T importSchemaMutation](ctx context.Context, m T) (context.Context, error) {
	key := mutationToFileKey(m)

	// get the file from the context, if it exists
	file, _ := objects.FilesFromContextWithKey(ctx, key)

	// return early if no file is provided
	if file == nil {
		return ctx, nil
	}

	// we should only have one file
	if len(file) > 1 {
		return ctx, ErrNotSingularUpload
	}

	// this should always be true, but check just in case
	if file[0].FieldName == key {
		file[0].Parent.ID, _ = m.ID()
		file[0].Parent.Type = strcase.SnakeCase(m.Type())

		ctx = objects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
}

// importFileToSchema handles the common logic for processing uploaded files
// and setting the name and details from the file content
func importFileToSchema[T importSchemaMutation](ctx context.Context, m T) error {
	key := mutationToFileKey(m)

	file, _ := objects.FilesFromContextWithKey(ctx, key)

	if len(file) == 0 {
		return nil
	}

	currentFileID, exists := m.FileID()
	// delete the existing file since we are replacing it
	// and we do not want old files laying around
	if exists {
		if err := m.Client().File.DeleteOneID(currentFileID).Exec(ctx); err != nil {
			return err
		}
	}

	content, err := m.Client().ObjectManager.Storage.Download(ctx, &objects.DownloadFileOptions{
		FileName: file[0].UploadedFileName,
	})
	if err != nil {
		return err
	}

	parsedContent, err := objects.ParseDocument(content.File, file[0].MimeType)
	if err != nil {
		return err
	}

	p := bluemonday.UGCPolicy()

	m.SetName(file[0].OriginalName)
	m.SetFileID(file[0].ID)
	m.SetDetails(p.Sanitize(parsedContent))

	return nil
}

const defaultImportTimeout = time.Second * 10

var client = &http.Client{
	Timeout: defaultImportTimeout,
}

func detectMimeTypeFromContent(content []byte, fallbackMimeType string) string {
	mimeType := http.DetectContentType(content)

	if mimeType == "" {
		return strings.ToLower(strings.TrimSpace(fallbackMimeType))
	}

	return mimeType
}

func importURLToSchema(m importSchemaMutation) error {
	downloadURL, exists := m.URL()
	if !exists {
		return nil
	}

	_, err := url.Parse(downloadURL)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultImportTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d not an accepted status code. Only 200 accepted", resp.StatusCode) // nolint:err113
	}

	reader := io.LimitReader(resp.Body, int64(m.Client().EntConfig.MaxSchemaImportSize))
	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err) // nolint:err113
	}

	fallbackMimeType := resp.Header.Get("Content-Type")
	mimeType := detectMimeTypeFromContent(buf, fallbackMimeType)

	switch mimeType {
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain", "text/markdown", "text/plain; charset=utf-8":

		content, err := objects.ParseDocument(buf, mimeType)
		if err != nil {
			return fmt.Errorf("failed to parse document: %w", err)
		}

		p := bluemonday.UGCPolicy()

		m.SetURL(downloadURL)
		m.SetDetails(p.Sanitize(content))

		return nil

	default:
		return fmt.Errorf("unspupported content-type ( %s)", mimeType) // nolint:err113
	}
}
