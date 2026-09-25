package gemini

import (
	"encoding/json"
	"time"

	"google.golang.org/genai"

	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/docextract"
	"github.com/theopenlane/core/v2/pkg/docextract/soc2"
	"github.com/theopenlane/core/v2/pkg/modelarmor"
)

var (
	// DefinitionID is the stable identifier for the Gemini integration definition
	DefinitionID = types.NewDefinitionRef("def_01K0GEMINI00000000000000001")
	// geminiClient is the client ref for the Gemini client used by this definition
	geminiClient = types.NewClientRef[*Client]()
	// reportScanRequestSchema is the operation ref for kicking off the parse of a pending report scan
	reportScanRequestSchema, ReportScanRequestOp = providerkit.OperationSchema[ReportScanRequest]() //nolint:revive
	// reportScanPartSchema is the operation ref for parsing one section of a report
	reportScanPartSchema, ReportScanPartOp = providerkit.OperationSchema[ReportScanPartRequest]() //nolint:revive
	// reportScanReleaseSchema is the operation ref for releasing the cached report once a scan finalizes
	reportScanReleaseSchema, ReportScanReleaseOp = providerkit.OperationSchema[ReportScanReleaseRequest]() //nolint:revive
	// runtimeGeminiSchema is the JSON schema and typed ref for the runtime Gemini config
	runtimeGeminiSchema, runtimeGeminiRef = providerkit.RuntimeSchema[RuntimeConfig]()
)

const (
	// ReportMetadataKey is the Scan.Metadata key holding the parsed report sections
	ReportMetadataKey = "report"
	// ReportTypeMetadataKey is the Scan.Metadata key naming which document kind parsed the report
	ReportTypeMetadataKey = "reportType"
	// ErrorMetadataKey is the Scan.Metadata key holding the last parse failure
	ErrorMetadataKey = "error"
	// SummaryMetadataKey is the Scan.Metadata key holding each section's state, item count, error, and
	// batch progress, seeded at submit and updated as parts land
	SummaryMetadataKey = "report_summary"
	// ValidationMetadataKey is the Scan.Metadata key holding the SOC 2 validation outcome for the upload
	ValidationMetadataKey = "report_validation"
	// RequestedPartsMetadataKey is the Scan.Metadata key naming the sections to parse; empty means every configured section
	RequestedPartsMetadataKey = "parts"
	// DocumentCacheTTL is how long the cached report outlives the scan submit; the cache is released when the scan finalizes
	DocumentCacheTTL = 3 * time.Hour

	// PerformedBy marks a Scan record as one the system should parse as an uploaded report
	PerformedBy = "openlane_report_parse"
	// FileUploadKey is the multipart key report PDFs are uploaded under on scan mutations
	FileUploadKey = "scanFiles"
	// SubmitInterval is the minimum time between report scans for one organization
	SubmitInterval = 24 * time.Hour

	// PartStatePending marks a section whose parse job has not finished
	PartStatePending = "pending"
	// PartStateCompleted marks a section whose parse job stored its result
	PartStateCompleted = "completed"
	// PartStateFailed marks a section whose parse job exhausted its attempts
	PartStateFailed = "failed"
)

// Client is the Gemini client for the runtime path: the generic extraction client plus the
// SOC 2 document kind built from the configured prompts
type Client struct {
	*docextract.Client

	// SOC2 is the SOC 2 report kind, ready to pass to Extract
	SOC2 soc2.Kind
	// Screener checks uploaded reports with Model Armor before extraction, nil when not configured
	Screener *modelarmor.Client
}

// Prompts is the operator-provided prompt text, kept in config rather than code
type Prompts struct {
	// SystemInstruction is the system prompt applied to every extraction request
	SystemInstruction string `json:"systemInstruction" koanf:"systeminstruction" jsonschema:"description=System instruction applied to every extraction request"`
	// SOC2 holds the prompts used to extract from a SOC 2 report
	SOC2 soc2.Prompts `json:"soc2" koanf:"soc2" jsonschema:"description=Prompts used to extract from a SOC 2 report"`
}

