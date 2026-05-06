package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/workflowdefinition"
	"github.com/theopenlane/core/pkg/logx"
)

const (
	defaultWorkflowSeedFetchTimeout = 10 * time.Second
	maxWorkflowSeedBytes            = int64(1 << 20) // 1 MiB
)

var (
	// ErrWorkflowSeedFailed is returned when workflow definition upsert fails.
	ErrWorkflowSeedFailed = errors.New("workflow definition seed failed")
	// ErrWorkflowSeedSourceFetchFailed is returned when a remote seed source cannot be fetched.
	ErrWorkflowSeedSourceFetchFailed = errors.New("workflow definition seed source fetch failed")
	// ErrWorkflowSeedSourceTooLarge is returned when a remote seed payload exceeds the max size.
	ErrWorkflowSeedSourceTooLarge = errors.New("workflow definition seed source too large")
	// ErrWorkflowSeedManifestInvalid is returned when seed payload is malformed.
	ErrWorkflowSeedManifestInvalid = errors.New("workflow definition seed manifest invalid")
	// ErrWorkflowSeedDefinitionKeyRequired is returned when definition key is missing.
	ErrWorkflowSeedDefinitionKeyRequired = errors.New("workflow definition seed key is required")
	// ErrWorkflowSeedSchemaTypeRequired is returned when schema type cannot be resolved.
	ErrWorkflowSeedSchemaTypeRequired = errors.New("workflow definition seed schema type is required")
	// ErrWorkflowSeedWorkflowKindRequired is returned when workflow kind cannot be resolved.
	ErrWorkflowSeedWorkflowKindRequired = errors.New("workflow definition seed workflow kind is required")
	// ErrWorkflowSeedURLInvalid is returned when the seed URL is invalid.
	ErrWorkflowSeedURLInvalid = errors.New("workflow definition seed URL is invalid")
)

// DefinitionSeedManifest defines a payload for seeding workflow definitions.
type DefinitionSeedManifest struct {
	Definitions []SeedWorkflowDefinition `json:"definitions"`
}

// SeedWorkflowDefinition defines a single workflow definition seed record.
type SeedWorkflowDefinition struct {
	// Key is a stable identifier persisted to system_internal_id.
	Key string `json:"key"`
	// Name is the workflow definition display name.
	Name string `json:"name"`
	// Description is optional definition description text.
	Description string `json:"description,omitempty"`
	// WorkflowKind optionally overrides the workflow kind derived from Definition.
	WorkflowKind enums.WorkflowKind `json:"workflow_kind,omitempty"`
	// SchemaType optionally overrides the schema type derived from Definition.
	SchemaType string `json:"schema_type,omitempty"`
	// Revision overrides workflow revision. Defaults to 1.
	Revision int `json:"revision,omitempty"`
	// Draft controls whether the definition is a draft. Defaults to false.
	Draft *bool `json:"draft,omitempty"`
	// Active controls whether the definition is active. Defaults to true.
	Active *bool `json:"active,omitempty"`
	// IsDefault controls whether this definition is default for schema type. Defaults to false.
	IsDefault *bool `json:"is_default,omitempty"`
	// CooldownSeconds configures duplicate trigger suppression.
	CooldownSeconds int `json:"cooldown_seconds,omitempty"`
	// ApprovalSubmissionMode optionally overrides the definition-level submission mode.
	ApprovalSubmissionMode enums.WorkflowApprovalSubmissionMode `json:"approval_submission_mode,omitempty"`
	// Definition is the workflow definition document to persist.
	Definition models.WorkflowDefinitionDocument `json:"definition"`
}

// DefinitionSeeder upserts system-owned workflow definitions from manifest inputs.
type DefinitionSeeder struct {
	client     *generated.Client
	httpClient *http.Client
}

// NewDefinitionSeeder creates a workflow definition seeder.
func NewDefinitionSeeder(client *generated.Client) *DefinitionSeeder {
	return &DefinitionSeeder{
		client:     client,
		httpClient: &http.Client{Timeout: defaultWorkflowSeedFetchTimeout},
	}
}

