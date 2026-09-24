package docextract

import "regexp"

// MinExtractedText is the fewest characters of text a pdf must yield before markers are checked;
// below it the file is treated as scanned or image only rather than as the wrong kind of document
const MinExtractedText = 200

// Confidence grades how strongly a validated document matched its profile
type Confidence string

const (
	// ConfidenceHigh marks a document whose strong marker count reached the profile's high-confidence threshold
	ConfidenceHigh Confidence = "high"
	// ConfidenceAccepted marks a document that met the profile's minimum strong marker count
	ConfidenceAccepted Confidence = "accepted"
)

// String returns the confidence as a string
func (c Confidence) String() string { return string(c) }

// ReportType is the subtype a profile detected for a validated document
type ReportType string

// ReportTypeUnknown is used when the document subtype could not be determined from the markers
const ReportTypeUnknown ReportType = "unknown"

// String returns the report type as a string
func (r ReportType) String() string { return string(r) }

// MarkerWeight is how much a matched marker counts toward accepting a document
type MarkerWeight string

const (
	// MarkerRequired markers must be present for the document to be the expected kind at all
	MarkerRequired MarkerWeight = "required"
	// MarkerStrong markers are sections every document of the kind contains and count toward acceptance
	MarkerStrong MarkerWeight = "strong"
	// MarkerSupporting markers are common but not decisive and never affect acceptance
	MarkerSupporting MarkerWeight = "supporting"
)

// Marker is one phrase to look for in the document text and how much it counts toward acceptance
type Marker struct {
	// Name identifies the marker in the validation result
	Name string
	// Weight is how much the marker counts toward acceptance
	Weight MarkerWeight
	// Pattern is matched against the normalized document text
	Pattern *regexp.Regexp
}

// Profile describes how to recognize one kind of document
type Profile struct {
	// Kind names the document kind in user-facing failure reasons, e.g. "SOC 2 report"
	Kind string
	// MinPages is the fewest pages the document can have
	MinPages int
	// MinStrong is how many strong markers must accompany the required markers to accept the document
	MinStrong int
	// HighConfidence is how many strong markers make the document a high-confidence match
	HighConfidence int
	// Markers is the table of phrases to look for
	Markers []Marker
	// Subtype optionally classifies the document from the matched markers, e.g. SOC 2 Type 1 vs Type 2
	Subtype func(found map[string]bool) ReportType
	// MissingSectionsHint names the sections a rejected document lacked, appended to the failure reason
	MissingSectionsHint string
}

// Validation is the outcome of checking whether a pdf is the expected kind of document
type Validation struct {
	// Kind is the document kind the pdf was validated against
	Kind string `json:"kind"`
	// Pages is the pdf page count
	Pages int `json:"pages"`
	// TextLength is how many characters of text were extracted, useful when a rejection looks wrong
	TextLength int `json:"textLength"`
	// Confidence is how strongly the document matched the profile
	Confidence Confidence `json:"confidence"`
	// ReportType is the detected subtype, e.g. SOC 2 type1 or type2
	ReportType ReportType `json:"reportType"`
	// Strong lists the strong markers found
	Strong []string `json:"strong"`
	// Supporting lists the supporting markers found
	Supporting []string `json:"supporting"`
}
