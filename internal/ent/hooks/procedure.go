package hooks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/microcosm-cc/bluemonday"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/objects"
)

type importSchemaMutation interface {
	SetName(string)
	SetDetails(string)
	SetFileID(string)
	FileID() (string, bool)
	SetURL(string)
	URL() (string, bool)
	Client() *generated.Client
}

const defaultImportTimeout = time.Second * 10

var client = &http.Client{
	Timeout: defaultImportTimeout,
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

	// 256KB limit
	// do we need to add this to the config?
	const maxBodySize = 256 * 1024

	reader := io.LimitReader(resp.Body, maxBodySize)
	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err) // nolint:err113
	}

	mimeType := strings.ToLower(resp.Header.Get("Content-Type"))

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

// importFileToSchema handles the common logic for processing uploaded files
// and setting the name and details from the file content
func importFileToSchema(ctx context.Context, m importSchemaMutation, objectManager *objects.Objects, fileKey string) error {
	file, _ := objects.FilesFromContextWithKey(ctx, fileKey)

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

	content, err := objectManager.Storage.Download(ctx, &objects.DownloadFileOptions{
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

// HookProcedure checks to see if we have an uploaded file.
// If we do, use that as the details of the procedure. and also
// use the name of the file
func HookProcedure() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ProcedureFunc(func(ctx context.Context, m *generated.ProcedureMutation) (generated.Value, error) {
			_, exists := m.URL()

			switch exists {
			case true:

				if err := importURLToSchema(m); err != nil {
					return nil, err
				}

			default:

				ctx, err := checkProcedureFile(ctx, m)
				if err != nil {
					return nil, err
				}

				if err := importFileToSchema(ctx, m, m.ObjectManager, "procedureFile"); err != nil {
					return nil, err
				}

			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

func checkProcedureFile[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "procedureFile"

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
		file[0].Parent.Type = m.Type()

		ctx = objects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
}
