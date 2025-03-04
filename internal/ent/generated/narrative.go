// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/narrative"
	"github.com/theopenlane/core/internal/ent/generated/organization"
)

// Narrative is the model entity for the Narrative schema.
type Narrative struct {
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
	// a shortened prefixed id field to use as a human readable identifier
	DisplayID string `json:"display_id,omitempty"`
	// tags associated with the object
	Tags []string `json:"tags,omitempty"`
	// the ID of the organization owner of the object
	OwnerID string `json:"owner_id,omitempty"`
	// the name of the narrative
	Name string `json:"name,omitempty"`
	// the description of the narrative
	Description string `json:"description,omitempty"`
	// which controls are satisfied by the narrative
	Satisfies string `json:"satisfies,omitempty"`
	// json data for the narrative document
	Details map[string]interface{} `json:"details,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the NarrativeQuery when eager-loading is set.
	Edges        NarrativeEdges `json:"edges"`
	selectValues sql.SelectValues
}

// NarrativeEdges holds the relations/edges for other nodes in the graph.
type NarrativeEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Organization `json:"owner,omitempty"`
	// groups that are blocked from viewing or editing the risk
	BlockedGroups []*Group `json:"blocked_groups,omitempty"`
	// provides edit access to the risk to members of the group
	Editors []*Group `json:"editors,omitempty"`
	// provides view access to the risk to members of the group
	Viewers []*Group `json:"viewers,omitempty"`
	// InternalPolicy holds the value of the internal_policy edge.
	InternalPolicy []*InternalPolicy `json:"internal_policy,omitempty"`
	// Control holds the value of the control edge.
	Control []*Control `json:"control,omitempty"`
	// Procedure holds the value of the procedure edge.
	Procedure []*Procedure `json:"procedure,omitempty"`
	// ControlObjective holds the value of the control_objective edge.
	ControlObjective []*ControlObjective `json:"control_objective,omitempty"`
	// Programs holds the value of the programs edge.
	Programs []*Program `json:"programs,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [9]bool
	// totalCount holds the count of the edges above.
	totalCount [9]map[string]int

	namedBlockedGroups    map[string][]*Group
	namedEditors          map[string][]*Group
	namedViewers          map[string][]*Group
	namedInternalPolicy   map[string][]*InternalPolicy
	namedControl          map[string][]*Control
	namedProcedure        map[string][]*Procedure
	namedControlObjective map[string][]*ControlObjective
	namedPrograms         map[string][]*Program
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e NarrativeEdges) OwnerOrErr() (*Organization, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// BlockedGroupsOrErr returns the BlockedGroups value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) BlockedGroupsOrErr() ([]*Group, error) {
	if e.loadedTypes[1] {
		return e.BlockedGroups, nil
	}
	return nil, &NotLoadedError{edge: "blocked_groups"}
}

// EditorsOrErr returns the Editors value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) EditorsOrErr() ([]*Group, error) {
	if e.loadedTypes[2] {
		return e.Editors, nil
	}
	return nil, &NotLoadedError{edge: "editors"}
}

// ViewersOrErr returns the Viewers value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) ViewersOrErr() ([]*Group, error) {
	if e.loadedTypes[3] {
		return e.Viewers, nil
	}
	return nil, &NotLoadedError{edge: "viewers"}
}

// InternalPolicyOrErr returns the InternalPolicy value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) InternalPolicyOrErr() ([]*InternalPolicy, error) {
	if e.loadedTypes[4] {
		return e.InternalPolicy, nil
	}
	return nil, &NotLoadedError{edge: "internal_policy"}
}

// ControlOrErr returns the Control value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) ControlOrErr() ([]*Control, error) {
	if e.loadedTypes[5] {
		return e.Control, nil
	}
	return nil, &NotLoadedError{edge: "control"}
}

// ProcedureOrErr returns the Procedure value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) ProcedureOrErr() ([]*Procedure, error) {
	if e.loadedTypes[6] {
		return e.Procedure, nil
	}
	return nil, &NotLoadedError{edge: "procedure"}
}

// ControlObjectiveOrErr returns the ControlObjective value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) ControlObjectiveOrErr() ([]*ControlObjective, error) {
	if e.loadedTypes[7] {
		return e.ControlObjective, nil
	}
	return nil, &NotLoadedError{edge: "control_objective"}
}

