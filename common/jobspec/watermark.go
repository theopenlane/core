package jobspec

import "github.com/riverqueue/river"

// WatermarkDocArgs for the worker to process watermarking of a document
type WatermarkDocArgs struct {
	// TrustCenterDocumentID is the ID of the trust center document to watermark
	TrustCenterDocumentID string `json:"trust_center_document_id"`
}

// Kind satisfies the river.Job interface
func (WatermarkDocArgs) Kind() string { return "watermark_doc" }

// InsertOpts provides the default configuration when processing this job.
func (WatermarkDocArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueTrustcenter}
}
