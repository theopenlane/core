package email

import "errors"

var (
	// ErrProviderNotSupported indicates the provider string is not a recognized email provider
	ErrProviderNotSupported = errors.New("email: provider not supported")
	// ErrClientBuildFailed indicates the email client could not be constructed
	ErrClientBuildFailed = errors.New("email: client build failed")
	// ErrTemplateNotFound indicates the requested template key is not in the registry
	ErrTemplateNotFound = errors.New("email: template not found")
	// ErrInvalidOperationClient indicates the operation request client is not the expected *EmailClient type
	ErrInvalidOperationClient = errors.New("email: invalid operation client type")
	// ErrTemplateRenderFailed indicates template rendering failed
	ErrTemplateRenderFailed = errors.New("email: template render failed")
	// ErrSendFailed indicates the email provider returned an error during send
	ErrSendFailed = errors.New("email: send failed")
	// ErrMissingRecipientAddress indicates no recipient email address was provided
	ErrMissingRecipientAddress = errors.New("email: missing recipient email address")
	// ErrMissingSenderAddress indicates no sender email address was provided
	ErrMissingSenderAddress = errors.New("email: missing sender email address")
	// ErrNotificationTemplateNotFound indicates an email notification template cannot be found
	ErrNotificationTemplateNotFound = errors.New("email: notification template not found")
	// ErrEmailTemplateNotFound indicates an email template cannot be found
	ErrEmailTemplateNotFound = errors.New("email: email template not found")
	// ErrTemplateDataInvalid indicates provided data does not satisfy template jsonschema
	ErrTemplateDataInvalid = errors.New("email: template data invalid")
	// ErrCampaignNotFound indicates a campaign cannot be found for email dispatch
	ErrCampaignNotFound = errors.New("email: campaign not found")
	// ErrSenderNotConfigured indicates the email client has no sender configured
	ErrSenderNotConfigured = errors.New("email: sender not configured")
	// ErrCampaignMissingAssessment indicates a questionnaire campaign has no linked assessment
	ErrCampaignMissingAssessment = errors.New("email: campaign has no assessment ID")
	// ErrAssessmentNotFound indicates the assessment linked to a campaign cannot be found
	ErrAssessmentNotFound = errors.New("email: assessment not found")
)