// SeedDefinitionsFromManifestFile loads and upserts definitions from a local manifest file.
func (s *DefinitionSeeder) SeedDefinitionsFromManifestFile(ctx context.Context, filePath string) error {
	payload, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("manifest_path", filePath).Msg("failed reading workflow definition seed manifest file")
		return ErrWorkflowSeedSourceFetchFailed
	}

	manifest := DefinitionSeedManifest{}
	if err := json.Unmarshal(payload, &manifest); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("manifest_path", filePath).Msg("failed unmarshalling workflow definition seed manifest file")
		return ErrWorkflowSeedManifestInvalid
	}

	return s.SeedDefinitionsFromManifest(ctx, manifest)
}

// SeedDefinitionsFromManifestURL loads and upserts definitions from a remote manifest URL.
func (s *DefinitionSeeder) SeedDefinitionsFromManifestURL(ctx context.Context, manifestURL string) error {
	if err := validateWorkflowSeedURL(manifestURL); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("manifest_url", manifestURL).Msg("invalid workflow definition seed manifest URL")
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("manifest_url", manifestURL).Msg("failed building workflow definition seed manifest request")
		return ErrWorkflowSeedSourceFetchFailed
	}

	response, err := s.httpClient.Do(req)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("manifest_url", manifestURL).Msg("failed fetching workflow definition seed manifest")
		return ErrWorkflowSeedSourceFetchFailed
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		logx.FromContext(ctx).Error().Int("status", response.StatusCode).Str("manifest_url", manifestURL).Msg("workflow definition seed manifest returned non-success status")
		return ErrWorkflowSeedSourceFetchFailed
	}

	payload, err := io.ReadAll(io.LimitReader(response.Body, maxWorkflowSeedBytes+1))
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("manifest_url", manifestURL).Msg("failed reading workflow definition seed manifest response")
		return ErrWorkflowSeedSourceFetchFailed
	}
	if int64(len(payload)) > maxWorkflowSeedBytes {
		logx.FromContext(ctx).Error().Int64("max_bytes", maxWorkflowSeedBytes).Str("manifest_url", manifestURL).Msg("workflow definition seed manifest exceeds max size")
		return ErrWorkflowSeedSourceTooLarge
	}

	manifest := DefinitionSeedManifest{}
	if err := json.Unmarshal(payload, &manifest); err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("manifest_url", manifestURL).Msg("failed unmarshalling workflow definition seed manifest")
		return ErrWorkflowSeedManifestInvalid
	}

	return s.SeedDefinitionsFromManifest(ctx, manifest)
}

// SeedDefinitionsFromManifest upserts system-owned definitions from an in-memory manifest payload.
func (s *DefinitionSeeder) SeedDefinitionsFromManifest(ctx context.Context, manifest DefinitionSeedManifest) error {
	allowCtx := AllowContext(ctx)

	for _, item := range manifest.Definitions {
		if _, err := s.upsertWorkflowDefinition(allowCtx, item); err != nil {
			return err
		}
	}

	return nil
}

