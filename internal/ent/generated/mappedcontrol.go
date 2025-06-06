// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/mappedcontrol"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/enums"
)

// MappedControl is the model entity for the MappedControl schema.
type MappedControl struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// CreatedBy holds the value of the "created_by" field.
	CreatedBy string `json:"created_by,omitempty"`
	// UpdatedBy holds the value of the "updated_by" field.
	UpdatedBy string `json:"updated_by,omitempty"`
	// DeletedAt holds the value of the "deleted_at" field.
	DeletedAt time.Time `json:"deleted_at,omitempty"`
	// DeletedBy holds the value of the "deleted_by" field.
	DeletedBy string `json:"deleted_by,omitempty"`
	// tags associated with the object
	Tags []string `json:"tags,omitempty"`
	// the organization id that owns the object
	OwnerID string `json:"owner_id,omitempty"`
	// the type of mapping between the two controls, e.g. subset, intersect, equal, superset
	MappingType enums.MappingType `json:"mapping_type,omitempty"`
	// description of how the two controls are related
	Relation string `json:"relation,omitempty"`
	// percentage (0-100) of confidence in the mapping
	Confidence *int `json:"confidence,omitempty"`
	// source of the mapping, e.g. manual, suggested, etc.
	Source enums.MappingSource `json:"source,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the MappedControlQuery when eager-loading is set.
	Edges        MappedControlEdges `json:"edges"`
	selectValues sql.SelectValues
}

// MappedControlEdges holds the relations/edges for other nodes in the graph.
type MappedControlEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Organization `json:"owner,omitempty"`
	// groups that are blocked from viewing or editing the risk
	BlockedGroups []*Group `json:"blocked_groups,omitempty"`
	// provides edit access to the risk to members of the group
	Editors []*Group `json:"editors,omitempty"`
	// controls that map to another control
	FromControls []*Control `json:"from_controls,omitempty"`
	// controls that are being mapped from another control
	ToControls []*Control `json:"to_controls,omitempty"`
	// subcontrols map to another control
	FromSubcontrols []*Subcontrol `json:"from_subcontrols,omitempty"`
	// subcontrols are being mapped from another control
	ToSubcontrols []*Subcontrol `json:"to_subcontrols,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [7]bool
	// totalCount holds the count of the edges above.
	totalCount [7]map[string]int

	namedBlockedGroups   map[string][]*Group
	namedEditors         map[string][]*Group
	namedFromControls    map[string][]*Control
	namedToControls      map[string][]*Control
	namedFromSubcontrols map[string][]*Subcontrol
	namedToSubcontrols   map[string][]*Subcontrol
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e MappedControlEdges) OwnerOrErr() (*Organization, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// BlockedGroupsOrErr returns the BlockedGroups value or an error if the edge
// was not loaded in eager-loading.
func (e MappedControlEdges) BlockedGroupsOrErr() ([]*Group, error) {
	if e.loadedTypes[1] {
		return e.BlockedGroups, nil
	}
	return nil, &NotLoadedError{edge: "blocked_groups"}
}

// EditorsOrErr returns the Editors value or an error if the edge
// was not loaded in eager-loading.
func (e MappedControlEdges) EditorsOrErr() ([]*Group, error) {
	if e.loadedTypes[2] {
		return e.Editors, nil
	}
	return nil, &NotLoadedError{edge: "editors"}
}

// FromControlsOrErr returns the FromControls value or an error if the edge
// was not loaded in eager-loading.
func (e MappedControlEdges) FromControlsOrErr() ([]*Control, error) {
	if e.loadedTypes[3] {
		return e.FromControls, nil
	}
	return nil, &NotLoadedError{edge: "from_controls"}
}

// ToControlsOrErr returns the ToControls value or an error if the edge
// was not loaded in eager-loading.
func (e MappedControlEdges) ToControlsOrErr() ([]*Control, error) {
	if e.loadedTypes[4] {
		return e.ToControls, nil
	}
	return nil, &NotLoadedError{edge: "to_controls"}
}

// FromSubcontrolsOrErr returns the FromSubcontrols value or an error if the edge
// was not loaded in eager-loading.
func (e MappedControlEdges) FromSubcontrolsOrErr() ([]*Subcontrol, error) {
	if e.loadedTypes[5] {
		return e.FromSubcontrols, nil
	}
	return nil, &NotLoadedError{edge: "from_subcontrols"}
}

// ToSubcontrolsOrErr returns the ToSubcontrols value or an error if the edge
// was not loaded in eager-loading.
func (e MappedControlEdges) ToSubcontrolsOrErr() ([]*Subcontrol, error) {
	if e.loadedTypes[6] {
		return e.ToSubcontrols, nil
	}
	return nil, &NotLoadedError{edge: "to_subcontrols"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*MappedControl) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case mappedcontrol.FieldTags:
			values[i] = new([]byte)
		case mappedcontrol.FieldConfidence:
			values[i] = new(sql.NullInt64)
		case mappedcontrol.FieldID, mappedcontrol.FieldCreatedBy, mappedcontrol.FieldUpdatedBy, mappedcontrol.FieldDeletedBy, mappedcontrol.FieldOwnerID, mappedcontrol.FieldMappingType, mappedcontrol.FieldRelation, mappedcontrol.FieldSource:
			values[i] = new(sql.NullString)
		case mappedcontrol.FieldCreatedAt, mappedcontrol.FieldUpdatedAt, mappedcontrol.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the MappedControl fields.
func (mc *MappedControl) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case mappedcontrol.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				mc.ID = value.String
			}
		case mappedcontrol.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				mc.CreatedAt = value.Time
			}
		case mappedcontrol.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				mc.UpdatedAt = value.Time
			}
		case mappedcontrol.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				mc.CreatedBy = value.String
			}
		case mappedcontrol.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				mc.UpdatedBy = value.String
			}
		case mappedcontrol.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				mc.DeletedAt = value.Time
			}
		case mappedcontrol.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				mc.DeletedBy = value.String
			}
		case mappedcontrol.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &mc.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case mappedcontrol.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				mc.OwnerID = value.String
			}
		case mappedcontrol.FieldMappingType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field mapping_type", values[i])
			} else if value.Valid {
				mc.MappingType = enums.MappingType(value.String)
			}
		case mappedcontrol.FieldRelation:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field relation", values[i])
			} else if value.Valid {
				mc.Relation = value.String
			}
		case mappedcontrol.FieldConfidence:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field confidence", values[i])
			} else if value.Valid {
				mc.Confidence = new(int)
				*mc.Confidence = int(value.Int64)
			}
		case mappedcontrol.FieldSource:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field source", values[i])
			} else if value.Valid {
				mc.Source = enums.MappingSource(value.String)
			}
		default:
			mc.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the MappedControl.
