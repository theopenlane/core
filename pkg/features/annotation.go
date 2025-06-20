package features

import "entgo.io/ent/schema"

// Annotation identifies which modules a schema belongs to.
// Base modules are always accessible regardless of entitlements.
type Annotation struct {
	Modules []string
	Base    bool
}

// Name implements the ent Annotation interface.
func (Annotation) Name() string { return "Feature" }

// Requires marks the schema as belonging to the provided modules.
func Requires(modules ...string) Annotation {
	return Annotation{Modules: modules}
}

// Base marks the schema as part of the base system accessible to everyone.
func Base() Annotation {
	return Annotation{Base: true}
}

// Get extracts the Annotation from a list of schema annotations if present.
func Get(annotations []schema.Annotation) (Annotation, bool) {
	for _, a := range annotations {
		if ann, ok := a.(Annotation); ok {
			return ann, true
		}
	}
	return Annotation{}, false
}
