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
	// MaxPoolSize is the max worker pool size that can be used by the ent client
	MaxPoolSize int `json:"maxpoolsize" koanf:"maxpoolsize" default:"200"`
	// Modules contains the configuration for the module system
	Modules Modules `json:"modules" koanf:"modules"`
	// MaxSchemaImportSize is the maximum size allowed for schema imports in bytes
	MaxSchemaImportSize int `json:"maxschemaimportsize" koanf:"maxschemaimportsize" default:"262144" description:"maximum size allowed for schema imports (256KB)"`
	// EmailValidation contains configuration for email validation
	EmailValidation validator.EmailVerificationConfig `json:"emailvalidation" koanf:"emailvalidation"`
	// Billing contains configuration for billing related features
	Billing Billing `json:"billing" koanf:"billing"`
	// Notifications contains configuration for notifications sent to users based on events
	Notifications Notifications `json:"notifications" koanf:"notifications"`
}

// Modules settings for features access
type Modules struct {
	// Enabled indicates whether to check and verify module access
	Enabled bool `json:"enabled" koanf:"enabled" default:"true"`
	// UseSandbox indicates whether to use the sandbox catalog for module access checks
	UseSandbox bool `json:"usesandbox" koanf:"usesandbox" default:"false"`
	// DevMode enables all modules for local development regardless of trial status
	DevMode bool `json:"devmode" koanf:"devmode" default:"false"`
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

// Notifications settings for notifications sent to users based on events
type Notifications struct {
	// ConsoleURL for ui links used in notifications
	ConsoleURL string `koanf:"consoleurl" json:"consoleurl" default:"http://localhost:3001" domain:"inherit" domainPrefix:"https://console"`
}
