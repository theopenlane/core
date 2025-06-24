package models

import (
	"io"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/sleuth/scan"
)

// Technology describes a technology fingerprint discovered during a domain scan.
// This structure can be extended with additional metadata as needed but for now
// mirrors fields returned by the sleuth tech package
type Technology struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Website     string   `json:"website,omitempty"`
	CPE         string   `json:"cpe,omitempty"`
	Categories  []string `json:"categories,omitempty"`
}

// DomainScan stores the basic results from performing reconnaissance on a domain
// at user registration time
type DomainScan struct {
	Domain          string               `json:"domain"`
	Type            enums.ScanType       `json:"type"`
	Metadata        map[string]any       `json:"metadata,omitempty"`
	Technologies    []Technology         `json:"technologies,omitempty"`
	Vulnerabilities []scan.Vulnerability `json:"vulnerabilities,omitempty"`
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (t Technology) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, t)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (t *Technology) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, t)
}

// MarshalGQL implements the Marshaler interface for gqlgen
func (d DomainScan) MarshalGQL(w io.Writer) {
	marshalGQLJSON(w, d)
}

// UnmarshalGQL implements the Unmarshaler interface for gqlgen
func (d *DomainScan) UnmarshalGQL(v any) error {
	return unmarshalGQLJSON(v, d)
}
