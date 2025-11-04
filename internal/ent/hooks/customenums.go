package hooks

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/customtypeenum"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
)

var (
	// ErrCustomEnumCreationFailed is returned when a custom enum value does not exist but is attempted to be set
	ErrCustomEnumCreationFailed = errors.New("value does not exist")
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
