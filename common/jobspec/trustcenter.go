package jobspec

import "github.com/riverqueue/river"

// AttestNDARequestArgs for the worker to process the signed nda pdf
type AttestNDARequestArgs struct {
	// NDARequestID is the id of the request
	NDARequestID string `json:"nda_request_id,omitempty"`
}

// Kind satisfies the river.Job interface
func (AttestNDARequestArgs) Kind() string { return "sign_nda_args" }

// InsertOpts provides the default configuration when processing this job.
func (AttestNDARequestArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueTrustcenter}
}
