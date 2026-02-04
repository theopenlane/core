package opsconfig

import (
	"strings"
)

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

// RepositorySelector captures repository selection settings.
type RepositorySelector struct {
	Repositories []string `mapstructure:"repositories"`
	Repos        []string `mapstructure:"repos"`
	Repository   string   `mapstructure:"repository"`
	Owner        string   `mapstructure:"owner"`
}

// List returns a merged, de-duplicated repository list.
func (r RepositorySelector) List() []string {
	out := make([]string, 0, len(r.Repositories)+len(r.Repos)+1)
	seen := map[string]struct{}{}

	appendValue := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		out = append(out, value)
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
