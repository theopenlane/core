package hooks

import (
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/pkg/gala"
)

// To add a listener family, create a listeners_<name>.go file in this package with:
//
//	// XListeners does ...
//	func XListeners() []gala.Registration { ... }
//
//	func init() { registerListeners(XListeners) }
//
// The init call is the only wiring step; server setup reads the registry through
// AllListeners, so nothing outside this package needs to change. A constructor that
// takes options registers a closure supplying the defaults, and a constructor whose
// result is already embedded in another family must not register itself again.

// listener priorities order the listeners that share a concern topic, since those run
// sequentially in one dispatch; unset listeners sit at zero and keep registration order
const (
	// listenerPriorityFirst runs a listener ahead of the unprioritized ones
	listenerPriorityFirst = -100
	// listenerPriorityLast runs a listener after the unprioritized ones
	listenerPriorityLast = 100
)

// listenerProviders holds every listener family constructor registered by the init
// functions in this package
var listenerProviders []func() []gala.Registration

// registerListeners takes in a variadic amount of listeners and sets them up
func registerListeners(providers ...func() []gala.Registration) {
	listenerProviders = append(listenerProviders, providers...)
}

// AllListeners builds the registrations for every listener family in this package; the
// constructors run on each call so they observe the current schema registry
func AllListeners() []gala.Registration {
	return lo.FlatMap(listenerProviders, func(provider func() []gala.Registration, _ int) []gala.Registration {
		return provider()
	})
}
