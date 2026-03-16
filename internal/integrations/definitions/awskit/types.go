package awskit

import "time"

const (
	// AccountScopeAll indicates operations should run across all accessible accounts
	AccountScopeAll = "all"
	// AccountScopeSpecific indicates operations should be limited to explicitly listed accounts
	AccountScopeSpecific = "specific"
)

// ProviderData holds the raw AWS provider data persisted in one credential's ProviderData field
// Definition-specific credential schemas may expose only a subset of these fields, but any fields
// they persist must continue to match these JSON tags so shared metadata decoding remains stable
type ProviderData struct {
	// Region is the AWS region for API calls
	Region string `json:"region"`
	// HomeRegion is the Security Hub home region for aggregated queries
	HomeRegion string `json:"homeRegion"`
	// LinkedRegions optionally limits queries to the listed regions
	LinkedRegions []string `json:"linkedRegions"`
	// OrganizationID is the AWS Organizations identifier associated with this integration
	OrganizationID string `json:"organizationId"`
	// AccountScope controls whether queries should use all accounts or a provided subset
	AccountScope string `json:"accountScope"`
	// AccountIDs optionally scopes collection to specific AWS account IDs
	AccountIDs []string `json:"accountIds"`
	// RoleARN is the ARN of the role to assume
	RoleARN string `json:"roleArn"`
	// AccountID is the AWS account ID
	AccountID string `json:"accountId"`
	// ExternalID is the external ID for role assumption
	ExternalID string `json:"externalId"`
	// SessionName is the name for the session
	SessionName string `json:"sessionName"`
	// SessionDuration is the duration for the session
	SessionDuration string `json:"sessionDuration"`
	// AccessKeyID is the AWS access key ID
	AccessKeyID string `json:"accessKeyId"`
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string `json:"secretAccessKey"`
	// SessionToken is the AWS session token
	SessionToken string `json:"sessionToken"`
}

// Metadata captures common AWS configuration fields stored in provider metadata
type Metadata struct {
	// Region is the AWS region for API calls
	Region string
	// HomeRegion is the Security Hub home region for aggregated queries
	HomeRegion string
	// LinkedRegions optionally limits queries to the listed regions
	LinkedRegions []string
	// OrganizationID is the AWS Organizations identifier associated with this integration
	OrganizationID string
	// AccountScope controls whether queries should use all accounts or a provided subset
	AccountScope string
	// AccountIDs optionally scopes collection to specific AWS account IDs
	AccountIDs []string
	// RoleARN is the ARN of the role to assume
	RoleARN string
	// AccountID is the AWS account ID
	AccountID string
	// ExternalID is the external ID for role assumption
	ExternalID string
	// SessionName is the name for the session
	SessionName string
	// SessionDuration is the duration for the session
	SessionDuration time.Duration
	// AccessKeyID is the AWS access key ID
	AccessKeyID string
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string
	// SessionToken is the AWS session token
	SessionToken string
}

// AssumeRole captures the optional STS assume-role settings
type AssumeRole struct {
	// RoleARN is the ARN of the role to assume
	RoleARN string
	// ExternalID is the external ID for role assumption
	ExternalID string
	// SessionName is the name for the session
	SessionName string
	// SessionDuration is the duration for the session
	SessionDuration time.Duration
}

// Credentials captures static AWS access key credentials
type Credentials struct {
	// AccessKeyID is the AWS access key ID
	AccessKeyID string
	// SecretAccessKey is the AWS secret access key
	SecretAccessKey string
	// SessionToken is the AWS session token
	SessionToken string
}