// ProgramsOrErr returns the Programs value or an error if the edge
// was not loaded in eager-loading.
func (e NarrativeEdges) ProgramsOrErr() ([]*Program, error) {
	if e.loadedTypes[8] {
		return e.Programs, nil
	}
	return nil, &NotLoadedError{edge: "programs"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Narrative) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case narrative.FieldTags, narrative.FieldDetails:
			values[i] = new([]byte)
		case narrative.FieldID, narrative.FieldCreatedBy, narrative.FieldUpdatedBy, narrative.FieldDeletedBy, narrative.FieldDisplayID, narrative.FieldOwnerID, narrative.FieldName, narrative.FieldDescription, narrative.FieldSatisfies:
			values[i] = new(sql.NullString)
		case narrative.FieldCreatedAt, narrative.FieldUpdatedAt, narrative.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Narrative fields.
func (n *Narrative) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case narrative.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				n.ID = value.String
			}
		case narrative.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				n.CreatedAt = value.Time
			}
		case narrative.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				n.UpdatedAt = value.Time
			}
		case narrative.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				n.CreatedBy = value.String
			}
		case narrative.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				n.UpdatedBy = value.String
			}
		case narrative.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				n.DeletedAt = value.Time
			}
		case narrative.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				n.DeletedBy = value.String
			}
		case narrative.FieldDisplayID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field display_id", values[i])
			} else if value.Valid {
				n.DisplayID = value.String
			}
		case narrative.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &n.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case narrative.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				n.OwnerID = value.String
			}
		case narrative.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				n.Name = value.String
			}
		case narrative.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				n.Description = value.String
			}
		case narrative.FieldSatisfies:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field satisfies", values[i])
			} else if value.Valid {
				n.Satisfies = value.String
			}
		case narrative.FieldDetails:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field details", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &n.Details); err != nil {
					return fmt.Errorf("unmarshal field details: %w", err)
				}
			}
		default:
			n.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Narrative.
// This includes values selected through modifiers, order, etc.
func (n *Narrative) Value(name string) (ent.Value, error) {
	return n.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the Narrative entity.
func (n *Narrative) QueryOwner() *OrganizationQuery {
	return NewNarrativeClient(n.config).QueryOwner(n)
}

// QueryBlockedGroups queries the "blocked_groups" edge of the Narrative entity.
func (n *Narrative) QueryBlockedGroups() *GroupQuery {
	return NewNarrativeClient(n.config).QueryBlockedGroups(n)
}

// QueryEditors queries the "editors" edge of the Narrative entity.
func (n *Narrative) QueryEditors() *GroupQuery {
	return NewNarrativeClient(n.config).QueryEditors(n)
}

// QueryViewers queries the "viewers" edge of the Narrative entity.
func (n *Narrative) QueryViewers() *GroupQuery {
	return NewNarrativeClient(n.config).QueryViewers(n)
}

// QueryInternalPolicy queries the "internal_policy" edge of the Narrative entity.
func (n *Narrative) QueryInternalPolicy() *InternalPolicyQuery {
	return NewNarrativeClient(n.config).QueryInternalPolicy(n)
}

// QueryControl queries the "control" edge of the Narrative entity.
func (n *Narrative) QueryControl() *ControlQuery {
	return NewNarrativeClient(n.config).QueryControl(n)
}

// QueryProcedure queries the "procedure" edge of the Narrative entity.
func (n *Narrative) QueryProcedure() *ProcedureQuery {
	return NewNarrativeClient(n.config).QueryProcedure(n)
}

// QueryControlObjective queries the "control_objective" edge of the Narrative entity.
func (n *Narrative) QueryControlObjective() *ControlObjectiveQuery {
	return NewNarrativeClient(n.config).QueryControlObjective(n)
}

// QueryPrograms queries the "programs" edge of the Narrative entity.
func (n *Narrative) QueryPrograms() *ProgramQuery {
	return NewNarrativeClient(n.config).QueryPrograms(n)
}

// Update returns a builder for updating this Narrative.
// Note that you need to call Narrative.Unwrap() before calling this method if this Narrative
// was returned from a transaction, and the transaction was committed or rolled back.
func (n *Narrative) Update() *NarrativeUpdateOne {
	return NewNarrativeClient(n.config).UpdateOne(n)
}

// Unwrap unwraps the Narrative entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (n *Narrative) Unwrap() *Narrative {
	_tx, ok := n.config.driver.(*txDriver)
	if !ok {
		panic("generated: Narrative is not a transactional entity")
	}
	n.config.driver = _tx.drv
	return n
}

