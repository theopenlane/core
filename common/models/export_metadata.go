package models

import "io"

// ExportMetadata contains metadata for an export record.
type ExportMetadata struct {
	KeepFileOriginalName bool `json:"keepFileOriginalName,omitempty"`
	// When exporting to PDF, the default behavior is to add metadata
	// at the top, setting this flag will exclude this from being set
	ExcludePDFMetadata bool `json:"excludePDFMetadata,omitempty"`
}

func (e ExportMetadata) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, e)
}

func (e *ExportMetadata) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, e)
}
