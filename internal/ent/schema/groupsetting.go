package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// GroupSetting holds the schema definition for the GroupSetting entity
type GroupSetting struct {
	SchemaFuncs

	ent.Schema
}

const SchemaGroupSetting = "group_setting"

func (GroupSetting) Name() string {
	return SchemaGroupSetting
}

func (GroupSetting) GetType() any {
	return GroupSetting.Type
}

func (GroupSetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaGroupSetting)
}

// Fields of the GroupSetting.
func (GroupSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("visibility").
			Comment("whether the group is visible to it's members / owners only or if it's searchable by anyone within the organization").
			GoType(enums.Visibility("")).
			Default(string(enums.VisibilityPublic)),
		field.Enum("join_policy").
			Comment("the policy governing ability to freely join a group, whether it requires an invitation, application, or either").
			GoType(enums.JoinPolicy("")).
			Default(string(enums.JoinPolicyInviteOrApplication)),
		field.Bool("sync_to_slack").
			Comment("whether to sync group members to slack groups").
			Optional().
			Default(false),
		field.Bool("sync_to_github").
			Comment("whether to sync group members to github groups").
			Optional().
			Default(false),
		field.String("group_id").
			Comment("the group id associated with the settings").
			Optional(),
	}
}

// Edges of the GroupSetting
func (g GroupSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: g,
			edgeSchema: Group{},
			field:      "group_id",
			ref:        "setting",
		}),
	}
}

// Annotations of the GroupSetting
func (GroupSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("group"),
	}
}

// Hooks of the GroupSetting
func (GroupSetting) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookGroupSettingVisibility(),
	}
}

// Interceptors of the GroupSetting
func (GroupSetting) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorGroupSetting(),
	}
}

// Mixin of the GroupSetting
func (GroupSetting) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
	}.getMixins()
}

// Policy defines the privacy policy of the GroupSetting
func (GroupSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.GroupSettingMutation](),
		),
	)
}
