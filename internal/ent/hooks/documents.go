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
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/objects/storage"
	"github.com/theopenlane/iam/auth"
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
	SetStatus(enums.DocumentStatus)
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
				logx.FromContext(ctx).Info().Msg("import document hook used on unsupported mutation type")

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
				isUpdate := mut.Op() != ent.OpCreate
				if err := importFileToSchema(ctx, mut, isUpdate); err != nil {
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
// strips the front matter if present from the details and sets frontmatter fields, such as title, on the mutation as well
// if updateOnly is true, it will only update the details if a file is uploaded, and no other fields are modified
func importFileToSchema[T importSchemaMutation](ctx context.Context, m T, updateOnly bool) error {
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

	// use a default name for the document on created
	if !updateOnly {
		m.SetName(filenameToTitle(file[0].OriginalName))
	}

	// if frontmatter is present and has a title, use it as the document name
	if parsedContent.Frontmatter != nil {
		// If frontmatter is present, see if title is set and use it as the document name
		if parsedContent.Frontmatter.Title != "" {
			m.SetName(parsedContent.Frontmatter.Title)
		}

		if parsedContent.Frontmatter.Status != "" {
			status := enums.ToDocumentStatus(parsedContent.Frontmatter.Status)

			m.SetStatus(*status)
		}
	}

	m.SetFileID(file[0].ID)

	var detailsStr string
	switch v := parsedContent.Data.(type) {
	case string:
		detailsStr = v
	case []byte:
		detailsStr = string(v)
	default:
		detailsStr = fmt.Sprintf("%v", v)
	}

	details := p.Sanitize(detailsStr)

	orgName := ""
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err == nil {
		org, err := m.Client().Organization.Get(ctx, orgID)
		if err != nil {
			return err
		}

		orgName = org.Name
	}

	details = updatePlaceholderText(details, orgName)

	m.SetDetails(p.Sanitize(details))

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
	switch v := parsed.Data.(type) {
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

// statusMutation is an interface for mutations that have status and approver/delegate fields
type statusMutation interface {
	utils.GenericMutation

	Status() (r enums.DocumentStatus, exists bool)
	OldApproverID(ctx context.Context) (v string, err error)
	OldDelegateID(ctx context.Context) (v string, err error)
	ApproverID() (id string, exists bool)
	DelegateID() (id string, exists bool)
	ApprovalRequired() (v bool, exists bool)
	OldApprovalRequired(ctx context.Context) (v bool, err error)
}

// getApproverDelegateIDs retrieves the approver and delegate group IDs based on the operation type
func getApproverDelegateIDs(ctx context.Context, mut statusMutation) (approverID, delegateID string) {
	switch mut.Op() {
	case ent.OpCreate:
		// For create operations, get the IDs from the mutation
		approverID, _ = mut.ApproverID()
		delegateID, _ = mut.DelegateID()
	case ent.OpUpdate, ent.OpUpdateOne:
		// For update operations, get the old values from the database
		approverID, _ = mut.OldApproverID(ctx)
		delegateID, _ = mut.OldDelegateID(ctx)
	}
	return approverID, delegateID
}

// getRequireApproval determines if approval is required based on the operation type
func getRequireApproval(ctx context.Context, mut statusMutation) (bool, error) {
	switch mut.Op() {
	case ent.OpCreate:
		if requireApproval, exists := mut.ApprovalRequired(); exists {
			return requireApproval, nil
		}

		// this field will default to true if not set on create
		return true, nil
	case ent.OpUpdate, ent.OpUpdateOne:
		// check if the field is being updated to not require approval anymore first
		requireApproval, exists := mut.ApprovalRequired()
		if exists {
			return requireApproval, nil
		}

		// otherwise fall back to the old value from the database
		return mut.OldApprovalRequired(ctx)
	default:
		return true, nil
	}
}

// checkUserInApproverGroups verifies if a user is a member of the approver or delegate group
func checkUserInApproverGroups(ctx context.Context, client *generated.Client, userID, approverID, delegateID string) (bool, error) {
	query := client.GroupMembership.Query().Where(groupmembership.UserID(userID))

	// Build the query to check membership in either group
	switch {
	case approverID != "" && delegateID != "":
		query = query.Where(
			groupmembership.Or(
				groupmembership.GroupID(approverID),
				groupmembership.GroupID(delegateID),
			),
		)
	case approverID != "":
		query = query.Where(groupmembership.GroupID(approverID))
	default:
		query = query.Where(groupmembership.GroupID(delegateID))
	}

	return query.Exist(ctx)
}

// HookStatusApproval is an ent hook that ensures only users in the approver or delegate group can set status to APPROVED
func HookStatusApproval() ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(statusMutation)
			if !ok {
				logx.FromContext(ctx).Info().Msg("status approval hook used on unsupported mutation type")
				return next.Mutate(ctx, m)
			}

			// Check if status is being set to APPROVED
			status, exists := mut.Status()
			if !exists || status != enums.DocumentApproved {
				// Not setting to APPROVED, allow the mutation
				return next.Mutate(ctx, m)
			}

			// check if the document requires approval
			requireApproval, err := getRequireApproval(ctx, mut)
			if err == nil && !requireApproval {
				// no approval required, allow the mutation
				return next.Mutate(ctx, m)
			}

			// Get the authenticated user
			actor, err := auth.GetAuthenticatedUserFromContext(ctx)
			if err != nil {
				return nil, err
			}

			// Determine the approver and delegate group IDs based on operation type
			approverID, delegateID := getApproverDelegateIDs(ctx, mut)

			// If neither approver nor delegate group is set, reject the approval
			if approverID == "" && delegateID == "" {
				return nil, ErrStatusApprovedNotAllowed
			}

			// Check if the user is a member of either the approver or delegate group
			isMember, err := checkUserInApproverGroups(ctx, mut.Client(), actor.SubjectID, approverID, delegateID)
			if err != nil {
				return nil, err
			}

			if !isMember {
				return nil, ErrStatusApprovedNotAllowed
			}

			// User is authorized, proceed with the mutation
			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
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

const (
	companyPlaceholder = "{{company_name}}"
)

// updatePlaceholderText replaces the company placeholder in details with the provided organization name
func updatePlaceholderText(details string, orgName string) string {
	if orgName == "" {
		log.Warn().Msg("organization name is empty, using default placeholder value")

		orgName = "[Company Name]"
	}

	return strings.ReplaceAll(details, companyPlaceholder, orgName)
}
