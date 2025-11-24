package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"entgo.io/ent"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/actionplan"
	"github.com/theopenlane/core/internal/ent/generated/control"
	"github.com/theopenlane/core/internal/ent/generated/customtypeenum"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/procedure"
	"github.com/theopenlane/core/internal/ent/generated/program"
	"github.com/theopenlane/core/internal/ent/generated/risk"
	"github.com/theopenlane/core/internal/ent/generated/subcontrol"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
)

var (
	// ErrCustomEnumCreationFailed is returned when a custom enum value does not exist but is attempted to be set
	ErrCustomEnumCreationFailed = errors.New("value does not exist")
	// ErrCustomEnumInUse is returned when a custom enum is in use and cannot be deleted
	ErrCustomEnumInUse = errors.New("enum is in use")
)

// CustomEnumFilter is used to filter custom enums based on object type and field
type CustomEnumFilter struct {
	// ObjectType is the object type the enum applies to, e.g. "risk", "control", "risk_category"
	ObjectType string
	// Field is the field the enum applies to, e.g. "kind", "category"
	Field string
	// EdgeFieldName is the edge field name the enum applies to that is the foreign key, e.g. "risk_kind_id"
	EdgeFieldName string
	// SchemaFieldName is the schema field name the enum applies to, e.g. "control_kind_name
	SchemaFieldName string
}

// HookCustomEnums ensures that a custom enum value exists for the given object type and field
// It looks up the enum by name and sets the corresponding edge field on the mutation
func HookCustomEnums(in CustomEnumFilter) ent.Hook {
	return hook.If(func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// get the value of the enum field from the mutation
			value, ok := m.Field(in.SchemaFieldName)
			if !ok {
				return next.Mutate(ctx, m)
			}

			// if the value is empty, skip the rest of the hook
			enumValue, ok := value.(string)
			if !ok || enumValue == "" {
				return next.Mutate(ctx, m)
			}

			// get the ent client from the mutation
			mut := m.(utils.GenericMutation)
			client := mut.Client()

			// look up the enum by name, object type, and field
			// and ensure it exists
			enum, err := client.CustomTypeEnum.Query().
				Where(
					customtypeenum.ObjectTypeEqualFold(in.ObjectType),
					customtypeenum.NameEqualFold(enumValue),
					customtypeenum.FieldEqualFold(in.Field),
					customtypeenum.DeletedAtIsNil(),
				).
				Only(ctx)
			if err != nil {
				// if the enum does not exist, return a custom error
				if generated.IsNotFound(err) {
					return nil, fmt.Errorf("%w: %s is not valid", ErrCustomEnumCreationFailed, enumValue)
				}

				return nil, err
			}

			// set the edge field on the mutation to the enum ID
			if err := m.SetField(in.EdgeFieldName, enum.ID); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, m)
		})
	}, hook.HasOp(ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne))
}

// HookCustomTypeEnumDelete checks if the enum(s) being deleted is in use by any other object.
// If in use, the deletion cannot proceed
func HookCustomTypeEnumDelete() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.CustomTypeEnumFunc(func(ctx context.Context, m *generated.CustomTypeEnumMutation) (generated.Value, error) {
			if !isDeleteOp(ctx, m) {
				return next.Mutate(ctx, m)
			}

			ids := getMutationIDs(ctx, m)
			if len(ids) == 0 {
				return next.Mutate(ctx, m)
			}

			client := m.Client()
			enums, err := client.CustomTypeEnum.Query().
				Where(customtypeenum.IDIn(ids...)).
				Select(
					customtypeenum.FieldID,
					customtypeenum.FieldObjectType,
					customtypeenum.FieldName,
				).
				All(ctx)
			if err != nil {
				return nil, err
			}

			var errs []string
			var mu sync.Mutex

			funcs := make([]func(), 0)
			for _, enum := range enums {
				funcs = append(funcs, isEnumInUse(ctx, client, enum.ID, strings.ToLower(enum.ObjectType), enum.ObjectType, enum.Name, &errs, &mu)...)
			}

			if len(funcs) == 0 {
				return next.Mutate(ctx, m)
			}

			client.PondPool.SubmitMultipleAndWait(funcs)

			if len(errs) > 0 {
				logx.FromContext(ctx).Error().
					Int("error_count", len(errs)).
					Strs("errors", errs).
					Msg("custom enum deletion failed: enums are in use")
				return nil, fmt.Errorf("%w: %d enum(s) are in use and cannot be deleted", ErrCustomEnumInUse, len(errs)) //nolint:err113
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpDeleteOne|ent.OpDelete|ent.OpUpdateOne|ent.OpUpdate)
}

func isEnumInUse(ctx context.Context, client *generated.Client, enumID, edgeName, objectType, name string, allErrors *[]string, mu *sync.Mutex) []func() {
	type countTask struct {
		queryFunc func(context.Context) (int, error)
		label     string
	}

	var operations []countTask

	switch edgeName {
	case "task":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Task.Query().Where(task.TaskKindID(enumID)).Count(ctx)
			},
			label: "task",
		})
	case "control":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Control.Query().Where(control.ControlKindID(enumID)).Count(ctx)
			},
			label: "control",
		})
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Subcontrol.Query().Where(subcontrol.SubcontrolKindID(enumID)).Count(ctx)
			},
			label: "subcontrol",
		})
	case "subcontrol":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Subcontrol.Query().Where(subcontrol.SubcontrolKindID(enumID)).Count(ctx)
			},
			label: "subcontrol",
		})
	case "risk":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Risk.Query().Where(risk.RiskKindID(enumID)).Count(ctx)
			},
			label: "risk",
		})
	case "risk_category":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Risk.Query().Where(risk.RiskCategoryID(enumID)).Count(ctx)
			},
			label: "risk",
		})
	case "internal_policy":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.InternalPolicy.Query().Where(internalpolicy.InternalPolicyKindID(enumID)).Count(ctx)
			},
			label: "internal policy",
		})
	case "procedure":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Procedure.Query().Where(procedure.ProcedureKindID(enumID)).Count(ctx)
			},
			label: "procedure",
		})
	case "action_plan":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.ActionPlan.Query().Where(actionplan.ActionPlanKindID(enumID)).Count(ctx)
			},
			label: "action plan",
		})
	case "program":
		operations = append(operations, countTask{
			queryFunc: func(ctx context.Context) (int, error) {
				return client.Program.Query().Where(program.ProgramKindID(enumID)).Count(ctx)
			},
			label: "program",
		})
	}

	funcs := make([]func(), 0, len(operations))

	for _, op := range operations {
		funcs = append(funcs, func() {
			if count, err := op.queryFunc(ctx); err == nil && count > 0 {
				mu.Lock()

				label := op.label
				if count != 1 {
					label = pluralize.NewClient().Plural(op.label)
				}

				*allErrors = append(*allErrors, fmt.Sprintf("the %s value of %s is in use by %d %s and cannot be deleted until those are updated",
					objectType, name, count, label))
				mu.Unlock()
			}
		})
	}

	return funcs
}
