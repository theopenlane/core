package entitlements

import (
	"errors"
)

var (
	// ErrWebhookNotFound is returned when a webhook endpoint is not found
	ErrWebhookNotFound = errors.New("webhook endpoint not found")
	// ErrWebhookAlreadyExists is returned when a webhook endpoint already exists
	ErrWebhookAlreadyExists = errors.New("webhook endpoint already exists")
	// ErrWebhookVersionMismatch is returned when webhook API version does not match SDK version
	ErrWebhookVersionMismatch = errors.New("webhook API version does not match SDK version")
	// ErrMultipleWebhooksFound is returned when multiple webhooks are found for the same URL
	ErrMultipleWebhooksFound = errors.New("multiple webhook endpoints found for URL")
	// ErrWebhookURLRequired is returned when webhook URL is required but not provided
	ErrWebhookURLRequired = errors.New("webhook URL is required")
	// ErrInvalidMigrationState is returned when the migration state is invalid for the requested operation
	ErrInvalidMigrationState = errors.New("invalid migration state for requested operation")
	// ErrOldWebhookNotFound is returned when the old webhook endpoint is not found during migration
	ErrOldWebhookNotFound = errors.New("old webhook endpoint not found for migration")
	// ErrNewWebhookAlreadyExists is returned when attempting to create a new webhook that already exists
	ErrNewWebhookAlreadyExists = errors.New("new webhook endpoint already exists")
	// ErrAPIVersionRequired is returned when an API version is required but not provided
	ErrAPIVersionRequired = errors.New("api version is required")
	// ErrEnabledWebhookNotFoundByVersion is returned when no enabled webhook matches the requested version
	ErrEnabledWebhookNotFoundByVersion = errors.New("enabled webhook not found for version")
)
