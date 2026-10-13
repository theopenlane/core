package operations

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
)

// EntitiesByDomain maps each domain to the organization entity already listing it in its domains, so an import
// targets that entity instead of creating a second one for the same vendor
func EntitiesByDomain(ctx context.Context, client *ent.Client, domains []string) (map[string]string, error) {
	byDomain := map[string]string{}

	if len(domains) == 0 {
		return byDomain, nil
	}

	entities, err := client.Entity.Query().Where(predicate.Entity(domainsContainAny(domains))).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up existing entities by domain: %w", err)
	}

	for _, e := range entities {
		for _, domain := range e.Domains {
			domain = NormalizeDomain(domain)

			if _, seen := byDomain[domain]; !seen && lo.Contains(domains, domain) {
				byDomain[domain] = e.ID
			}
		}
	}

	return byDomain, nil
}

// NormalizeDomain lowercases a domain and drops its trailing dot so domains from different sources compare equal
func NormalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
}

// domainsContainAny matches entities whose domains list contains any of the given domains
func domainsContainAny(domains []string) func(*sql.Selector) {
	return func(s *sql.Selector) {
		s.Where(sql.Or(lo.Map(domains, func(domain string, _ int) *sql.Predicate {
			return sqljson.ValueContains(s.C(entity.FieldDomains), domain)
		})...))
	}
}
