package jobspec

const (
	// QueueDefault holds jobs that have not been put in any specific queue
	QueueDefault = "default"
	// QueueCompliance is the queue for compliance jobs.
	QueueCompliance = "compliance"
	// QueueTrustcenter is the queue for trust center jobs.
	QueueTrustcenter = "trustcenter"
	// QueueNotification is the queue for notification jobs.
	QueueNotification = "notifications"
)
