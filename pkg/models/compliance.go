package models

import (
	"io"
	"slices"
	"strconv"
	"strings"
)

type Sortable interface {
	GetSortField() string
}

// ensure the types implement the Sortable interface
var (
	_ Sortable = (*ImplementationGuidance)(nil)
	_ Sortable = (*AssessmentMethod)(nil)
	_ Sortable = (*ExampleEvidence)(nil)
	_ Sortable = (*AssessmentObjective)(nil)
	_ Sortable = (*Reference)(nil)
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
	marshalGQLJSON(w, a)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (a *AssessmentObjective) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, a)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (a AssessmentMethod) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, a)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (a *AssessmentMethod) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, a)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (e ExampleEvidence) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, e)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (e *ExampleEvidence) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, e)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (i ImplementationGuidance) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, i)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (i *ImplementationGuidance) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, i)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (r Reference) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, r)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (r *Reference) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, r)
}

// GetSortField returns the field to sort on for the Sortable interface
func (i ImplementationGuidance) GetSortField() string {
	return i.ReferenceID
}

// GetSortField returns the field to sort on for the Sortable interface
func (a AssessmentMethod) GetSortField() string {
	return a.ID
}

// GetSortField returns the field to sort on for the Sortable interface
func (e ExampleEvidence) GetSortField() string {
	return e.DocumentationType + " " + e.Description
}

// GetSortField returns the field to sort on for the Sortable interface
func (a AssessmentObjective) GetSortField() string {
	return a.ID
}

// GetSortField returns the field to sort on for the Sortable interface
func (r Reference) GetSortField() string {
	return r.Name
}

// Sort a slice of Sortable items by their sort field
func Sort[T Sortable](items []T) []T {
	slices.SortFunc(items, func(a, b T) int {
		return compareStrings(a.GetSortField(), b.GetSortField())
	})

	return items
}

func compareStrings(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")

	n := len(as)
	if len(bs) < n {
		n = len(bs)
	}

	for i := range n {
		aInt, aErr := strconv.Atoi(as[i])
		bInt, bErr := strconv.Atoi(bs[i])

		if aErr == nil && bErr == nil {
			if aInt != bInt {
				return aInt - bInt
			}
		} else {
			cmp := strings.Compare(as[i], bs[i])
			if cmp != 0 {
				return cmp
			}
		}
	}

	// all equal up to min length => shorter wins
	switch {
	case len(as) < len(bs):
		return -1
	case len(as) > len(bs):
		return 1
	default:
		return 0
	}
}