// upsertWorkflowDefinition creates or updates a system-owned workflow definition seed record.
func (s *DefinitionSeeder) upsertWorkflowDefinition(ctx context.Context, input SeedWorkflowDefinition) (*generated.WorkflowDefinition, error) {
	key := strings.TrimSpace(input.Key)
	if key == "" {
		return nil, ErrWorkflowSeedDefinitionKeyRequired
	}

	definition := input.Definition
	schemaType := resolveSeedSchemaType(input.SchemaType, definition)
	if schemaType == "" {
		return nil, ErrWorkflowSeedSchemaTypeRequired
	}

	workflowKind := resolveSeedWorkflowKind(input.WorkflowKind, definition)
	if workflowKind == enums.WorkflowKind("") {
		return nil, ErrWorkflowSeedWorkflowKindRequired
	}

	definition.SchemaType = schemaType
	if definition.WorkflowKind == "" {
		definition.WorkflowKind = workflowKind
	}

	approvalSubmissionMode := input.ApprovalSubmissionMode
	if approvalSubmissionMode == "" {
		approvalSubmissionMode = definition.ApprovalSubmissionMode
	}
	if approvalSubmissionMode == "" {
		approvalSubmissionMode = enums.WorkflowApprovalSubmissionModeAutoSubmit
	}
	definition.ApprovalSubmissionMode = approvalSubmissionMode

	revision := input.Revision
	if revision <= 0 {
		revision = 1
	}

	draft := boolOrDefault(input.Draft, false)
	active := boolOrDefault(input.Active, true)
	isDefault := boolOrDefault(input.IsDefault, false)
	operations, fields := DeriveTriggerPrefilter(definition)

	records, err := s.client.WorkflowDefinition.Query().
		Where(workflowdefinition.SystemInternalIDEQ(key)).
		All(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("seed_key", key).Msg("failed querying workflow definition seed")
		return nil, ErrWorkflowSeedFailed
	}

	var record *generated.WorkflowDefinition
	for _, candidate := range records {
		if candidate.SystemOwned {
			record = candidate
			break
		}
	}
	if record == nil && len(records) > 0 {
		record = records[0]
	}

	if record == nil {
		createBuilder := s.client.WorkflowDefinition.Create().
			SetSystemOwned(true).
			SetSystemInternalID(key).
			SetName(input.Name).
			SetWorkflowKind(workflowKind).
			SetSchemaType(schemaType).
			SetRevision(revision).
			SetDraft(draft).
			SetCooldownSeconds(input.CooldownSeconds).
			SetIsDefault(isDefault).
			SetActive(active).
			SetTriggerOperations(operations).
			SetTriggerFields(fields).
			SetApprovalSubmissionMode(approvalSubmissionMode).
			SetDefinitionJSON(definition)

		if strings.TrimSpace(input.Description) != "" {
			createBuilder = createBuilder.SetDescription(input.Description)
		}

		if !draft {
			createBuilder = createBuilder.SetPublishedAt(time.Now())
		}

		created, createErr := createBuilder.Save(ctx)
		if createErr != nil {
			logx.FromContext(ctx).Error().Err(createErr).Str("seed_key", key).Msg("failed creating workflow definition seed")
			return nil, ErrWorkflowSeedFailed
		}

		return created, nil
	}

	updateBuilder := record.Update().
		SetName(input.Name).
		SetWorkflowKind(workflowKind).
		SetSchemaType(schemaType).
		SetRevision(revision).
		SetDraft(draft).
		SetCooldownSeconds(input.CooldownSeconds).
		SetIsDefault(isDefault).
		SetActive(active).
		SetTriggerOperations(operations).
		SetTriggerFields(fields).
		SetApprovalSubmissionMode(approvalSubmissionMode).
		SetDefinitionJSON(definition)

	if strings.TrimSpace(input.Description) != "" {
		updateBuilder = updateBuilder.SetDescription(input.Description)
	} else {
		updateBuilder = updateBuilder.ClearDescription()
	}

	if draft {
		updateBuilder = updateBuilder.ClearPublishedAt()
	} else if record.PublishedAt == nil {
		updateBuilder = updateBuilder.SetPublishedAt(time.Now())
	}

	updated, updateErr := updateBuilder.Save(ctx)
	if updateErr != nil {
		logx.FromContext(ctx).Error().Err(updateErr).Str("seed_key", key).Msg("failed updating workflow definition seed")
		return nil, ErrWorkflowSeedFailed
	}

	return updated, nil
}

// resolveSeedSchemaType resolves schema type from explicit input, definition schemaType, or first trigger.
func resolveSeedSchemaType(explicit string, definition models.WorkflowDefinitionDocument) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	if strings.TrimSpace(definition.SchemaType) != "" {
		return strings.TrimSpace(definition.SchemaType)
	}
	if len(definition.Triggers) > 0 {
		return definition.Triggers[0].ObjectType.String()
	}

	return ""
}

// resolveSeedWorkflowKind resolves workflow kind from explicit input or document.
func resolveSeedWorkflowKind(explicit enums.WorkflowKind, definition models.WorkflowDefinitionDocument) enums.WorkflowKind {
	if explicit != "" {
		return explicit
	}
	if definition.WorkflowKind != "" {
		return definition.WorkflowKind
	}

	if DefinitionHasApprovalAction(definition) {
		return enums.WorkflowKindApproval
	}
	if DefinitionHasReviewAction(definition) {
		return enums.WorkflowKindLifecycle
	}

	return enums.WorkflowKindNotification
}

// boolOrDefault returns pointer value if provided, otherwise fallback.
func boolOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}

	return *value
}

// validateWorkflowSeedURL enforces http/https manifest URLs with hosts.
func validateWorkflowSeedURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil {
		return ErrWorkflowSeedURLInvalid
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrWorkflowSeedURLInvalid
	}

	if parsed.Host == "" {
		return ErrWorkflowSeedURLInvalid
	}

	return nil
}
