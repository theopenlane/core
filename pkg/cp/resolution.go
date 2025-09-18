package cp

import (
	"context"

	"github.com/samber/mo"
)

// Resolution is a struct that represents the result of rule evaluation
type Resolution struct {
	ClientType  ProviderType
	Credentials map[string]string
	Config      map[string]any
	CacheKey    string
}

// ResolutionRule is a generic struct that evaluates context and returns resolution
type ResolutionRule[T any] struct {
	Evaluate func(ctx context.Context) mo.Option[Resolution]
}

// Resolver is a generic struct that handles rule-based client resolution
type Resolver[T any] struct {
	rules       []ResolutionRule[T]
	defaultRule mo.Option[ResolutionRule[T]]
}

// NewResolver is a constructor function that creates a resolver
func NewResolver[T any]() *Resolver[T] {
	return &Resolver[T]{
		rules: make([]ResolutionRule[T], 0),
	}
}

// AddRule is a method that adds a resolution rule to the resolver
func (r *Resolver[T]) AddRule(rule ResolutionRule[T]) *Resolver[T] {
	r.rules = append(r.rules, rule)
	return r
}


// SetDefaultRule is a method that sets a fallback rule that always matches
func (r *Resolver[T]) SetDefaultRule(rule ResolutionRule[T]) *Resolver[T] {
	r.defaultRule = mo.Some(rule)
	return r
}

// Resolve is a method that evaluates rules and returns resolution
func (r *Resolver[T]) Resolve(ctx context.Context) mo.Option[Resolution] {
	// Try each rule in order
	for _, rule := range r.rules {
		if resolution := rule.Evaluate(ctx); resolution.IsPresent() {
			return resolution
		}
	}

	// Try default rule
	if r.defaultRule.IsPresent() {
		defaultRule := r.defaultRule.MustGet()
		return defaultRule.Evaluate(ctx)
	}

	return mo.None[Resolution]()
}
