package soc2

import (
	"encoding/json"
	"slices"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

// scopedEntry is the part of an extracted entry the scope filter reads
type scopedEntry struct {
	RefCodes []string `json:"refCodes"`
}

// scopeFilter drops entries that reference none of the control ref codes a batch was scoped to,
// which is how fabricated findings and reviews show up
type scopeFilter struct {
	refCodes []string
}

// newScopeFilter builds a filter for the given control ref codes
func newScopeFilter(refCodes []string) scopeFilter {
	return scopeFilter{refCodes: refCodes}
}

// Filter prunes every section array down to entries referencing an in-scope ref code
func (f scopeFilter) Filter(sections docextract.Sections) (docextract.Sections, int) {
	dropped := 0

	for name, raw := range sections {
		var entries []json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil {
			continue
		}

		kept := make([]json.RawMessage, 0, len(entries))

		for _, entry := range entries {
			if f.inScope(entry) {
				kept = append(kept, entry)
			}
		}

		if len(kept) == len(entries) {
			continue
		}

		filtered, err := json.Marshal(kept)
		if err != nil {
			continue
		}

		dropped += len(entries) - len(kept)
		sections[name] = filtered
	}

	return sections, dropped
}

// inScope reports whether the entry references at least one scoped ref code
func (f scopeFilter) inScope(entry json.RawMessage) bool {
	var scoped scopedEntry
	if err := json.Unmarshal(entry, &scoped); err != nil {
		return false
	}

	return slices.ContainsFunc(scoped.RefCodes, f.isScopedRefCode)
}

// isScopedRefCode reports whether the ref code is one the batch was scoped to
func (f scopeFilter) isScopedRefCode(refCode string) bool {
	return slices.Contains(f.refCodes, refCode)
}

// ControlRefCodes lists the ref codes from an extracted controls section, in report order, so the
// control-scoped sections can be requested in batches
func ControlRefCodes(controls json.RawMessage) []string {
	var items []struct {
		RefCode string `json:"refCode"`
	}

	if err := json.Unmarshal(controls, &items); err != nil {
		return nil
	}

	refCodes := make([]string, 0, len(items))

	for _, item := range items {
		if item.RefCode != "" && !slices.Contains(refCodes, item.RefCode) {
			refCodes = append(refCodes, item.RefCode)
		}
	}

	return refCodes
}
