// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/risk"
	"github.com/theopenlane/core/pkg/enums"
)

// Risk is the model entity for the Risk schema.
type Risk struct {
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
	// the name of the risk
	Name string `json:"name,omitempty"`
	// status of the risk - open, mitigated, ongoing, in-progress, and archived.
	Status enums.RiskStatus `json:"status,omitempty"`
	// type of the risk, e.g. strategic, operational, financial, external, etc.
	RiskType string `json:"risk_type,omitempty"`
	// category of the risk, e.g. human resources, operations, IT, etc.
	Category string `json:"category,omitempty"`
	// impact of the risk -critical, high, medium, low
	Impact enums.RiskImpact `json:"impact,omitempty"`
	// likelihood of the risk occurring; unlikely, likely, highly likely
	Likelihood enums.RiskLikelihood `json:"likelihood,omitempty"`
	// score of the risk based on impact and likelihood (1-4 unlikely, 5-9 likely, 10-16 highly likely, 17-20 critical)
	Score int `json:"score,omitempty"`
	// mitigation for the risk
	Mitigation string `json:"mitigation,omitempty"`
	// details of the risk
	Details string `json:"details,omitempty"`
	// business costs associated with the risk
	BusinessCosts string `json:"business_costs,omitempty"`
	// the id of the group responsible for risk oversight
	StakeholderID string `json:"stakeholder_id,omitempty"`
	// the id of the group responsible for risk oversight on behalf of the stakeholder
	DelegateID string `json:"delegate_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the RiskQuery when eager-loading is set.
	Edges                   RiskEdges `json:"edges"`
	control_objective_risks *string
	selectValues            sql.SelectValues
}

// RiskEdges holds the relations/edges for other nodes in the graph.
type RiskEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Organization `json:"owner,omitempty"`
	// groups that are blocked from viewing or editing the risk
	BlockedGroups []*Group `json:"blocked_groups,omitempty"`
	// provides edit access to the risk to members of the group
	Editors []*Group `json:"editors,omitempty"`
	// provides view access to the risk to members of the group
	Viewers []*Group `json:"viewers,omitempty"`
	// Controls holds the value of the controls edge.
	Controls []*Control `json:"controls,omitempty"`
	// Subcontrols holds the value of the subcontrols edge.
	Subcontrols []*Subcontrol `json:"subcontrols,omitempty"`
	// Procedures holds the value of the procedures edge.
	Procedures []*Procedure `json:"procedures,omitempty"`
	// InternalPolicies holds the value of the internal_policies edge.
	InternalPolicies []*InternalPolicy `json:"internal_policies,omitempty"`
	// Programs holds the value of the programs edge.
	Programs []*Program `json:"programs,omitempty"`
	// ActionPlans holds the value of the action_plans edge.
	ActionPlans []*ActionPlan `json:"action_plans,omitempty"`
	// Tasks holds the value of the tasks edge.
	Tasks []*Task `json:"tasks,omitempty"`
	// Assets holds the value of the assets edge.
	Assets []*Asset `json:"assets,omitempty"`
	// Entities holds the value of the entities edge.
	Entities []*Entity `json:"entities,omitempty"`
	// Scans holds the value of the scans edge.
	Scans []*Scan `json:"scans,omitempty"`
	// the group of users who are responsible for risk oversight
	Stakeholder *Group `json:"stakeholder,omitempty"`
	// temporary delegates for the risk, used for temporary ownership
	Delegate *Group `json:"delegate,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [16]bool
	// totalCount holds the count of the edges above.
	totalCount [16]map[string]int

	namedBlockedGroups    map[string][]*Group
	namedEditors          map[string][]*Group
	namedViewers          map[string][]*Group
	namedControls         map[string][]*Control
	namedSubcontrols      map[string][]*Subcontrol
	namedProcedures       map[string][]*Procedure
	namedInternalPolicies map[string][]*InternalPolicy
	namedPrograms         map[string][]*Program
	namedActionPlans      map[string][]*ActionPlan
	namedTasks            map[string][]*Task
	namedAssets           map[string][]*Asset
	namedEntities         map[string][]*Entity
	namedScans            map[string][]*Scan
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RiskEdges) OwnerOrErr() (*Organization, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// BlockedGroupsOrErr returns the BlockedGroups value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) BlockedGroupsOrErr() ([]*Group, error) {
	if e.loadedTypes[1] {
		return e.BlockedGroups, nil
	}
	return nil, &NotLoadedError{edge: "blocked_groups"}
}

