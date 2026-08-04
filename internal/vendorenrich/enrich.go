package vendorenrich

import (
	"context"
	"slices"
	"strings"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/entity"
	"github.com/theopenlane/core/internal/ent/generated/entitytype"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/pkg/domain"
	"github.com/theopenlane/core/pkg/jsonx"
)

// systemVendorEntityType is the EntityType.Name that marks an Entity as a vendor
const systemVendorEntityType = "vendor"

// candidate is one scanned vendor's matching signals to lookup against system owned entities
type candidate struct {
	vendor map[string]any
	name   string
	domain string
}

// reference holds the known-good fields pulled off a matching system-owned vendor Entity
type reference struct {
	entityID      string
	name          string
	displayName   string
	description   string
	domains       []string
	aliases       []string
	logoRemoteURL string
	logoFileID    string
}

// EnrichVendors matches every vendor entry against Openlane's system-owned
// vendor and returns the map of data
func EnrichVendors(ctx context.Context, db *generated.Client, data map[string]any) map[string]any {
	var vendors []map[string]any
	if err := jsonx.RoundTrip(data["vendors"], &vendors); err != nil || len(vendors) == 0 {
		return data
	}

	candidates := buildCandidates(vendors)
	if len(candidates) == 0 {
		return data
	}

	references, err := lookupReferences(ctx, db, candidates)
	if err != nil || len(references) == 0 {
		return data
	}

	for _, c := range candidates {
		if ref := matchReference(c, references); ref != nil {
			applyReference(c.vendor, ref)
		}
	}

	data["vendors"] = vendors

	return data
}

// buildCandidates extracts the name/domain matching signals from each vendor entry, dropping any with neither
func buildCandidates(vendors []map[string]any) []candidate {
	candidates := make([]candidate, 0, len(vendors))

	for _, vendor := range vendors {
		name, _ := vendor["name"].(string)
		name = strings.TrimSpace(name)

		domain := vendorDomain(vendor)
		if name == "" && domain == "" {
			continue
		}

		candidates = append(candidates, candidate{vendor: vendor, name: name, domain: domain})
	}

	return candidates
}

// vendorDomain returns the lowercased hostname parsed from vendor's url
func vendorDomain(vendor map[string]any) string {
	rawURL, ok := vendor["url"].(string)
	if !ok || rawURL == "" {
		return ""
	}

	hostname, err := domain.NormalizeHostname(rawURL)
	if err != nil {
		return ""
	}

	return hostname
}

// lookupReferences queries system-owned vendor Entities matching any candidate's name, display
// name, domain, or alias in a single query scoped to the candidates at hand, rather than loading
// the entire vendor catalog
func lookupReferences(ctx context.Context, db *generated.Client, candidates []candidate) ([]reference, error) {
	const predicatesPerCandidate = 4

	predicates := make([]predicate.Entity, 0, len(candidates)*predicatesPerCandidate)
	seen := make(map[string]bool, len(candidates)*2) //nolint:mnd

	for _, c := range candidates {
		if c.name != "" && !seen["name:"+c.name] {
			seen["name:"+c.name] = true

			predicates = append(predicates,
				entity.NameEqualFold(c.name),
				entity.DisplayNameEqualFold(c.name),
				aliasesContain(c.name),
			)
		}

		if c.domain != "" && !seen["domain:"+c.domain] {
			seen["domain:"+c.domain] = true

			predicates = append(predicates, domainsContain(c.domain))
		}
	}

	if len(predicates) == 0 {
		return nil, nil
	}

	entities, err := db.Entity.Query().
		Where(
			entity.SystemOwned(true),
			entity.HasEntityTypeWith(entitytype.NameEqualFold(systemVendorEntityType)),
			entity.Or(predicates...),
		).
		Select(
			entity.FieldID,
			entity.FieldName,
			entity.FieldDisplayName,
			entity.FieldDescription,
			entity.FieldDomains,
			entity.FieldAliases,
			entity.FieldLogoRemoteURL,
			entity.FieldLogoFileID,
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	references := make([]reference, 0, len(entities))

	for _, e := range entities {
		ref := reference{
			entityID:    e.ID,
			name:        e.Name,
			displayName: e.DisplayName,
			description: e.Description,
			domains:     e.Domains,
			aliases:     e.Aliases,
		}

		if e.LogoRemoteURL != nil {
			ref.logoRemoteURL = *e.LogoRemoteURL
		}

		if e.LogoFileID != nil {
			ref.logoFileID = *e.LogoFileID
		}

		references = append(references, ref)
	}

	return references, nil
}

// domainsContain matches Entity rows whose domains JSON array contains domain
func domainsContain(domain string) predicate.Entity {
	return func(s *sql.Selector) {
		s.Where(sqljson.ValueContains(entity.FieldDomains, domain))
	}
}

// aliasesContain matches Entity rows whose aliases JSON array contains name
func aliasesContain(name string) predicate.Entity {
	return func(s *sql.Selector) {
		s.Where(sqljson.ValueContains(entity.FieldAliases, name))
	}
}

// matchReference finds the single best reference among the pre-filtered query result for
// candidate c, checking match signals in priority order: domain, name, display name, alias
func matchReference(c candidate, references []reference) *reference {
	if c.domain != "" {
		for i := range references {
			if slices.ContainsFunc(references[i].domains, func(d string) bool { return strings.EqualFold(d, c.domain) }) {
				return &references[i]
			}
		}
	}

	if c.name == "" {
		return nil
	}

	for i := range references {
		if strings.EqualFold(references[i].name, c.name) {
			return &references[i]
		}
	}

	for i := range references {
		if strings.EqualFold(references[i].displayName, c.name) {
			return &references[i]
		}
	}

	for i := range references {
		if slices.ContainsFunc(references[i].aliases, func(a string) bool { return strings.EqualFold(a, c.name) }) {
			return &references[i]
		}
	}

	return nil
}

// applyReference overwrites vendor's name/description/domains/logo fields with ref's, whenever ref has a value for them
// since the system-owned catalog is vetted data and takes precedence
// over whatever the scan found
func applyReference(vendor map[string]any, ref *reference) {
	vendor["entity_id"] = ref.entityID

	switch {
	case ref.displayName != "":
		vendor["name"] = ref.displayName
	case ref.name != "":
		vendor["name"] = ref.name
	}

	if ref.description != "" {
		vendor["description"] = ref.description
	}

	if len(ref.domains) > 0 {
		vendor["domains"] = ref.domains
	}

	if ref.logoRemoteURL != "" {
		vendor["logo_remote_url"] = ref.logoRemoteURL
	}

	if ref.logoFileID != "" {
		vendor["logo_file_id"] = ref.logoFileID
	}
}
