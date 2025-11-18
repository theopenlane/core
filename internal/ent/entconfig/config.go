package entconfig

import (
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/summarizer"
)

// Config holds the configuration for the ent server
type Config struct {
	// EntityTypes is the list of entity types to create by default for the organization
	EntityTypes []string `json:"entitytypes" koanf:"entitytypes" default:"" description:"entity types to create for the organization"`
	// Summarizer contains configuration for text summarization
	Summarizer summarizer.Config `json:"summarizer" koanf:"summarizer"`
	// Windmill contains configuration for Windmill workflow automation
	Windmill Windmill `json:"windmill" koanf:"windmill"`
	// MaxPoolSize is the max pond pool workers that can be used by the ent client
	MaxPoolSize int `json:"maxpoolsize" koanf:"maxpoolsize" default:"100"`
	// Modules contains the configuration for the module system
	Modules Modules `json:"modules" koanf:"modules"`
	// MaxSchemaImportSize is the maximum size allowed for schema imports in bytes
	MaxSchemaImportSize int `json:"maxschemaimportsize" koanf:"maxschemaimportsize" default:"262144" description:"maximum size allowed for schema imports (256KB)"`
	// EmailValidation contains configuration for email validation
	EmailValidation validator.EmailVerificationConfig `json:"emailvalidation" koanf:"emailvalidation"`
	// Billing contains configuration for billing related features
	Billing Billing `json:"billing" koanf:"billing"`
}

// Windmill holds configuration for the Windmill workflow automation platform
type Windmill struct {
	// Enabled specifies whether Windmill integration is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false"`

	// BaseURL is the base URL of the Windmill instance
	BaseURL string `json:"baseurl" koanf:"baseurl" default:"https://app.windmill.dev"`

	// Workspace is the Windmill workspace to use
	Workspace string `json:"workspace" koanf:"workspace"`

	// Token is the API token for authentication with Windmill
	Token string `json:"token" koanf:"token" sensitive:"true"`

	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout string `json:"defaulttimeout" koanf:"defaulttimeout" default:"30s"`

	// Timezone for scheduled jobs
	Timezone string `json:"timezone" koanf:"timezone" default:"UTC"`

	// OnFailureScript script to run when a scheduled job fails
	OnFailureScript string `json:"onfailurescript" koanf:"onfailurescript"`

	// OnSuccessScript script to run when a scheduled job succeeds
	OnSuccessScript string `json:"onsuccessscript" koanf:"onsuccessscript"`
}

// Modules settings for features access
type Modules struct {
	// Enabled indicates whether to check and verify module access
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// UseSandbox indicates whether to use the sandbox catalog for module access checks
	UseSandbox bool `json:"usesandbox" koanf:"usesandbox" default:"false"`
}

// Billing settings for feature access
type Billing struct {
	// RequirePaymentMethod indicates whether to check if a payment method
	// exists for orgs before they can access some resource
	RequirePaymentMethod bool `json:"requirepaymentmethod" koanf:"requirepaymentmethod"`
	// BypassEmailDomains is a list of domains that should be allowed to bypass
	// the checks if RequirePaymentMethod above is enabled
	BypassEmailDomains []string `json:"bypassemaildomains" koanf:"bypassemaildomains"`
}
