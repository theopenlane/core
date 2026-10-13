package soc2

import (
	"fmt"

	"github.com/theopenlane/core/v2/pkg/docextract"
)

// Kind is the SOC 2 report document kind, carrying the configured prompts
type Kind struct {
	prompts Prompts
}

// NewKind builds the SOC 2 kind from configured prompts
func NewKind(prompts Prompts) Kind {
	return Kind{prompts: prompts}
}

// Name identifies the kind
func (Kind) Name() string {
	return KindName
}

// Prompts returns the configured prompts
func (k Kind) Prompts() Prompts {
	return k.prompts
}

// Plan builds the prompt and response schema for one section, scoped to the request's control
// ref codes when given; the reviews section streams and is merged across continuations
func (k Kind) Plan(req docextract.Request) (docextract.Plan, error) {
	part, ok := PartByName(req.Section)
	if !ok {
		return docextract.Plan{}, fmt.Errorf("%w: %s", ErrUnknownPart, req.Section)
	}

	prompt, err := k.prompts.sectionPrompt(part)
	if err != nil {
		return docextract.Plan{}, err
	}

	plan := docextract.Plan{Prompt: prompt, ResponseSchema: sectionSchemas[part]}

	if len(req.Scope) > 0 {
		scope, err := k.prompts.scopePrompt(req.Scope)
		if err != nil {
			return docextract.Plan{}, err
		}

		plan.Prompt += scope
		plan.Filter = newScopeFilter(req.Scope)
	}

	if part == IncludeReviews {
		plan.Stream = newReviewStream(k.prompts)
	}

	return plan, nil
}
