package graphapi

import "errors"

// Campaign error constants used by dispatch/test-email flows.
var (
	// ErrCampaignDispatchNotQuestionnaire is returned when dispatching a non-questionnaire campaign.
	ErrCampaignDispatchNotQuestionnaire = errors.New("campaign must be of type QUESTIONNAIRE to dispatch")
	// ErrCampaignTestEmailNotQuestionnaire is returned when sending test emails for a non-questionnaire campaign.
	ErrCampaignTestEmailNotQuestionnaire = errors.New("campaign must be of type QUESTIONNAIRE to send a test email")
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
)
