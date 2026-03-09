package operations

import (
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/types"
)

// Pagination captures common paging controls
type Pagination struct {
	// PerPage sets the number of items to request per page
	PerPage int `json:"per_page"`
	// PageSize provides an alternate page size input
	PageSize int `json:"page_size"`
	// Page selects the page index to request
	Page int `json:"page"`
}

// EffectivePageSize returns the configured page size or the provided default
func (p Pagination) EffectivePageSize(defaultValue int) int {
	if p.PerPage > 0 {
		return p.PerPage
	}
	if p.PageSize > 0 {
		return p.PageSize
	}

	return defaultValue
}

// PayloadOptions captures optional payload controls
type PayloadOptions struct {
	// IncludePayloads controls whether raw payloads are returned
	IncludePayloads bool `json:"include_payloads"`
}

// EnsureIncludePayloads forces include_payloads to true in a JSON config document.
func EnsureIncludePayloads(config json.RawMessage) json.RawMessage {
	var m map[string]json.RawMessage
	if len(config) > 0 {
		if err := json.Unmarshal(config, &m); err != nil {
			m = map[string]json.RawMessage{}
		}
	} else {
		m = map[string]json.RawMessage{}
	}

	m["include_payloads"] = json.RawMessage(`true`)

	out, err := json.Marshal(m)
	if err != nil {
		return config
	}

	return out
}

// RepositorySelector captures repository selection settings
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

// List returns a merged, de-duplicated repository list
func (r RepositorySelector) List() []string {
	combined := make([]types.TrimmedString, 0, len(r.Repositories)+len(r.Repos)+1)
	combined = append(combined, r.Repositories...)
	combined = append(combined, r.Repos...)
	combined = append(combined, r.Repository)

	return lo.Uniq(lo.FilterMap(combined, func(v types.TrimmedString, _ int) (string, bool) {
		return v.String(), v != ""
	}))
}
