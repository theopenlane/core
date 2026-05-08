package email

import "errors"

var (
	// ErrProviderNotSupported indicates the provider string is not a recognized email provider
	ErrProviderNotSupported = errors.New("email: provider not supported")
	// ErrClientBuildFailed indicates the email client could not be constructed
	ErrClientBuildFailed = errors.New("email: client build failed")
	// ErrDispatcherNotFound indicates no registered email dispatcher matches the template key
	ErrDispatcherNotFound = errors.New("email: dispatcher not found for template key")
	// ErrInvalidOperationClient indicates the operation request client is not the expected *Client type
	ErrInvalidOperationClient = errors.New("email: invalid operation client type")
	// ErrTemplateRenderFailed indicates template rendering failed
	ErrTemplateRenderFailed = errors.New("email: template render failed")
	// ErrSendFailed indicates the email provider returned an error during send
	ErrSendFailed = errors.New("email: send failed")
	// ErrEmailTemplateNotFound indicates an email template cannot be found
	ErrEmailTemplateNotFound = errors.New("email: email template not found")
	// ErrCampaignNotFound indicates a campaign cannot be found for email dispatch
	ErrCampaignNotFound = errors.New("email: campaign not found")
	// ErrSenderNotConfigured indicates the email client has no sender configured
	ErrSenderNotConfigured = errors.New("email: sender not configured")
	// ErrCampaignMissingAssessment indicates a questionnaire campaign has no linked assessment
	ErrCampaignMissingAssessment = errors.New("email: campaign has no assessment ID")
	// ErrAssessmentNotFound indicates the assessment linked to a campaign cannot be found
	ErrAssessmentNotFound = errors.New("email: assessment not found")
	// ErrQuestionnaireDispatchFailed indicates failure during dispatch of a questionnaire access email to a campaign target
	ErrQuestionnaireDispatchFailed = errors.New("email: questionnaire dispatch failed")
)
