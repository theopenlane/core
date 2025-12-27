package jobspec

// CreatePirschDomainArgs for the worker to process the pirsch domain creation
type CreatePirschDomainArgs struct {
	// TrustCenterID is the ID of the trust center to create a Pirsch domain for
	TrustCenterID string `json:"trust_center_id"`
}

// Kind satisfies the river.Job interface
func (CreatePirschDomainArgs) Kind() string { return "create_pirsch_domain" }

// DeletePirschDomainArgs for the worker to delete a Pirsch domain
type DeletePirschDomainArgs struct {
	// PirschDomainID is the ID of the Pirsch domain to delete
	PirschDomainID string `json:"pirsch_domain_id"`
}

// Kind satisfies the river.Job interface
func (DeletePirschDomainArgs) Kind() string { return "delete_pirsch_domain" }

// UpdatePirschDomainArgs for the worker to update the pirsch domain
type UpdatePirschDomainArgs struct {
	// TrustCenterID is the ID of the trust center to update the Pirsch domain for
	TrustCenterID string `json:"trust_center_id"`
}

// Kind satisfies the river.Job interface
func (UpdatePirschDomainArgs) Kind() string { return "update_pirsch_domain" }