// String implements the fmt.Stringer.
func (n *Narrative) String() string {
	var builder strings.Builder
	builder.WriteString("Narrative(")
	builder.WriteString(fmt.Sprintf("id=%v, ", n.ID))
	builder.WriteString("created_at=")
	builder.WriteString(n.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(n.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(n.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(n.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(n.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(n.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("display_id=")
	builder.WriteString(n.DisplayID)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", n.Tags))
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(n.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(n.Name)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(n.Description)
	builder.WriteString(", ")
	builder.WriteString("satisfies=")
	builder.WriteString(n.Satisfies)
	builder.WriteString(", ")
	builder.WriteString("details=")
	builder.WriteString(fmt.Sprintf("%v", n.Details))
	builder.WriteByte(')')
	return builder.String()
}

// NamedBlockedGroups returns the BlockedGroups named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedBlockedGroups(name string) ([]*Group, error) {
	if n.Edges.namedBlockedGroups == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedBlockedGroups[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedBlockedGroups(name string, edges ...*Group) {
	if n.Edges.namedBlockedGroups == nil {
		n.Edges.namedBlockedGroups = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		n.Edges.namedBlockedGroups[name] = []*Group{}
	} else {
		n.Edges.namedBlockedGroups[name] = append(n.Edges.namedBlockedGroups[name], edges...)
	}
}

// NamedEditors returns the Editors named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedEditors(name string) ([]*Group, error) {
	if n.Edges.namedEditors == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedEditors[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedEditors(name string, edges ...*Group) {
	if n.Edges.namedEditors == nil {
		n.Edges.namedEditors = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		n.Edges.namedEditors[name] = []*Group{}
	} else {
		n.Edges.namedEditors[name] = append(n.Edges.namedEditors[name], edges...)
	}
}

// NamedViewers returns the Viewers named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedViewers(name string) ([]*Group, error) {
	if n.Edges.namedViewers == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedViewers[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedViewers(name string, edges ...*Group) {
	if n.Edges.namedViewers == nil {
		n.Edges.namedViewers = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		n.Edges.namedViewers[name] = []*Group{}
	} else {
		n.Edges.namedViewers[name] = append(n.Edges.namedViewers[name], edges...)
	}
}

// NamedInternalPolicy returns the InternalPolicy named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedInternalPolicy(name string) ([]*InternalPolicy, error) {
	if n.Edges.namedInternalPolicy == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedInternalPolicy[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedInternalPolicy(name string, edges ...*InternalPolicy) {
	if n.Edges.namedInternalPolicy == nil {
		n.Edges.namedInternalPolicy = make(map[string][]*InternalPolicy)
	}
	if len(edges) == 0 {
		n.Edges.namedInternalPolicy[name] = []*InternalPolicy{}
	} else {
		n.Edges.namedInternalPolicy[name] = append(n.Edges.namedInternalPolicy[name], edges...)
	}
}

// NamedControl returns the Control named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedControl(name string) ([]*Control, error) {
	if n.Edges.namedControl == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedControl[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedControl(name string, edges ...*Control) {
	if n.Edges.namedControl == nil {
		n.Edges.namedControl = make(map[string][]*Control)
	}
	if len(edges) == 0 {
		n.Edges.namedControl[name] = []*Control{}
	} else {
		n.Edges.namedControl[name] = append(n.Edges.namedControl[name], edges...)
	}
}

// NamedProcedure returns the Procedure named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedProcedure(name string) ([]*Procedure, error) {
	if n.Edges.namedProcedure == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedProcedure[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedProcedure(name string, edges ...*Procedure) {
	if n.Edges.namedProcedure == nil {
		n.Edges.namedProcedure = make(map[string][]*Procedure)
	}
	if len(edges) == 0 {
		n.Edges.namedProcedure[name] = []*Procedure{}
	} else {
		n.Edges.namedProcedure[name] = append(n.Edges.namedProcedure[name], edges...)
	}
}

// NamedControlObjective returns the ControlObjective named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedControlObjective(name string) ([]*ControlObjective, error) {
	if n.Edges.namedControlObjective == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedControlObjective[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedControlObjective(name string, edges ...*ControlObjective) {
	if n.Edges.namedControlObjective == nil {
		n.Edges.namedControlObjective = make(map[string][]*ControlObjective)
	}
	if len(edges) == 0 {
		n.Edges.namedControlObjective[name] = []*ControlObjective{}
	} else {
		n.Edges.namedControlObjective[name] = append(n.Edges.namedControlObjective[name], edges...)
	}
}

// NamedPrograms returns the Programs named value or an error if the edge was not
// loaded in eager-loading with this name.
func (n *Narrative) NamedPrograms(name string) ([]*Program, error) {
	if n.Edges.namedPrograms == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := n.Edges.namedPrograms[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (n *Narrative) appendNamedPrograms(name string, edges ...*Program) {
	if n.Edges.namedPrograms == nil {
		n.Edges.namedPrograms = make(map[string][]*Program)
	}
	if len(edges) == 0 {
		n.Edges.namedPrograms[name] = []*Program{}
	} else {
		n.Edges.namedPrograms[name] = append(n.Edges.namedPrograms[name], edges...)
	}
}

// Narratives is a parsable slice of Narrative.
type Narratives []*Narrative