// EditorsOrErr returns the Editors value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) EditorsOrErr() ([]*Group, error) {
	if e.loadedTypes[2] {
		return e.Editors, nil
	}
	return nil, &NotLoadedError{edge: "editors"}
}

// ViewersOrErr returns the Viewers value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) ViewersOrErr() ([]*Group, error) {
	if e.loadedTypes[3] {
		return e.Viewers, nil
	}
	return nil, &NotLoadedError{edge: "viewers"}
}

// ControlsOrErr returns the Controls value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) ControlsOrErr() ([]*Control, error) {
	if e.loadedTypes[4] {
		return e.Controls, nil
	}
	return nil, &NotLoadedError{edge: "controls"}
}

// SubcontrolsOrErr returns the Subcontrols value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) SubcontrolsOrErr() ([]*Subcontrol, error) {
	if e.loadedTypes[5] {
		return e.Subcontrols, nil
	}
	return nil, &NotLoadedError{edge: "subcontrols"}
}

// ProceduresOrErr returns the Procedures value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) ProceduresOrErr() ([]*Procedure, error) {
	if e.loadedTypes[6] {
		return e.Procedures, nil
	}
	return nil, &NotLoadedError{edge: "procedures"}
}

// InternalPoliciesOrErr returns the InternalPolicies value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) InternalPoliciesOrErr() ([]*InternalPolicy, error) {
	if e.loadedTypes[7] {
		return e.InternalPolicies, nil
	}
	return nil, &NotLoadedError{edge: "internal_policies"}
}

// ProgramsOrErr returns the Programs value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) ProgramsOrErr() ([]*Program, error) {
	if e.loadedTypes[8] {
		return e.Programs, nil
	}
	return nil, &NotLoadedError{edge: "programs"}
}

// ActionPlansOrErr returns the ActionPlans value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) ActionPlansOrErr() ([]*ActionPlan, error) {
	if e.loadedTypes[9] {
		return e.ActionPlans, nil
	}
	return nil, &NotLoadedError{edge: "action_plans"}
}

// TasksOrErr returns the Tasks value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) TasksOrErr() ([]*Task, error) {
	if e.loadedTypes[10] {
		return e.Tasks, nil
	}
	return nil, &NotLoadedError{edge: "tasks"}
}

// AssetsOrErr returns the Assets value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) AssetsOrErr() ([]*Asset, error) {
	if e.loadedTypes[11] {
		return e.Assets, nil
	}
	return nil, &NotLoadedError{edge: "assets"}
}

// EntitiesOrErr returns the Entities value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) EntitiesOrErr() ([]*Entity, error) {
	if e.loadedTypes[12] {
		return e.Entities, nil
	}
	return nil, &NotLoadedError{edge: "entities"}
}

// ScansOrErr returns the Scans value or an error if the edge
// was not loaded in eager-loading.
func (e RiskEdges) ScansOrErr() ([]*Scan, error) {
	if e.loadedTypes[13] {
		return e.Scans, nil
	}
	return nil, &NotLoadedError{edge: "scans"}
}

// StakeholderOrErr returns the Stakeholder value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RiskEdges) StakeholderOrErr() (*Group, error) {
	if e.Stakeholder != nil {
		return e.Stakeholder, nil
	} else if e.loadedTypes[14] {
		return nil, &NotFoundError{label: group.Label}
	}
	return nil, &NotLoadedError{edge: "stakeholder"}
}

