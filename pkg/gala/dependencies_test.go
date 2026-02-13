package gala

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/do/v2"
)

type dependencyTestConfig struct {
	Name string
}

func expectPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic")
		}
	}()

	fn()
}

func TestDependencyHelpersProvideAndResolve(t *testing.T) {
	injector := NewInjector()

	ProvideValue(injector, "root")
	ProvideNamedValue(injector, "named_cfg", dependencyTestConfig{Name: "named"})
	Provide(injector, func(i do.Injector) (int, error) {
		value, err := Resolve[string](i)
		if err != nil {
			return 0, err
		}

		return len(value), nil
	})

	resolvedLength, err := Resolve[int](injector)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}

	if resolvedLength != len("root") {
		t.Fatalf("unexpected resolved int: %d", resolvedLength)
	}

	resolvedNamed, err := ResolveNamed[dependencyTestConfig](injector, "named_cfg")
	if err != nil {
		t.Fatalf("unexpected named resolve error: %v", err)
	}

	if resolvedNamed.Name != "named" {
		t.Fatalf("unexpected named resolve value: %#v", resolvedNamed)
	}

	option := ResolveOption[int](injector)
	if !option.IsPresent() {
		t.Fatalf("expected option to be present")
	}

	optionValue, ok := option.Get()
	if !ok {
		t.Fatalf("expected present option value")
	}

	if optionValue != len("root") {
		t.Fatalf("unexpected option value: %d", optionValue)
	}
}

func TestDependencyHelpersResolveOptionReturnsNoneWhenMissing(t *testing.T) {
	injector := NewInjector()

	option := ResolveOption[bool](injector)
	if option.IsPresent() {
		t.Fatalf("expected missing option")
	}
}

func TestDependencyHelpersContextResolution(t *testing.T) {
	injector := NewInjector()
	ProvideValue(injector, dependencyTestConfig{Name: "ctx"})
	ProvideNamedValue(injector, "ctx_named", dependencyTestConfig{Name: "ctx_named"})

	handlerContext := HandlerContext{
		Context:  context.Background(),
		Injector: injector,
	}

	resolved, err := ResolveFromContext[dependencyTestConfig](handlerContext)
	if err != nil {
		t.Fatalf("unexpected context resolve error: %v", err)
	}

	if resolved.Name != "ctx" {
		t.Fatalf("unexpected context resolve value: %#v", resolved)
	}

	option := ResolveOptionFromContext[dependencyTestConfig](handlerContext)
	if !option.IsPresent() {
		t.Fatalf("expected context option to be present")
	}

	mustResolved := MustResolveFromContext[dependencyTestConfig](handlerContext)
	if mustResolved.Name != "ctx" {
		t.Fatalf("unexpected must-resolved value: %#v", mustResolved)
	}

	resolvedNamed, err := ResolveNamedFromContext[dependencyTestConfig](handlerContext, "ctx_named")
	if err != nil {
		t.Fatalf("unexpected named context resolve error: %v", err)
	}
	if resolvedNamed.Name != "ctx_named" {
		t.Fatalf("unexpected named context resolve value: %#v", resolvedNamed)
	}

	mustResolvedNamed := MustResolveNamedFromContext[dependencyTestConfig](handlerContext, "ctx_named")
	if mustResolvedNamed.Name != "ctx_named" {
		t.Fatalf("unexpected must-resolved named value: %#v", mustResolvedNamed)
	}

	missing := ResolveOptionFromContext[int](handlerContext)
	if missing.IsPresent() {
		t.Fatalf("expected missing context option")
	}
}

func TestResolveFromContextRequiresInjector(t *testing.T) {
	_, err := ResolveFromContext[string](HandlerContext{})
	if !errors.Is(err, ErrInjectorRequired) {
		t.Fatalf("expected ErrInjectorRequired, got %v", err)
	}

	_, err = ResolveNamedFromContext[string](HandlerContext{}, "missing")
	if !errors.Is(err, ErrInjectorRequired) {
		t.Fatalf("expected ErrInjectorRequired for named resolve, got %v", err)
	}

	option := ResolveOptionFromContext[string](HandlerContext{})
	if option.IsPresent() {
		t.Fatalf("expected missing option for nil injector")
	}
}

func TestDependencyHelpersMustResolveNamed(t *testing.T) {
	injector := NewInjector()
	ProvideNamedValue(injector, "named_cfg", dependencyTestConfig{Name: "named"})

	resolved := MustResolveNamed[dependencyTestConfig](injector, "named_cfg")
	if resolved.Name != "named" {
		t.Fatalf("unexpected must-resolved named value: %#v", resolved)
	}
}

func TestDependencyHelpersMustResolvePanicsWhenMissing(t *testing.T) {
	injector := NewInjector()

	expectPanic(t, func() {
		_ = MustResolve[int](injector)
	})

	expectPanic(t, func() {
		_ = MustResolveNamed[int](injector, "missing")
	})

	expectPanic(t, func() {
		_ = MustResolveFromContext[int](HandlerContext{
			Context:  context.Background(),
			Injector: injector,
		})
	})

	expectPanic(t, func() {
		_ = MustResolveNamedFromContext[int](HandlerContext{
			Context:  context.Background(),
			Injector: injector,
		}, "missing")
	})
}