// This includes values selected through modifiers, order, etc.
func (mc *MappedControl) Value(name string) (ent.Value, error) {
	return mc.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the MappedControl entity.
func (mc *MappedControl) QueryOwner() *OrganizationQuery {
	return NewMappedControlClient(mc.config).QueryOwner(mc)
}

// QueryBlockedGroups queries the "blocked_groups" edge of the MappedControl entity.
func (mc *MappedControl) QueryBlockedGroups() *GroupQuery {
	return NewMappedControlClient(mc.config).QueryBlockedGroups(mc)
}

// QueryEditors queries the "editors" edge of the MappedControl entity.
func (mc *MappedControl) QueryEditors() *GroupQuery {
	return NewMappedControlClient(mc.config).QueryEditors(mc)
}

// QueryFromControls queries the "from_controls" edge of the MappedControl entity.
func (mc *MappedControl) QueryFromControls() *ControlQuery {
	return NewMappedControlClient(mc.config).QueryFromControls(mc)
}

// QueryToControls queries the "to_controls" edge of the MappedControl entity.
func (mc *MappedControl) QueryToControls() *ControlQuery {
	return NewMappedControlClient(mc.config).QueryToControls(mc)
}

// QueryFromSubcontrols queries the "from_subcontrols" edge of the MappedControl entity.
func (mc *MappedControl) QueryFromSubcontrols() *SubcontrolQuery {
	return NewMappedControlClient(mc.config).QueryFromSubcontrols(mc)
}

// QueryToSubcontrols queries the "to_subcontrols" edge of the MappedControl entity.
func (mc *MappedControl) QueryToSubcontrols() *SubcontrolQuery {
	return NewMappedControlClient(mc.config).QueryToSubcontrols(mc)
}

// Update returns a builder for updating this MappedControl.
// Note that you need to call MappedControl.Unwrap() before calling this method if this MappedControl
// was returned from a transaction, and the transaction was committed or rolled back.
func (mc *MappedControl) Update() *MappedControlUpdateOne {
	return NewMappedControlClient(mc.config).UpdateOne(mc)
}

// Unwrap unwraps the MappedControl entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (mc *MappedControl) Unwrap() *MappedControl {
	_tx, ok := mc.config.driver.(*txDriver)
	if !ok {
		panic("generated: MappedControl is not a transactional entity")
	}
	mc.config.driver = _tx.drv
	return mc
}

