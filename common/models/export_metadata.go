package models

import "io"

// ExportMetadata contains metadata for an export record.
type ExportMetadata struct {
	KeepFileOriginalName bool `json:"keepFileOriginalName,omitempty"`
}

func (e ExportMetadata) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, e)
}

func (e *ExportMetadata) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, e)
}
