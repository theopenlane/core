package gala

import (
	"github.com/samber/do/v2"
	"github.com/samber/mo"
)

// NewInjector creates a root dependency injector with optional setup modules
func NewInjector(modules ...func(do.Injector)) do.Injector {
	return do.New(modules...)
}

// Provide registers a lazy provider for a dependency type
func Provide[T any](injector do.Injector, provider do.Provider[T]) {
	do.Provide(injector, provider)
}

// ProvideValue registers an eager value for a dependency type
func ProvideValue[T any](injector do.Injector, value T) {
	do.ProvideValue(injector, value)
}

// ProvideNamedValue registers an eager named value for a dependency type
func ProvideNamedValue[T any](injector do.Injector, name string, value T) {
	do.ProvideNamedValue(injector, name, value)
}

// Resolve resolves a dependency by type
func Resolve[T any](injector do.Injector) (T, error) {
	return do.Invoke[T](injector)
}

// ResolveNamed resolves a dependency by type and name
func ResolveNamed[T any](injector do.Injector, name string) (T, error) {
	return do.InvokeNamed[T](injector, name)
}

// ResolveOption resolves a dependency by type and returns an Option
func ResolveOption[T any](injector do.Injector) mo.Option[T] {
	value, err := Resolve[T](injector)
	if err != nil {
		return mo.None[T]()
	}

	return mo.Some(value)
}

// MustResolve resolves a dependency by type and panics on failure
func MustResolve[T any](injector do.Injector) T {
	return do.MustInvoke[T](injector)
}

// MustResolveNamed resolves a named dependency and panics on failure
func MustResolveNamed[T any](injector do.Injector, name string) T {
	return do.MustInvokeNamed[T](injector, name)
}

// ResolveFromContext resolves a dependency from handler context
func ResolveFromContext[T any](ctx HandlerContext) (T, error) {
	if ctx.Injector == nil {
		var zero T

		return zero, ErrInjectorRequired
	}

	return Resolve[T](ctx.Injector)
}

// ResolveNamedFromContext resolves a named dependency from handler context
func ResolveNamedFromContext[T any](ctx HandlerContext, name string) (T, error) {
	if ctx.Injector == nil {
		var zero T

		return zero, ErrInjectorRequired
	}

	return ResolveNamed[T](ctx.Injector, name)
}

// ResolveOptionFromContext resolves a dependency from handler context as an Option
func ResolveOptionFromContext[T any](ctx HandlerContext) mo.Option[T] {
	value, err := ResolveFromContext[T](ctx)
	if err != nil {
		return mo.None[T]()
	}

	return mo.Some(value)
}

// MustResolveFromContext resolves a dependency from handler context and panics on failure
func MustResolveFromContext[T any](ctx HandlerContext) T {
	return MustResolve[T](ctx.Injector)
}

// MustResolveNamedFromContext resolves a named dependency from handler context and panics on failure
func MustResolveNamedFromContext[T any](ctx HandlerContext, name string) T {
	return MustResolveNamed[T](ctx.Injector, name)
}