// DelegateOrErr returns the Delegate value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RiskEdges) DelegateOrErr() (*Group, error) {
	if e.Delegate != nil {
		return e.Delegate, nil
	} else if e.loadedTypes[15] {
		return nil, &NotFoundError{label: group.Label}
	}
	return nil, &NotLoadedError{edge: "delegate"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Risk) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case risk.FieldTags:
			values[i] = new([]byte)
		case risk.FieldScore:
			values[i] = new(sql.NullInt64)
		case risk.FieldID, risk.FieldCreatedBy, risk.FieldUpdatedBy, risk.FieldDeletedBy, risk.FieldDisplayID, risk.FieldOwnerID, risk.FieldName, risk.FieldStatus, risk.FieldRiskType, risk.FieldCategory, risk.FieldImpact, risk.FieldLikelihood, risk.FieldMitigation, risk.FieldDetails, risk.FieldBusinessCosts, risk.FieldStakeholderID, risk.FieldDelegateID:
			values[i] = new(sql.NullString)
		case risk.FieldCreatedAt, risk.FieldUpdatedAt, risk.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		case risk.ForeignKeys[0]: // control_objective_risks
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Risk fields.
func (r *Risk) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case risk.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				r.ID = value.String
			}
		case risk.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				r.CreatedAt = value.Time
			}
		case risk.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				r.UpdatedAt = value.Time
			}
		case risk.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				r.CreatedBy = value.String
			}
		case risk.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				r.UpdatedBy = value.String
			}
		case risk.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				r.DeletedAt = value.Time
			}
		case risk.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				r.DeletedBy = value.String
			}
		case risk.FieldDisplayID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field display_id", values[i])
			} else if value.Valid {
				r.DisplayID = value.String
			}
		case risk.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &r.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case risk.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				r.OwnerID = value.String
			}
		case risk.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				r.Name = value.String
			}
		case risk.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				r.Status = enums.RiskStatus(value.String)
			}
		case risk.FieldRiskType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field risk_type", values[i])
			} else if value.Valid {
				r.RiskType = value.String
			}
		case risk.FieldCategory:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field category", values[i])
			} else if value.Valid {
				r.Category = value.String
			}
		case risk.FieldImpact:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field impact", values[i])
			} else if value.Valid {
				r.Impact = enums.RiskImpact(value.String)
			}
		case risk.FieldLikelihood:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field likelihood", values[i])
			} else if value.Valid {
				r.Likelihood = enums.RiskLikelihood(value.String)
			}
		case risk.FieldScore:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field score", values[i])
			} else if value.Valid {
				r.Score = int(value.Int64)
			}
		case risk.FieldMitigation:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field mitigation", values[i])
			} else if value.Valid {
				r.Mitigation = value.String
			}
		case risk.FieldDetails:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field details", values[i])
			} else if value.Valid {
				r.Details = value.String
			}
		case risk.FieldBusinessCosts:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field business_costs", values[i])
			} else if value.Valid {
				r.BusinessCosts = value.String
			}
		case risk.FieldStakeholderID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field stakeholder_id", values[i])
			} else if value.Valid {
				r.StakeholderID = value.String
			}
		case risk.FieldDelegateID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field delegate_id", values[i])
			} else if value.Valid {
				r.DelegateID = value.String
			}
		case risk.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field control_objective_risks", values[i])
			} else if value.Valid {
				r.control_objective_risks = new(string)
				*r.control_objective_risks = value.String
			}
		default:
			r.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Risk.
