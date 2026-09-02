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

// listenerProviders holds every listener family constructor registered by the init
// functions in this package
var listenerProviders []func() []gala.Registration

// registerListeners adds a listener family constructor to the registry; call it from an
// init function in the file that declares the constructor
func registerListeners(provider func() []gala.Registration) {
	listenerProviders = append(listenerProviders, provider)
}

// AllListeners builds the registrations for every listener family in this package; the
// constructors run on each call so they observe the current schema registry
func AllListeners() []gala.Registration {
	return lo.FlatMap(listenerProviders, func(provider func() []gala.Registration, _ int) []gala.Registration {
		return provider()
	})
}
