package fossa

import (
	"context"
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
	// statusAll requests active and ignored issues so a dismissal in FOSSA closes the record here
	statusAll = "all"
	// scopeGlobal collects issues across every project in the organization
	scopeGlobal = "global"
)

// issuesResponse is the envelope returned by the issues endpoint
type issuesResponse struct {
	// Issues is the page of issues returned for the requested category
	Issues []map[string]any `json:"issues"`
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

// listIssues fetches one page of issues for the supplied category
func (c *APIClient) listIssues(ctx context.Context, category string, page int) ([]map[string]any, error) {
	params := map[string]string{
		"category":    category,
		"status":      statusAll,
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
