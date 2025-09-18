package cp

import (
	"context"

	"github.com/samber/mo"
)

// NewRule creates a rule builder for static resolution
func NewRule[T any]() *RuleBuilder[T] {
	return &RuleBuilder[T]{}
}

// RuleBuilder provides an interface for creating static resolution rules
type RuleBuilder[T any] struct {
	conditions []func(context.Context) bool
}

// DefaultRule creates a rule that always matches (for fallbacks)
func DefaultRule[T any](resolution Resolution) ResolutionRule[T] {
	return ResolutionRule[T]{
		Evaluate: func(_ context.Context) mo.Option[Resolution] {
			return mo.Some(resolution)
		},
	}
}

// WhenFunc adds a custom condition function
func (b *RuleBuilder[T]) WhenFunc(condition func(context.Context) bool) *RuleBuilder[T] {
	b.conditions = append(b.conditions, condition)
	return b
}

// ResolvedProvider represents a resolved provider configuration
type ResolvedProvider struct {
	Type        ProviderType
	Credentials map[string]string
	Config      map[string]any
}

// Resolve creates a rule that uses a function to resolve the provider
func (b *RuleBuilder[T]) Resolve(resolver func(context.Context) (*ResolvedProvider, error)) ResolutionRule[T] {
	conditions := b.conditions // Capture conditions
	return ResolutionRule[T]{
		Evaluate: func(ctx context.Context) mo.Option[Resolution] {
			// Check all conditions
			for _, condition := range conditions {
				if !condition(ctx) {
					return mo.None[Resolution]()
				}
			}

			// All conditions match, call the resolver
			provider, err := resolver(ctx)
			if err != nil || provider == nil {
				return mo.None[Resolution]()
			}

			return mo.Some(Resolution{
				ClientType:  provider.Type,
				Credentials: provider.Credentials,
				Config:      provider.Config,
			})
		},
	}
}
