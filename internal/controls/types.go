package controls

import "github.com/theopenlane/core/internal/ent/generated"

type CloneOptions struct {
	programID *string
	orgID     string
}

// CloneOption is a function type that modifies the CloneOptions
type CloneOption func(*CloneOptions)

// WithProgramID sets the ProgramID in the CloneOptions
func WithProgramID(programID string) CloneOption {
	return func(co *CloneOptions) {
		co.programID = &programID
	}
}

// WithOrgID sets the OrgID in the CloneOptions
func WithOrgID(orgID string) CloneOption {
	return func(co *CloneOptions) {
		co.orgID = orgID
	}
}

// SubcontrolToCreate is used to track which subcontrols need to be created for a given control
type SubcontrolToCreate struct {
	NewControlID string
	RefControl   *generated.Control
}

// ControlToUpdate is used to track existing controls that need to be updated due to changes
// in the revision of their connected standards
type ControlToUpdate struct {
	ExistingControlID string
	SourceControl     *generated.Control
}