// String implements the fmt.Stringer.
func (mc *MappedControl) String() string {
	var builder strings.Builder
	builder.WriteString("MappedControl(")
	builder.WriteString(fmt.Sprintf("id=%v, ", mc.ID))
	builder.WriteString("created_at=")
	builder.WriteString(mc.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(mc.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(mc.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(mc.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(mc.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(mc.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", mc.Tags))
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(mc.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("mapping_type=")
	builder.WriteString(fmt.Sprintf("%v", mc.MappingType))
	builder.WriteString(", ")
	builder.WriteString("relation=")
	builder.WriteString(mc.Relation)
	builder.WriteString(", ")
	if v := mc.Confidence; v != nil {
		builder.WriteString("confidence=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	builder.WriteString("source=")
	builder.WriteString(fmt.Sprintf("%v", mc.Source))
	builder.WriteByte(')')
	return builder.String()
}

// NamedBlockedGroups returns the BlockedGroups named value or an error if the edge was not
// loaded in eager-loading with this name.
func (mc *MappedControl) NamedBlockedGroups(name string) ([]*Group, error) {
	if mc.Edges.namedBlockedGroups == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := mc.Edges.namedBlockedGroups[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (mc *MappedControl) appendNamedBlockedGroups(name string, edges ...*Group) {
	if mc.Edges.namedBlockedGroups == nil {
		mc.Edges.namedBlockedGroups = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		mc.Edges.namedBlockedGroups[name] = []*Group{}
	} else {
		mc.Edges.namedBlockedGroups[name] = append(mc.Edges.namedBlockedGroups[name], edges...)
	}
}

// NamedEditors returns the Editors named value or an error if the edge was not
// loaded in eager-loading with this name.
func (mc *MappedControl) NamedEditors(name string) ([]*Group, error) {
	if mc.Edges.namedEditors == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := mc.Edges.namedEditors[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (mc *MappedControl) appendNamedEditors(name string, edges ...*Group) {
	if mc.Edges.namedEditors == nil {
		mc.Edges.namedEditors = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		mc.Edges.namedEditors[name] = []*Group{}
	} else {
		mc.Edges.namedEditors[name] = append(mc.Edges.namedEditors[name], edges...)
	}
}

// NamedFromControls returns the FromControls named value or an error if the edge was not
// loaded in eager-loading with this name.
func (mc *MappedControl) NamedFromControls(name string) ([]*Control, error) {
	if mc.Edges.namedFromControls == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := mc.Edges.namedFromControls[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (mc *MappedControl) appendNamedFromControls(name string, edges ...*Control) {
	if mc.Edges.namedFromControls == nil {
		mc.Edges.namedFromControls = make(map[string][]*Control)
	}
	if len(edges) == 0 {
		mc.Edges.namedFromControls[name] = []*Control{}
	} else {
		mc.Edges.namedFromControls[name] = append(mc.Edges.namedFromControls[name], edges...)
	}
}

// NamedToControls returns the ToControls named value or an error if the edge was not
// loaded in eager-loading with this name.
func (mc *MappedControl) NamedToControls(name string) ([]*Control, error) {
	if mc.Edges.namedToControls == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := mc.Edges.namedToControls[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (mc *MappedControl) appendNamedToControls(name string, edges ...*Control) {
	if mc.Edges.namedToControls == nil {
		mc.Edges.namedToControls = make(map[string][]*Control)
	}
	if len(edges) == 0 {
		mc.Edges.namedToControls[name] = []*Control{}
	} else {
		mc.Edges.namedToControls[name] = append(mc.Edges.namedToControls[name], edges...)
	}
}

// NamedFromSubcontrols returns the FromSubcontrols named value or an error if the edge was not
// loaded in eager-loading with this name.
func (mc *MappedControl) NamedFromSubcontrols(name string) ([]*Subcontrol, error) {
	if mc.Edges.namedFromSubcontrols == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := mc.Edges.namedFromSubcontrols[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (mc *MappedControl) appendNamedFromSubcontrols(name string, edges ...*Subcontrol) {
	if mc.Edges.namedFromSubcontrols == nil {
		mc.Edges.namedFromSubcontrols = make(map[string][]*Subcontrol)
	}
	if len(edges) == 0 {
		mc.Edges.namedFromSubcontrols[name] = []*Subcontrol{}
	} else {
		mc.Edges.namedFromSubcontrols[name] = append(mc.Edges.namedFromSubcontrols[name], edges...)
	}
}

// NamedToSubcontrols returns the ToSubcontrols named value or an error if the edge was not
// loaded in eager-loading with this name.
func (mc *MappedControl) NamedToSubcontrols(name string) ([]*Subcontrol, error) {
	if mc.Edges.namedToSubcontrols == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := mc.Edges.namedToSubcontrols[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (mc *MappedControl) appendNamedToSubcontrols(name string, edges ...*Subcontrol) {
	if mc.Edges.namedToSubcontrols == nil {
		mc.Edges.namedToSubcontrols = make(map[string][]*Subcontrol)
	}
	if len(edges) == 0 {
		mc.Edges.namedToSubcontrols[name] = []*Subcontrol{}
	} else {
		mc.Edges.namedToSubcontrols[name] = append(mc.Edges.namedToSubcontrols[name], edges...)
	}
}

// MappedControls is a parsable slice of MappedControl.
type MappedControls []*MappedControl
