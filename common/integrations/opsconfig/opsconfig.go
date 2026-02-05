package opsconfig

import "github.com/theopenlane/core/common/integrations/helpers"

// Pagination captures common paging controls.
type Pagination struct {
	PerPage  int `mapstructure:"per_page"`
	PageSize int `mapstructure:"page_size"`
	Page     int `mapstructure:"page"`
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
	IncludePayloads bool `mapstructure:"include_payloads"`
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
	Repositories []helpers.TrimmedString `mapstructure:"repositories"`
	Repos        []helpers.TrimmedString `mapstructure:"repos"`
	Repository   helpers.TrimmedString   `mapstructure:"repository"`
	Owner        helpers.TrimmedString   `mapstructure:"owner"`
}

// List returns a merged, de-duplicated repository list.
func (r RepositorySelector) List() []string {
	out := make([]string, 0, len(r.Repositories)+len(r.Repos)+1)
	seen := map[string]struct{}{}

	appendValue := func(value helpers.TrimmedString) {
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