// This includes values selected through modifiers, order, etc.
func (r *Risk) Value(name string) (ent.Value, error) {
	return r.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the Risk entity.
func (r *Risk) QueryOwner() *OrganizationQuery {
	return NewRiskClient(r.config).QueryOwner(r)
}

// QueryBlockedGroups queries the "blocked_groups" edge of the Risk entity.
func (r *Risk) QueryBlockedGroups() *GroupQuery {
	return NewRiskClient(r.config).QueryBlockedGroups(r)
}

// QueryEditors queries the "editors" edge of the Risk entity.
func (r *Risk) QueryEditors() *GroupQuery {
	return NewRiskClient(r.config).QueryEditors(r)
}

// QueryViewers queries the "viewers" edge of the Risk entity.
func (r *Risk) QueryViewers() *GroupQuery {
	return NewRiskClient(r.config).QueryViewers(r)
}

// QueryControls queries the "controls" edge of the Risk entity.
func (r *Risk) QueryControls() *ControlQuery {
	return NewRiskClient(r.config).QueryControls(r)
}

// QuerySubcontrols queries the "subcontrols" edge of the Risk entity.
func (r *Risk) QuerySubcontrols() *SubcontrolQuery {
	return NewRiskClient(r.config).QuerySubcontrols(r)
}

// QueryProcedures queries the "procedures" edge of the Risk entity.
func (r *Risk) QueryProcedures() *ProcedureQuery {
	return NewRiskClient(r.config).QueryProcedures(r)
}

// QueryInternalPolicies queries the "internal_policies" edge of the Risk entity.
func (r *Risk) QueryInternalPolicies() *InternalPolicyQuery {
	return NewRiskClient(r.config).QueryInternalPolicies(r)
}

// QueryPrograms queries the "programs" edge of the Risk entity.
func (r *Risk) QueryPrograms() *ProgramQuery {
	return NewRiskClient(r.config).QueryPrograms(r)
}

// QueryActionPlans queries the "action_plans" edge of the Risk entity.
func (r *Risk) QueryActionPlans() *ActionPlanQuery {
	return NewRiskClient(r.config).QueryActionPlans(r)
}

// QueryTasks queries the "tasks" edge of the Risk entity.
func (r *Risk) QueryTasks() *TaskQuery {
	return NewRiskClient(r.config).QueryTasks(r)
}

// QueryAssets queries the "assets" edge of the Risk entity.
func (r *Risk) QueryAssets() *AssetQuery {
	return NewRiskClient(r.config).QueryAssets(r)
}

// QueryEntities queries the "entities" edge of the Risk entity.
func (r *Risk) QueryEntities() *EntityQuery {
	return NewRiskClient(r.config).QueryEntities(r)
}

// QueryScans queries the "scans" edge of the Risk entity.
func (r *Risk) QueryScans() *ScanQuery {
	return NewRiskClient(r.config).QueryScans(r)
}

// QueryStakeholder queries the "stakeholder" edge of the Risk entity.
func (r *Risk) QueryStakeholder() *GroupQuery {
	return NewRiskClient(r.config).QueryStakeholder(r)
}

// QueryDelegate queries the "delegate" edge of the Risk entity.
func (r *Risk) QueryDelegate() *GroupQuery {
	return NewRiskClient(r.config).QueryDelegate(r)
}

// Update returns a builder for updating this Risk.
// Note that you need to call Risk.Unwrap() before calling this method if this Risk
// was returned from a transaction, and the transaction was committed or rolled back.
func (r *Risk) Update() *RiskUpdateOne {
	return NewRiskClient(r.config).UpdateOne(r)
}

// Unwrap unwraps the Risk entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (r *Risk) Unwrap() *Risk {
	_tx, ok := r.config.driver.(*txDriver)
	if !ok {
		panic("generated: Risk is not a transactional entity")
	}
	r.config.driver = _tx.drv
	return r
}

