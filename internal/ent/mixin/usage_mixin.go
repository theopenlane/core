package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/pkg/enums"
)

// UsageMixin attaches a usage hook so that creations count toward an organization's usage for the provided type.
// Schemas embedding this mixin must expose OwnerID() and Client() on their mutations, which is the case for org owned entities in this repo
type UsageMixin struct {
	mixin.Schema

	// Type defines which resource type this record consumes
	Type enums.UsageType
}

// Hooks registers the usage hook if a valid type is configured
func (m UsageMixin) Hooks() []ent.Hook {
	if m.Type == "" || m.Type == enums.UsageInvalid {
		return nil
	}

	return []ent.Hook{hooks.HookUsage(m.Type)}
}
