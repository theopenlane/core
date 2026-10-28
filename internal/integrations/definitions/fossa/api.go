package fossa

import (
	"context"
	"encoding/json"
	"strconv"
)

const (
	// pathIssues lists issues for a scope, category and filter
	pathIssues = "/api/v2/issues"
	// pathIssueCategories returns issue counts per category
	pathIssueCategories = "/api/v2/issues/categories"
	// pathOrganization returns the organization details for the authenticated token
	pathOrganization = "/api/cli/organization"
)

const (
	// categoryVulnerability identifies security vulnerability issues
	categoryVulnerability = "vulnerability"
	// categoryLicensing identifies OSS license compliance issues
	categoryLicensing = "licensing"
	// statusActive limits collection to issues that still require attention
	statusActive = "active"
	// statusAll includes issues that have been dismissed in FOSSA
	statusAll = "all"
	// scopeGlobal collects issues across every project in the organization
	scopeGlobal = "global"
)

// issuesResponse is the envelope returned by the issues endpoint. Each issue is retained as raw
// JSON so the untouched provider payload is what reaches the mapping layer.
type issuesResponse struct {
	// Issues is the page of issues returned for the requested category
	Issues []json.RawMessage `json:"issues"`
}

// issueIdentity is the minimal projection of an issue needed to build an ingest envelope
type issueIdentity struct {
	// ID is the numeric FOSSA issue identifier
	ID int64 `json:"id"`
	// Projects lists the FOSSA projects the issue was found in
	Projects []issueProject `json:"projects"`
}

// issueProject is a FOSSA project reference attached to an issue
type issueProject struct {
	// ID is the FOSSA project locator, for example git+github.com/org/repo
	ID string `json:"id"`
}

// resourceID returns the project locator used as the ingest envelope resource
func (i issueIdentity) resourceID() string {
	for _, project := range i.Projects {
		if project.ID != "" {
			return project.ID
		}
	}

	return ""
}

// organizationResponse holds the organization details reported for the authenticated token
type organizationResponse struct {
	// OrganizationID is the numeric FOSSA organization identifier
	OrganizationID int64 `json:"organizationId"`
	// Subscription is the FOSSA subscription tier for the organization
	Subscription string `json:"subscription"`
	// UsesSAML reports whether the organization authenticates through SAML
	UsesSAML bool `json:"usesSAML"`
}

// identifier renders the organization ID as the stable string used for installation identity
func (o organizationResponse) identifier() string {
	if o.OrganizationID == 0 {
		return ""
	}

	return strconv.FormatInt(o.OrganizationID, 10)
}

// issueCategories returns the issue counts keyed by category
func (c *APIClient) issueCategories(ctx context.Context) (map[string]int, error) {
	counts := map[string]int{}
	if err := c.get(ctx, pathIssueCategories, nil, &counts); err != nil {
		return nil, err
	}

	return counts, nil
}

// organization returns the organization details for the authenticated token
func (c *APIClient) organization(ctx context.Context) (organizationResponse, error) {
	org := organizationResponse{}
	if err := c.get(ctx, pathOrganization, nil, &org); err != nil {
		return organizationResponse{}, err
	}

	return org, nil
}

// listIssues fetches one page of issues for the supplied category and status
func (c *APIClient) listIssues(ctx context.Context, category, status string, page int) ([]json.RawMessage, error) {
	params := map[string]string{
		"category":    category,
		"status":      status,
		"scope[type]": scopeGlobal,
		"page":        strconv.Itoa(page),
		"count":       strconv.Itoa(issuePageSize),
	}

	response := issuesResponse{}
	if err := c.get(ctx, pathIssues, params, &response); err != nil {
		return nil, err
	}

	return response.Issues, nil
}
