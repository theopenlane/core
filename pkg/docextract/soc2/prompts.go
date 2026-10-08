package soc2

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

var (
	// ErrReportPromptRequired is returned when an extraction runs without the base report prompt
	ErrReportPromptRequired = errors.New("soc2: report prompt is required to extract from a soc2 report")
	// ErrPromptNotConfigured is returned when an extraction is requested for a part with no prompt
	ErrPromptNotConfigured = errors.New("soc2: no prompt configured for part")
	// ErrScopePromptRequired is returned when an extraction is scoped to controls without the scope template configured
	ErrScopePromptRequired = errors.New("soc2: scope prompt is required to scope an extraction to controls")
	// ErrContinuationPromptRequired is returned when a trimmed reviews stream cannot be resumed without the continuation template
	ErrContinuationPromptRequired = errors.New("soc2: continuation prompt is required to resume a trimmed stream")
	// ErrUnknownPart is returned when a request names a section the kind does not know
	ErrUnknownPart = errors.New("soc2: unknown report section")
)

// Prompts holds the operator-provided prompt text for SOC 2 extraction; the text is sourced from
// configuration rather than shipped in code
type Prompts struct {
	// Report is the base prompt describing the SOC 2 report layout, prepended to every section prompt
	Report string `json:"report" koanf:"report" jsonschema:"description=Base prompt describing the SOC 2 report layout"`
	// Sections maps a section name (entities, assets, controls, reviews, ...) to its extraction prompt
	Sections map[string]string `json:"sections" koanf:"sections" jsonschema:"description=Per-section extraction prompts keyed by section name"`
	// Scope is a text/template appended when a section extraction is limited to specific controls; it receives .RefCodes
	Scope string `json:"scope" koanf:"scope" jsonschema:"description=Template appended to scope an extraction to the control ref codes in .RefCodes"`
	// Continuation is a text/template sent to resume a trimmed reviews stream; it receives .LastControl and .LastCriterion
	Continuation string `json:"continuation" koanf:"continuation" jsonschema:"description=Template sent to resume a trimmed reviews stream after .LastControl mapped to .LastCriterion"`
}

// scopeVars are the values available to the Scope template
type scopeVars struct {
	RefCodes string
}

// continuationVars are the values available to the Continuation template
type continuationVars struct {
	LastControl   string
	LastCriterion string
}

// sectionVars are the values available to the Report and Sections templates
type sectionVars struct {
	// Year is the current calendar year, for naming programs and external ids
	Year int
}

// sectionPrompt builds the prompt for one part from the base report prompt plus the part's own
// prompt, rendered as templates, ending with the list of sections the model should return
func (p Prompts) sectionPrompt(part int) (string, error) {
	if p.Report == "" {
		return "", ErrReportPromptRequired
	}

	name := PartName(part)

	prompt := p.Sections[name]
	if prompt == "" {
		return "", fmt.Errorf("%w: %s", ErrPromptNotConfigured, name)
	}

	rendered, err := docextract.RenderPrompt(name, p.Report+"\n"+prompt, sectionVars{Year: time.Now().Year()})
	if err != nil {
		return "", err
	}

	requested := make([]string, 0, len(sectionSchemas[part].Properties))
	for key := range sectionSchemas[part].Properties {
		requested = append(requested, key)
	}

	slices.Sort(requested)

	return strings.TrimPrefix(rendered, "\n") + "\n\nRequested Sections: " + strings.Join(requested, ", "), nil
}

// scopePrompt renders the configured scope template for the given control ref codes
func (p Prompts) scopePrompt(refCodes []string) (string, error) {
	if p.Scope == "" {
		return "", ErrScopePromptRequired
	}

	return docextract.RenderPrompt("scope", p.Scope, scopeVars{RefCodes: strings.Join(refCodes, ", ")})
}

// continuationPrompt renders the configured continuation template for the last control processed
func (p Prompts) continuationPrompt(lastControl, lastCriterion string) (string, error) {
	if p.Continuation == "" {
		return "", ErrContinuationPromptRequired
	}

	return docextract.RenderPrompt("continuation", p.Continuation, continuationVars{LastControl: lastControl, LastCriterion: lastCriterion})
}

// UnknownSections lists configured section names the kind does not recognize, so a typo in
// config can be reported instead of silently never extracted
func (p Prompts) UnknownSections() []string {
	var unknown []string

	for name := range p.Sections {
		if _, ok := PartByName(name); !ok {
			unknown = append(unknown, name)
		}
	}

	return unknown
}

// ConfiguredParts lists the parts that have a prompt, in request order
func (p Prompts) ConfiguredParts() []string {
	parts := make([]string, 0, len(p.Sections))

	for _, name := range PartNames() {
		if p.Sections[name] != "" {
			parts = append(parts, name)
		}
	}

	return parts
}
