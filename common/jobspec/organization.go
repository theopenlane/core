package jobspec

import "github.com/riverqueue/river"

// OrganizationDeletionReminderArgs for the peridoci worker to poll the orgs without payment methods and
// send notifications/emails to them if they are yet to add a card after some period of time
type OrganizationDeletionReminderArgs struct{}

// Kind satisfies the river.Job interface
func (OrganizationDeletionReminderArgs) Kind() string { return "org_deletion_reminder" }

// InsertOpts provides the default configuration when processing this job.
func (OrganizationDeletionReminderArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueDefault}
}

// OrganizationDeletionReminderArgs for the periodic worker to delete an organization
type OrganizationDeletionArgs struct {
	// OrganizationID is the organization that is deleted
	OrganizationID string `json:"organization_id"`
}

// Kind satisfies the river.Job interface
func (OrganizationDeletionArgs) Kind() string { return "org_deletion" }

// InsertOpts provides the default configuration when processing this job.
func (OrganizationDeletionArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueDefault}
}
