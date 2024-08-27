package tokens

import "time"

// Config defines the configuration settings for authentication tokens used in the server
type Config struct {
	// KID represents the Key ID used in the configuration.
	KID string `json:"kid" koanf:"kid" jsonschema:"required"`
	// Audience represents the target audience for the tokens.
	Audience string `json:"audience" koanf:"audience" jsonschema:"required" default:"https://theopenlane.io"`
	// RefreshAudience represents the audience for refreshing tokens.
	RefreshAudience string `json:"refreshAudience" koanf:"refreshAudience"`
	// Issuer represents the issuer of the tokens
	Issuer string `json:"issuer" koanf:"issuer" jsonschema:"required" default:"https://auth.theopenlane.io" `
	// AccessDuration represents the duration of the access token is valid for
	AccessDuration time.Duration `json:"accessDuration" koanf:"accessDuration" default:"1h"`
	// RefreshDuration represents the duration of the refresh token is valid for
	RefreshDuration time.Duration `json:"refreshDuration" koanf:"refreshDuration" default:"2h"`
	// RefreshOverlap represents the overlap time for a refresh and access token
	RefreshOverlap time.Duration `json:"refreshOverlap" koanf:"refreshOverlap" default:"-15m" `
	// JWKSEndpoint represents the endpoint for the JSON Web Key Set
	JWKSEndpoint string `json:"jwksEndpoint" koanf:"jwksEndpoint" default:"https://api.theopenlane.io/.well-known/jwks.json"`
	// Keys represents the key pairs used for signing the tokens
	Keys map[string]string `json:"keys" koanf:"keys" jsonschema:"required"`
	// GenerateKeys is a boolean to determine if the keys should be generated
	GenerateKeys bool `json:"generateKeys" koanf:"generateKeys" default:"true"`
}
