package gcpscc

import "time"

// Config holds operator-level workload identity configuration for the GCP Security Command Center definition
type Config struct {
	// Audience is the STS audience for the workload identity pool
	Audience string `json:"audience" koanf:"audience"`
	// ServiceAccount is the GCP service account email to impersonate
	ServiceAccount string `json:"serviceaccount" koanf:"serviceaccount"`
	// SubjectTokenType overrides the default subject token type for STS exchanges
	SubjectTokenType string `json:"subjecttokentype" koanf:"subjecttokentype"`
	// Scopes lists the GCP OAuth scopes to request on impersonated access tokens
	Scopes []string `json:"scopes" koanf:"scopes"`
	// TokenLifetime configures the default lifetime for impersonated access tokens
	TokenLifetime time.Duration `json:"tokenlifetime" koanf:"tokenlifetime"`
}
