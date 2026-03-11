// Package awssts provides a shared AWS STS provider base used by AWS sub-providers.
// It validates AWS federation credentials from ProviderData and mints a combined
// CredentialSet containing both the STS metadata and access key fields.
package awssts
