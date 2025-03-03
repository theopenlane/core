package models

import (
	"io"
)

// AssessmentObjective are objectives that are validated during the audit to ensure the control is implemented
type AssessmentObjective struct {
	// Class is the class of the assessment objective which is typically what framework it origins from
	Class string `json:"class,omitempty"`
	// ID is the unique identifier for the assessment objective
	ID string `json:"id,omitempty"`
	// Objective is the associated language describing the assessment objective
	Objective string `json:"objective,omitempty" `
}

// AssessmentMethod are methods that can be used during the audit to assess the control implementation
type AssessmentMethod struct {
	// ID is the unique identifier for the assessment method
	ID string `json:"id,omitempty"`
	// Type is the type of assessment being performed, e.g. Interview, Test, etc.
	Type string `json:"type,omitempty"`
	// Method is the associated language describing the assessment method
	Method string `json:"method,omitempty"`
}

// ExampleEvidence is example evidence that can be used to satisfy the control
type ExampleEvidence struct {
	// DocumentationType is the documentation artifact type for the example evidence
	DocumentationType string `json:"documentationType,omitempty"`
	// Description is the description of the example documentation artifact for the evidence
	Description string `json:"description,omitempty"`
}

// ImplementationGuidance is the steps to take to implement the control
// they can come directly from the control source or pulled from external sources
// if the reference id matches the control ref code, the guidance is directly from the control
// if the reference id is different, the guidance is from an external source
type ImplementationGuidance struct {
	// ReferenceID is the unique identifier for where the guidance was sourced from
	ReferenceID string `json:"referenceId,omitempty"`
	// Guidance are the steps to take to implement the control
	Guidance []string `json:"guidance,omitempty"`
}

// Reference are links to external sources that can be used to gain more information about the control
type Reference struct {
	// Name is the name of the reference
	Name string `json:"name,omitempty"`
	// URL is the link to the reference
	URL string `json:"url,omitempty"`
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (a AssessmentObjective) MarshalGQL(w io.Writer) {
	marshalGQL(w, a)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (a *AssessmentObjective) UnmarshalGQL(v any) error {
	return unmarshalGQL(v, a)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (a AssessmentMethod) MarshalGQL(w io.Writer) {
	marshalGQL(w, a)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (a *AssessmentMethod) UnmarshalGQL(v any) error {
	return unmarshalGQL(v, a)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (e ExampleEvidence) MarshalGQL(w io.Writer) {
	marshalGQL(w, e)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (e *ExampleEvidence) UnmarshalGQL(v any) error {
	return unmarshalGQL(v, e)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (i ImplementationGuidance) MarshalGQL(w io.Writer) {
	marshalGQL(w, i)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (i *ImplementationGuidance) UnmarshalGQL(v any) error {
	return unmarshalGQL(v, i)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (r Reference) MarshalGQL(w io.Writer) {
	marshalGQL(w, r)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (r *Reference) UnmarshalGQL(v any) error {
	return unmarshalGQL(v, r)
}
