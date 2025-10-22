package hooks

import (
	"bytes"
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
	"github.com/theopenlane/core/pkg/objects/storage"
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

// HookSummarizeDetails is an ent hook that summarizes long details fields into a short human readable summary
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

// HookImportDocument is an ent hook that imports document content from either an uploaded file or a provided URL
// If a file is uploaded it becomes the source of the details and sets the document name to the original filename
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
				// Derive the expected file key for this mutation and attach parent metadata to uploaded files
				key := mutationToFileKey(mut)
				adapter := objects.NewGenericMutationAdapter(mut,
					func(mm importSchemaMutation) (string, bool) { return mm.ID() },
					func(mm importSchemaMutation) string { return mm.Type() },
				)

				var err error
				ctx, err = objects.ProcessFilesForMutation(ctx, adapter, key)
				if err != nil {
					return nil, err
				}

				// Parse the uploaded file and write values into the mutation
				if err := importFileToSchema(ctx, mut); err != nil {
					return nil, err
				}
			}

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}

// mutationToFileKey is a helper that converts a mutation type into the expected upload field name
func mutationToFileKey(m importSchemaMutation) string {
	return fmt.Sprintf("%sFile", strcase.LowerCamelCase(m.Type()))
}

// importFileToSchema is a helper that reads an uploaded file from context, downloads it from storage, parses it,
// sanitizes the content and sets the document name, fileID and details on the mutation
func importFileToSchema[T importSchemaMutation](ctx context.Context, m T) error {
	key := mutationToFileKey(m)

	file, _ := objects.FilesFromContextWithKey(ctx, key)

	if len(file) == 0 {
		return nil
	}

	currentFileID, exists := m.FileID()
	// If the mutation already has a file, delete the old record to avoid leaving stale files around
	if exists {
		if err := m.Client().File.DeleteOneID(currentFileID).Exec(ctx); err != nil {
			return err
		}
	}

	// Fallback: Download the uploaded file contents using the object storage service and parse into details
	downloaded, err := m.Client().ObjectManager.Download(ctx, nil, &file[0], &objects.DownloadOptions{
		FileName:    file[0].OriginalName,
		ContentType: file[0].ContentType,
	})
	if err != nil {
		return err
	}

	parsedContent, err := storage.ParseDocument(bytes.NewReader(downloaded.File), file[0].ContentType)
	if err != nil {
		return err
	}

	p := bluemonday.UGCPolicy()

	m.SetName(filenameToTitle(file[0].OriginalName))
	m.SetFileID(file[0].ID)

	var detailsStr string
	switch v := parsedContent.(type) {
	case string:
		detailsStr = v
	case []byte:
		detailsStr = string(v)
	default:
		detailsStr = fmt.Sprintf("%v", v)
	}

	m.SetDetails(p.Sanitize(detailsStr))

	return nil
}

const defaultImportTimeout = time.Second * 10

var client = &http.Client{
	Timeout: defaultImportTimeout,
}

// importURLToSchema is a helper that fetches content from a URL, detects its MIME type, parses it and writes
// the sanitized content into the mutation details, recording the URL used
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

	// Read a bounded amount of data based on configured import size to prevent memory overuse
	reader := io.LimitReader(resp.Body, int64(m.Client().EntConfig.MaxSchemaImportSize))

	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err) // nolint:err113
	}

	// Detect MIME using storage helper with fallback to header to handle servers with incorrect content type
	mimeType := resp.Header.Get("Content-Type")
	if detected, derr := storage.DetectContentType(bytes.NewReader(buf)); derr == nil && detected != "" {
		mimeType = detected
	} else {
		mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	}

	// Parse the document using the detected MIME type
	parsed, err := storage.ParseDocument(bytes.NewReader(buf), mimeType)
	if err != nil {
		return fmt.Errorf("failed to parse document: %w", err)
	}

	// Convert structured results into a string representation for details
	var detailsStr string
	switch v := parsed.(type) {
	case string:
		detailsStr = v
	default:
		detailsStr = fmt.Sprintf("%v", v)
	}

	p := bluemonday.UGCPolicy()
	m.SetURL(downloadURL)
	m.SetDetails(p.Sanitize(detailsStr))

	return nil
}

func filenameToTitle(filename string) string {
	// remove file extension if present
	filename = strings.TrimSpace(filename)
	if dotIdx := strings.LastIndex(filename, "."); dotIdx != -1 {
		filename = filename[:dotIdx]
	}

	// replace underscores and hyphens with spaces
	filename = strings.ReplaceAll(filename, "_", " ")
	filename = strings.ReplaceAll(filename, "-", " ")

	// capitalize first letter of each word
	return caser.String(filename)
}
