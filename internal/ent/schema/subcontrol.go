package schema

import (
	"context"
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/oscalgen"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/groupmembership"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/directives"
)

// Subcontrol defines the file schema.
type Subcontrol struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSubcontrol is the name of the Subcontrol schema.
const SchemaSubcontrol = "subcontrol"

// Name returns the name of the Subcontrol schema.
func (Subcontrol) Name() string {
	return SchemaSubcontrol
}

// GetType returns the type of the Subcontrol schema.
func (Subcontrol) GetType() any {
	return Subcontrol.Type
}

// PluralName returns the plural name of the Subcontrol schema.
func (Subcontrol) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSubcontrol)
}

// Fields returns file fields.
func (Subcontrol) Fields() []ent.Field {
	// add any fields that are specific to the subcontrol here
	additionalFields := []ent.Field{
		field.String("ref_code").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("ref_code"),
				directives.ExternalSourceDirectiveAnnotation,
				oscalgen.NewOSCALField(
					oscalgen.OSCALFieldRoleStatementID,
					oscalgen.WithOSCALFieldModels(oscalgen.OSCALModelComponentDefinition, oscalgen.OSCALModelSSP),
					oscalgen.WithOSCALIdentityAnchor(),
				),
			).
			Comment("the unique reference code for the control"),
		field.String("control_id").
			Unique().
			Comment("the id of the parent control").
			NotEmpty(),
	}

	return additionalFields
}

// Edges of the Subcontrol
func (s Subcontrol) Edges() []ent.Edge {
	return []ent.Edge{
		// subcontrols are required to have a parent control
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Control{},
			field:      "control_id",
			required:   true,
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
				oscalgen.NewOSCALRelationship(
					oscalgen.OSCALRelationshipRoleLinksToControlID,
					oscalgen.WithOSCALRelationshipModels(oscalgen.OSCALModelComponentDefinition, oscalgen.OSCALModelSSP),
				),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ControlImplementation{},
			comment:    "the implementation(s) of the subcontrol",
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
				oscalgen.NewOSCALRelationship(
					oscalgen.OSCALRelationshipRoleLinksToStatementID,
					oscalgen.WithOSCALRelationshipModels(oscalgen.OSCALModelComponentDefinition, oscalgen.OSCALModelSSP),
				),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ScheduledJob{},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			ref:        "to_subcontrols",
			name:       "mapped_to_subcontrols",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			ref:        "from_subcontrols",
			name:       "mapped_from_subcontrols",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: WorkflowObjectRef{},
			ref:        "subcontrol",
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		defaultEdgeToWithPagination(s, Asset{}),
		defaultEdgeToWithPagination(s, Entity{}),
		defaultEdgeToWithPagination(s, IdentityHolder{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Vulnerability{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Vulnerability{}.Name()),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Finding{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Finding{}.Name()),
			},
		}),
	}
}

// Mixin of the Subcontrol
func (s Subcontrol) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "SCL",
		additionalMixins: []ent.Mixin{
			// add the common overlap between control and subcontrol
			ControlMixin{
				SchemaType: s,
			},
			// subcontrols inherit view access from the parent control via FGA
			newObjectOwnedMixin[generated.Subcontrol](s,
				withParents(Control{}),
				withOrganizationOwner(),
				withSkipForSystemAdmin(),
				withSkipFilterInterceptor(interceptors.SkipAllQuery|interceptors.SkipIDsQuery),
			),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
			newCustomEnumMixin(s, withWorkflowEnumEdges()),
			WorkflowApprovalMixin{},
		},
	}.getMixins(s)
}

// Indexes of the Subcontrol
func (Subcontrol) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the parent control
		index.Fields("control_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
		index.Fields("control_id", "ref_code", "owner_id").
			Annotations(
				entsql.IndexWhere("deleted_at is NULL"),
			),
		index.Fields("reference_id", "deleted_at", "owner_id"),
		index.Fields("auditor_reference_id", "deleted_at", "owner_id"),
	}
}

// Hooks of the Subcontrol
func (Subcontrol) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookSubcontrolCreate(),
		hooks.HookSubcontrolUpdate(),
	}
}

// Interceptors of the Subcontrol
func (Subcontrol) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// check the parent control's blocked groups so list queries don't require per-object FGA checks
		parentBlockedGroupsInterceptor(SchemaControl),
	}
}

// Policy of the Subcontrol
func (s Subcontrol) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowIfContextAllowRule(),
			policy.CanCreateObjectsUnderParents([]string{
				Control{}.Name(),
			}),
			policy.CheckCreateAccess(),
			rule.CheckIfCommentOnly(),
			entfga.CheckEditAccess[*generated.SubcontrolMutation](),
		),
	)
}

// Annotations of the Subcontrol
func (Subcontrol) Annotations() []schema.Annotation {
	return []schema.Annotation{
		oscalgen.NewOSCALModel(
			oscalgen.WithOSCALModels(oscalgen.OSCALModelComponentDefinition, oscalgen.OSCALModelSSP),
			oscalgen.WithOSCALAssembly("implemented-requirement-statement"),
		),
		entx.FGACrudParent(Control{}.Name()),
		// list queries use SQL blocked group filtering via newParentBlockedGroupsInterceptor
		SkipFGAOverfetch{Enabled: true},
	}
}

// Annotations of the Standard
func (Subcontrol) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// parentBlockedGroupsInterceptor returns an interceptor that filters out objects
// where the user's group is in the parent type's blocked_groups join table.
// This allows skipping per-object FGA checks on list queries when access is
// controlled by the parent's blocked groups rather than the object's own
// to be used for subcontrols, whose parent is controls
func parentBlockedGroupsInterceptor(parentType string) ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		if _, ok := auth.ActiveTrustCenterIDKey.Get(ctx); ok {
			return nil
		}

		caller, ok := auth.CallerFromContext(ctx)
		if !ok || caller == nil {
			return auth.ErrNoAuthUser
		}

		if skip := groupPermissionInterceptorSkipper(ctx, caller); skip {
			return nil
		}

		groupIDs, err := generated.FromContext(ctx).Group.Query().Where(
			group.HasMembersWith(
				groupmembership.UserID(caller.SubjectID),
			),
		).IDs(ctx)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to get group IDs for user")

			return err
		}

		addParentBlockedGroupPredicate(q, groupIDs, parentType)

		return nil
	})
}

// addParentBlockedGroupPredicate adds a predicate that excludes objects whose parent
// has the user's group in the parent type's blocked_groups join table
func addParentBlockedGroupPredicate(q intercept.Query, groupIDs []string, parentType string) {
	parentSnakeCase := strcase.SnakeCase(parentType)
	parentFKField := fmt.Sprintf("%s_id", parentSnakeCase)
	joinTableName := fmt.Sprintf("%s_blocked_groups", parentSnakeCase)

	if len(groupIDs) == 0 {
		return
	}

	q.WhereP(func(s *sql.Selector) {
		t := sql.Table(joinTableName).As(joinTableName)
		subquery := sql.SelectExpr(sql.Raw("1")).From(t).Where(
			sql.And(
				sql.EQ(t.C(parentFKField), s.C(parentFKField)),
				sql.In(t.C("group_id"), lo.ToAnySlice(groupIDs)...),
			),
		)
		s.Where(sql.Not(sql.Exists(subquery)))
	})
}
