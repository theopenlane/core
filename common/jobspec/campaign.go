package jobspec

import (
	"time"

	"github.com/riverqueue/river"
)

// CampaignDispatchAction identifies which campaign dispatch workflow should run.
type CampaignDispatchAction string

// CampaignDispatchActions defines the supported dispatch actions for campaigns.
const (
	// CampaignDispatchActionLaunch dispatches the initial campaign launch.
	CampaignDispatchActionLaunch CampaignDispatchAction = "launch"
	// CampaignDispatchActionResend dispatches a resend to previously-sent targets.
	CampaignDispatchActionResend CampaignDispatchAction = "resend"
	// CampaignDispatchActionResendIncomplete dispatches a resend to incomplete targets.
	CampaignDispatchActionResendIncomplete CampaignDispatchAction = "resend_incomplete"
)

// CampaignDispatchArgs schedules or dispatches campaign notifications.
type CampaignDispatchArgs struct {
	// CampaignID is the campaign to dispatch (required).
	CampaignID string `json:"campaign_id" river:"unique"`
	// Action identifies the dispatch behavior (required).
	Action CampaignDispatchAction `json:"action" river:"unique"`
	// ScheduledAt is the requested schedule time (optional).
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" river:"unique"`
	// RequestedBy is the user that requested the dispatch (optional).
	RequestedBy string `json:"requested_by,omitempty"`
}

// Kind satisfies the river.Job interface.
func (CampaignDispatchArgs) Kind() string { return "campaign_dispatch" }

// InsertOpts provides the default configuration when scheduling this job.
func (CampaignDispatchArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:      QueueNotification,
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	}
}
