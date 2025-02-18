package utils

import "strings"

// Context keys for authorization checks
const (
	emailDomainContextKey = "email_domain"
	conditionContextKey   = "allowed_domains"

	// OrgEmailConditionName is the name of the condition for organization email domain checks
	OrgEmailConditionName = "email_domains_allowed"
	// OrgAccessCheckRelation is the relation for organization access checks
	OrgAccessCheckRelation = "access"
)

// NewOrganizationContextKey creates a new context key for organization checks
// if the full email is provided it will take the domain after the `@` symbol
func NewOrganizationContextKey(e string) *map[string]any {
	domain := e
	if strings.Contains(e, "@") {
		domain = strings.Split(e, "@")[1]
	}

	return &map[string]any{
		emailDomainContextKey: domain,
	}
}

// NewOrganizationConditionContext creates a new context key for organization condition checks
func NewOrganizationConditionContext(domains []string) *map[string]any {
	return &map[string]any{
		conditionContextKey: domains,
	}
}
