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
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
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

	reader := io.LimitReader(resp.Body, int64(m.Client().EntConfig.MaxSchemaImportSize))
	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err) // nolint:err113
	}

	mimeType := strings.ToLower(resp.Header.Get("Content-Type"))

	switch mimeType {
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain", "text/markdown", "text/plain; charset=utf-8":

		content, err := storage.ParseDocument(strings.NewReader(string(buf)), mimeType)
		if err != nil {
			return fmt.Errorf("failed to parse document: %w", err)
		}

		p := bluemonday.UGCPolicy()

		m.SetURL(downloadURL)
		m.SetDetails(p.Sanitize(fmt.Sprintf("%v", content)))

		return nil

	default:

		return fmt.Errorf("unspupported content-type ( %s)", mimeType) // nolint:err113
	}
}

// checkProcedureFile checks for uploaded procedure files in context
func checkProcedureFile[T utils.GenericMutation](ctx context.Context, m T) (context.Context, error) {
	key := "procedureFile"

	// get the file from the context, if it exists
	file, err := pkgobjects.FilesFromContextWithKey(ctx, key)
	if err != nil {
		return ctx, err
	}

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

		ctx = pkgobjects.UpdateFileInContextByKey(ctx, key, file[0])
	}

	return ctx, nil
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

				var err error
				ctx, err = checkProcedureFile(ctx, m)
				if err != nil {
					return nil, err
				}

			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}
