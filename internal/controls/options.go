package controls

// CloneOptions holds the options for cloning controls
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
