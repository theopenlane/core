package models

import (
	"io"

	"github.com/theopenlane/core/pkg/enums"
)

type Days []enums.JobWeekday

type JobCadence struct {
	Days      Days                      `json:"days,omitempty"`
	Time      string                    `json:"time,omitempty"`
	Frequency enums.JobCadenceFrequency `json:"frequency,omitempty"`
}

func (j JobCadence) String() string { return "" }

// MarshalGQL implement the Marshaler interface for gqlgen
func (j JobCadence) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, j)
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (j *JobCadence) UnmarshalGQL(v interface{}) error {
	return unmarshalGQLJSON(v, j)
}
