package utils

import "strings"

// Context keys for authorization checks
const (
	emailDomainContextKey = "email_domain"
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
