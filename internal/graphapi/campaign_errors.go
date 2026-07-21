package graphapi

import "errors"

// Campaign error constants used by dispatch/test-email flows.
var (
	// ErrCampaignTestEmailTooManyRecipients is returned when a test email request exceeds the recipient cap
	ErrCampaignTestEmailTooManyRecipients = errors.New("test emails are limited to 5 recipients per request")
	// ErrCampaignTestEmailRateLimited is returned when a campaign exhausts its hourly test email budget
	ErrCampaignTestEmailRateLimited = errors.New("campaign test email hourly limit reached, try again later")
	// ErrCampaignMissingAssessmentID is returned when a campaign lacks an assessment reference.
	ErrCampaignMissingAssessmentID = errors.New("campaign is missing an assessment_id")
	// ErrCampaignDispatchInactive is returned when attempting to dispatch an inactive campaign.
	ErrCampaignDispatchInactive = errors.New("campaign is not active for dispatch")
	// ErrCampaignDispatchActionUnsupported is returned for unsupported dispatch actions.
	ErrCampaignDispatchActionUnsupported = errors.New("unsupported campaign dispatch action")
	// ErrCampaignDispatchScheduledAtInPast is returned when scheduling a dispatch in the past.
	ErrCampaignDispatchScheduledAtInPast = errors.New("scheduledAt must be in the future")
	// ErrCampaignDispatchScheduleRequired is returned when scheduling a job without a schedule time.
	ErrCampaignDispatchScheduleRequired = errors.New("schedule time is required")
	// ErrCampaignDispatchRuntimeRequired is returned when integration runtime dispatch is unavailable.
	ErrCampaignDispatchRuntimeRequired = errors.New("campaign dispatch requires integration runtime")
	// ErrCampaignTargetLimitExceeded is returned when a campaign exceeds the maximum allowed targets.
	ErrCampaignTargetLimitExceeded = errors.New("campaign cannot exceed 500 targets")
	// ErrCampaignMissingEmailTemplate is returned when a branded campaign has no linked email template.
	ErrCampaignMissingEmailTemplate = errors.New("campaign requires a linked email template for dispatch")
	// ErrCampaignMissingTrustCenter is returned when a trust center update campaign has no linked trust center.
	ErrCampaignMissingTrustCenter = errors.New("campaign requires a linked trust center for dispatch")
)