// String implements the fmt.Stringer.
func (r *Risk) String() string {
	var builder strings.Builder
	builder.WriteString("Risk(")
	builder.WriteString(fmt.Sprintf("id=%v, ", r.ID))
	builder.WriteString("created_at=")
	builder.WriteString(r.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(r.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(r.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(r.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(r.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(r.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("display_id=")
	builder.WriteString(r.DisplayID)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", r.Tags))
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(r.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(r.Name)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", r.Status))
	builder.WriteString(", ")
	builder.WriteString("risk_type=")
	builder.WriteString(r.RiskType)
	builder.WriteString(", ")
	builder.WriteString("category=")
	builder.WriteString(r.Category)
	builder.WriteString(", ")
	builder.WriteString("impact=")
	builder.WriteString(fmt.Sprintf("%v", r.Impact))
	builder.WriteString(", ")
	builder.WriteString("likelihood=")
	builder.WriteString(fmt.Sprintf("%v", r.Likelihood))
	builder.WriteString(", ")
	builder.WriteString("score=")
	builder.WriteString(fmt.Sprintf("%v", r.Score))
	builder.WriteString(", ")
	builder.WriteString("mitigation=")
	builder.WriteString(r.Mitigation)
	builder.WriteString(", ")
	builder.WriteString("details=")
	builder.WriteString(r.Details)
	builder.WriteString(", ")
	builder.WriteString("business_costs=")
	builder.WriteString(r.BusinessCosts)
	builder.WriteString(", ")
	builder.WriteString("stakeholder_id=")
	builder.WriteString(r.StakeholderID)
	builder.WriteString(", ")
	builder.WriteString("delegate_id=")
	builder.WriteString(r.DelegateID)
	builder.WriteByte(')')
	return builder.String()
}

// NamedBlockedGroups returns the BlockedGroups named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedBlockedGroups(name string) ([]*Group, error) {
	if r.Edges.namedBlockedGroups == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedBlockedGroups[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedBlockedGroups(name string, edges ...*Group) {
	if r.Edges.namedBlockedGroups == nil {
		r.Edges.namedBlockedGroups = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		r.Edges.namedBlockedGroups[name] = []*Group{}
	} else {
		r.Edges.namedBlockedGroups[name] = append(r.Edges.namedBlockedGroups[name], edges...)
	}
}

// NamedEditors returns the Editors named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedEditors(name string) ([]*Group, error) {
	if r.Edges.namedEditors == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedEditors[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedEditors(name string, edges ...*Group) {
	if r.Edges.namedEditors == nil {
		r.Edges.namedEditors = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		r.Edges.namedEditors[name] = []*Group{}
	} else {
		r.Edges.namedEditors[name] = append(r.Edges.namedEditors[name], edges...)
	}
}

// NamedViewers returns the Viewers named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedViewers(name string) ([]*Group, error) {
	if r.Edges.namedViewers == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedViewers[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedViewers(name string, edges ...*Group) {
	if r.Edges.namedViewers == nil {
		r.Edges.namedViewers = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		r.Edges.namedViewers[name] = []*Group{}
	} else {
		r.Edges.namedViewers[name] = append(r.Edges.namedViewers[name], edges...)
	}
}

// NamedControls returns the Controls named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedControls(name string) ([]*Control, error) {
	if r.Edges.namedControls == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedControls[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedControls(name string, edges ...*Control) {
	if r.Edges.namedControls == nil {
		r.Edges.namedControls = make(map[string][]*Control)
	}
	if len(edges) == 0 {
		r.Edges.namedControls[name] = []*Control{}
	} else {
		r.Edges.namedControls[name] = append(r.Edges.namedControls[name], edges...)
	}
}

// NamedSubcontrols returns the Subcontrols named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedSubcontrols(name string) ([]*Subcontrol, error) {
	if r.Edges.namedSubcontrols == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedSubcontrols[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedSubcontrols(name string, edges ...*Subcontrol) {
	if r.Edges.namedSubcontrols == nil {
		r.Edges.namedSubcontrols = make(map[string][]*Subcontrol)
	}
	if len(edges) == 0 {
		r.Edges.namedSubcontrols[name] = []*Subcontrol{}
	} else {
		r.Edges.namedSubcontrols[name] = append(r.Edges.namedSubcontrols[name], edges...)
	}
}

// NamedProcedures returns the Procedures named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedProcedures(name string) ([]*Procedure, error) {
	if r.Edges.namedProcedures == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedProcedures[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedProcedures(name string, edges ...*Procedure) {
	if r.Edges.namedProcedures == nil {
		r.Edges.namedProcedures = make(map[string][]*Procedure)
	}
	if len(edges) == 0 {
		r.Edges.namedProcedures[name] = []*Procedure{}
	} else {
		r.Edges.namedProcedures[name] = append(r.Edges.namedProcedures[name], edges...)
	}
}

// NamedInternalPolicies returns the InternalPolicies named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedInternalPolicies(name string) ([]*InternalPolicy, error) {
	if r.Edges.namedInternalPolicies == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedInternalPolicies[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedInternalPolicies(name string, edges ...*InternalPolicy) {
	if r.Edges.namedInternalPolicies == nil {
		r.Edges.namedInternalPolicies = make(map[string][]*InternalPolicy)
	}
	if len(edges) == 0 {
		r.Edges.namedInternalPolicies[name] = []*InternalPolicy{}
	} else {
		r.Edges.namedInternalPolicies[name] = append(r.Edges.namedInternalPolicies[name], edges...)
	}
}

// NamedPrograms returns the Programs named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedPrograms(name string) ([]*Program, error) {
	if r.Edges.namedPrograms == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedPrograms[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedPrograms(name string, edges ...*Program) {
	if r.Edges.namedPrograms == nil {
		r.Edges.namedPrograms = make(map[string][]*Program)
	}
	if len(edges) == 0 {
		r.Edges.namedPrograms[name] = []*Program{}
	} else {
		r.Edges.namedPrograms[name] = append(r.Edges.namedPrograms[name], edges...)
	}
}

// NamedActionPlans returns the ActionPlans named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedActionPlans(name string) ([]*ActionPlan, error) {
	if r.Edges.namedActionPlans == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedActionPlans[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedActionPlans(name string, edges ...*ActionPlan) {
	if r.Edges.namedActionPlans == nil {
		r.Edges.namedActionPlans = make(map[string][]*ActionPlan)
	}
	if len(edges) == 0 {
		r.Edges.namedActionPlans[name] = []*ActionPlan{}
	} else {
		r.Edges.namedActionPlans[name] = append(r.Edges.namedActionPlans[name], edges...)
	}
}

// NamedTasks returns the Tasks named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedTasks(name string) ([]*Task, error) {
	if r.Edges.namedTasks == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedTasks[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedTasks(name string, edges ...*Task) {
	if r.Edges.namedTasks == nil {
		r.Edges.namedTasks = make(map[string][]*Task)
	}
	if len(edges) == 0 {
		r.Edges.namedTasks[name] = []*Task{}
	} else {
		r.Edges.namedTasks[name] = append(r.Edges.namedTasks[name], edges...)
	}
}

// NamedAssets returns the Assets named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedAssets(name string) ([]*Asset, error) {
	if r.Edges.namedAssets == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedAssets[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedAssets(name string, edges ...*Asset) {
	if r.Edges.namedAssets == nil {
		r.Edges.namedAssets = make(map[string][]*Asset)
	}
	if len(edges) == 0 {
		r.Edges.namedAssets[name] = []*Asset{}
	} else {
		r.Edges.namedAssets[name] = append(r.Edges.namedAssets[name], edges...)
	}
}

// NamedEntities returns the Entities named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedEntities(name string) ([]*Entity, error) {
	if r.Edges.namedEntities == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedEntities[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedEntities(name string, edges ...*Entity) {
	if r.Edges.namedEntities == nil {
		r.Edges.namedEntities = make(map[string][]*Entity)
	}
	if len(edges) == 0 {
		r.Edges.namedEntities[name] = []*Entity{}
	} else {
		r.Edges.namedEntities[name] = append(r.Edges.namedEntities[name], edges...)
	}
}

// NamedScans returns the Scans named value or an error if the edge was not
// loaded in eager-loading with this name.
func (r *Risk) NamedScans(name string) ([]*Scan, error) {
	if r.Edges.namedScans == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := r.Edges.namedScans[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (r *Risk) appendNamedScans(name string, edges ...*Scan) {
	if r.Edges.namedScans == nil {
		r.Edges.namedScans = make(map[string][]*Scan)
	}
	if len(edges) == 0 {
		r.Edges.namedScans[name] = []*Scan{}
	} else {
		r.Edges.namedScans[name] = append(r.Edges.namedScans[name], edges...)
	}
}

// Risks is a parsable slice of Risk.
type Risks []*Risk
