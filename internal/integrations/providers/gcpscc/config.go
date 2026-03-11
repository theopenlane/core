package gcpscc

import "time"

// Config holds operator-level configuration for the GCP Security Command Center provider.
type Config struct {
	// Audience is the workload identity pool/provider audience for STS token exchanges.
	Audience string `json:"audience" koanf:"audience"`
	// ServiceAccount is the GCP service account to impersonate.
	ServiceAccount string `json:"serviceaccount" koanf:"serviceaccount"`
	// SubjectTokenType overrides the default subject token type for STS exchanges.
	SubjectTokenType string `json:"subjecttokentype" koanf:"subjecttokentype"`
	// Scopes lists the GCP OAuth scopes to request on impersonated access tokens.
	Scopes []string `json:"scopes" koanf:"scopes"`
	// TokenLifetime configures the default lifetime for impersonated access tokens.
	TokenLifetime time.Duration `json:"tokenlifetime" koanf:"tokenlifetime"`
}