// RuntimeConfig is the runtime-provisioned configuration for the operator-owned Gemini account.
// Sourced from koanf/environment at startup; used for system-initiated Gemini calls such as
// report scans that are not tied to a customer installation
type RuntimeConfig struct {
	// Backend selects the API the client talks to: gemini for the Gemini Developer API or vertex for Vertex AI
	Backend Backend `json:"backend" koanf:"backend" jsonschema:"description=Gemini backend to use: gemini for the Gemini Developer API or vertex for Vertex AI,enum=gemini,enum=vertex,default=gemini"`
	// APIKey is the Gemini API key for the operator-owned account; on Vertex AI it enables express mode instead of project credentials
	APIKey string `json:"apikey" koanf:"apikey" jsonschema:"description=Gemini API key for the operator-owned account" sensitive:"true"`
	// Project is the Google Cloud project used by the Vertex AI backend
	Project string `json:"project" koanf:"project" jsonschema:"description=Google Cloud project for the Vertex AI backend"`
	// Location is the Google Cloud region used by the Vertex AI backend
	Location string `json:"location" koanf:"location" jsonschema:"description=Google Cloud region for the Vertex AI backend, for example us-central1"`
	// ModelArmorTemplate is the full resource name of the Model Armor template uploaded reports are screened against; empty disables screening
	ModelArmorTemplate string `json:"modelarmortemplate" koanf:"modelarmortemplate" jsonschema:"description=Model Armor template resource name used to screen uploaded reports for prompt injection, empty disables screening"`
	// Model is the Gemini model used for extraction, defaults to the client's default when empty
	Model string `json:"model" koanf:"model" jsonschema:"description=Gemini model used for extraction"`
	// Prompts holds the prompt text sent to the model
	Prompts Prompts `json:"prompts" koanf:"prompts" jsonschema:"description=Prompt text used for extraction"`
}

// Backend names the API the Gemini client talks to
type Backend string

const (
	// BackendGemini is the Gemini Developer API, authenticated with an api key
	BackendGemini Backend = "gemini"
	// BackendVertex is Vertex AI, authenticated with Google Cloud credentials, with the ability to use model armour
	BackendVertex Backend = "vertex"
)

// Provisioned reports whether the runtime config has the minimum required fields to make Gemini calls
func (c RuntimeConfig) Provisioned() bool {
	if c.Backend == BackendVertex {
		return c.Project != "" && c.Location != ""
	}

	return c.APIKey != ""
}

// GenAIBackend maps the configured backend name to the genai client backend
func (c RuntimeConfig) GenAIBackend() genai.Backend {
	if c.Backend == BackendVertex {
		return genai.BackendVertexAI
	}

	return genai.BackendGeminiAPI
}

// ReportScanRequest schedules the parse of the pdf attached to a pending report-type Scan record
type ReportScanRequest struct {
	// ScanID identifies the Scan record whose attached report should be parsed
	ScanID string `json:"scanId" jsonschema:"required,title=Scan ID,description=Scan record holding the uploaded report"`
	// OrganizationID is the organization the scan belongs to
	OrganizationID string `json:"organizationId,omitempty"`

	// parts lists the sections with a configured prompt, set by the builder from the runtime config
	parts []string
}

// ReportScanRequestResult acknowledges what happened to a report scan request
type ReportScanRequestResult struct {
	// Message describes what happened
	Message string `json:"message"`
	// ScanID is the id of the Scan record for this request
	ScanID string `json:"scanId"`
}

// ReportScanPartRequest parses one section of the pdf attached to a report scan
type ReportScanPartRequest struct {
	// ScanID identifies the Scan record whose attached report should be parsed
	ScanID string `json:"scanId" jsonschema:"required,title=Scan ID,description=Scan record holding the uploaded report"`
	// OrganizationID is the organization the scan belongs to
	OrganizationID string `json:"organizationId" jsonschema:"required,title=Organization ID"`
	// Part is the section name to parse, e.g. controls or reviews
	Part string `json:"part" jsonschema:"required,title=Part,description=Report section to parse"`
	// RefCodes optionally scopes the parse to the controls with these ref codes
	RefCodes []string `json:"refCodes,omitempty" jsonschema:"title=Ref Codes,description=Control ref codes to scope the section to"`
	// Cache names the cached content holding the report, carried on the job rather than the scan
	// record because the resource name identifies the operator's cloud project
	Cache string `json:"cache,omitempty" jsonschema:"title=Cache,description=Cached content holding the report for this parse"`
}

// ReportScanReleaseRequest releases the cached report once a scan has finalized
type ReportScanReleaseRequest struct {
	// ScanID identifies the Scan record whose cached report should be released
	ScanID string `json:"scanId" jsonschema:"required,title=Scan ID,description=Scan record whose cached report should be released"`
	// OrganizationID is the organization the scan belongs to
	OrganizationID string `json:"organizationId" jsonschema:"required,title=Organization ID"`
	// Cache names the cached content to delete
	Cache string `json:"cache" jsonschema:"required,title=Cache,description=Cached content to release"`
}

// ReportScanReleaseResult acknowledges the cache release
type ReportScanReleaseResult struct {
	// Released is true when a cache existed and was deleted
	Released bool `json:"released"`
}

// ReportScanPartResult carries the sections one part job extracted
type ReportScanPartResult struct {
	// Sections holds the parsed json keyed by section name
	Sections map[string]json.RawMessage `json:"sections"`
}
