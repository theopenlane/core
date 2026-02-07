package operations

import "github.com/theopenlane/core/common/integrations/types"

// Pagination captures common paging controls.
type Pagination struct {
	// PerPage sets the number of items to request per page
	PerPage int `json:"per_page"`
	// PageSize provides an alternate page size input
	PageSize int `json:"page_size"`
	// Page selects the page index to request
	Page int `json:"page"`
}

// EffectivePageSize returns the configured page size or the provided default.
func (p Pagination) EffectivePageSize(defaultValue int) int {
	if p.PerPage > 0 {
		return p.PerPage
	}
	if p.PageSize > 0 {
		return p.PageSize
	}
	return defaultValue
}

// PayloadOptions captures optional payload controls.
type PayloadOptions struct {
	// IncludePayloads controls whether raw payloads are returned
	IncludePayloads bool `json:"include_payloads"`
}

// EnsureIncludePayloads forces include_payloads to true.
func EnsureIncludePayloads(config map[string]any) map[string]any {
	if config == nil {
		config = map[string]any{}
	}
	config["include_payloads"] = true

	return config
}

// RepositorySelector captures repository selection settings.
type RepositorySelector struct {
	// Repositories lists repository names to include
	Repositories []types.TrimmedString `json:"repositories"`
	// Repos lists repository names using a shorter alias
	Repos []types.TrimmedString `json:"repos"`
	// Repository selects a single repository name
	Repository types.TrimmedString `json:"repository"`
	// Owner filters repositories by owner
	Owner types.TrimmedString `json:"owner"`
}

// List returns a merged, de-duplicated repository list.
func (r RepositorySelector) List() []string {
	out := make([]string, 0, len(r.Repositories)+len(r.Repos)+1)
	seen := map[string]struct{}{}

	appendValue := func(value types.TrimmedString) {
		if value == "" {
			return
		}
		key := string(value)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}

	for _, value := range r.Repositories {
		appendValue(value)
	}
	for _, value := range r.Repos {
		appendValue(value)
	}
	if r.Repository != "" {
		appendValue(r.Repository)
	}

	return out
}
