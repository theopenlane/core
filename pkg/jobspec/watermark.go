package jobspec

// WatermarkDocArgs for the worker to process watermarking of a document
type WatermarkDocArgs struct {
	// ID of the trust center document
	TrustCenterDocumentID string `json:"trust_center_document_id"`
}

// Kind satisfies the river.Job interface
func (WatermarkDocArgs) Kind() string { return "watermark_doc" }
