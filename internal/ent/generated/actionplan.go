// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/actionplan"
	"github.com/theopenlane/core/internal/ent/generated/group"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/enums"
)

// ActionPlan is the model entity for the ActionPlan schema.
type ActionPlan struct {
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
	// the name of the action_plan
	Name string `json:"name,omitempty"`
	// status of the action_plan, e.g. draft, published, archived, etc.
	Status enums.DocumentStatus `json:"status,omitempty"`
	// type of the action_plan, e.g. compliance, operational, health and safety, etc.
	ActionPlanType string `json:"action_plan_type,omitempty"`
	// details of the action_plan
	Details string `json:"details,omitempty"`
	// whether approval is required for edits to the action_plan
	ApprovalRequired bool `json:"approval_required,omitempty"`
	// the date the action_plan should be reviewed, calculated based on the review_frequency if not directly set
	ReviewDue time.Time `json:"review_due,omitempty"`
	// the frequency at which the action_plan should be reviewed, used to calculate the review_due date
	ReviewFrequency enums.Frequency `json:"review_frequency,omitempty"`
	// revision of the object as a semver (e.g. v1.0.0), by default any update will bump the patch version, unless the revision_bump field is set
	Revision string `json:"revision,omitempty"`
	// the organization id that owns the object
	OwnerID string `json:"owner_id,omitempty"`
	// due date of the action plan
	DueDate time.Time `json:"due_date,omitempty"`
	// priority of the action plan
	Priority enums.Priority `json:"priority,omitempty"`
	// source of the action plan
	Source string `json:"source,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the ActionPlanQuery when eager-loading is set.
	Edges                   ActionPlanEdges `json:"edges"`
	action_plan_approver    *string
	action_plan_delegate    *string
	subcontrol_action_plans *string
	selectValues            sql.SelectValues
}

// ActionPlanEdges holds the relations/edges for other nodes in the graph.
type ActionPlanEdges struct {
	// the group of users who are responsible for approving the action_plan
	Approver *Group `json:"approver,omitempty"`
	// temporary delegates for the action_plan, used for temporary approval
	Delegate *Group `json:"delegate,omitempty"`
	// Owner holds the value of the owner edge.
	Owner *Organization `json:"owner,omitempty"`
	// Risk holds the value of the risk edge.
	Risk []*Risk `json:"risk,omitempty"`
	// Control holds the value of the control edge.
	Control []*Control `json:"control,omitempty"`
	// User holds the value of the user edge.
	User []*User `json:"user,omitempty"`
	// Program holds the value of the program edge.
	Program []*Program `json:"program,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [7]bool
	// totalCount holds the count of the edges above.
	totalCount [7]map[string]int

	namedRisk    map[string][]*Risk
	namedControl map[string][]*Control
	namedUser    map[string][]*User
	namedProgram map[string][]*Program
}

// ApproverOrErr returns the Approver value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ActionPlanEdges) ApproverOrErr() (*Group, error) {
	if e.Approver != nil {
		return e.Approver, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: group.Label}
	}
	return nil, &NotLoadedError{edge: "approver"}
}

// DelegateOrErr returns the Delegate value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ActionPlanEdges) DelegateOrErr() (*Group, error) {
	if e.Delegate != nil {
		return e.Delegate, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: group.Label}
	}
	return nil, &NotLoadedError{edge: "delegate"}
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e ActionPlanEdges) OwnerOrErr() (*Organization, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// RiskOrErr returns the Risk value or an error if the edge
// was not loaded in eager-loading.
func (e ActionPlanEdges) RiskOrErr() ([]*Risk, error) {
	if e.loadedTypes[3] {
		return e.Risk, nil
	}
	return nil, &NotLoadedError{edge: "risk"}
}

// ControlOrErr returns the Control value or an error if the edge
// was not loaded in eager-loading.
func (e ActionPlanEdges) ControlOrErr() ([]*Control, error) {
	if e.loadedTypes[4] {
		return e.Control, nil
	}
	return nil, &NotLoadedError{edge: "control"}
}

// UserOrErr returns the User value or an error if the edge
// was not loaded in eager-loading.
func (e ActionPlanEdges) UserOrErr() ([]*User, error) {
	if e.loadedTypes[5] {
		return e.User, nil
	}
	return nil, &NotLoadedError{edge: "user"}
}

// ProgramOrErr returns the Program value or an error if the edge
// was not loaded in eager-loading.
func (e ActionPlanEdges) ProgramOrErr() ([]*Program, error) {
	if e.loadedTypes[6] {
		return e.Program, nil
	}
	return nil, &NotLoadedError{edge: "program"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*ActionPlan) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case actionplan.FieldTags:
			values[i] = new([]byte)
		case actionplan.FieldApprovalRequired:
			values[i] = new(sql.NullBool)
		case actionplan.FieldID, actionplan.FieldCreatedBy, actionplan.FieldUpdatedBy, actionplan.FieldDeletedBy, actionplan.FieldName, actionplan.FieldStatus, actionplan.FieldActionPlanType, actionplan.FieldDetails, actionplan.FieldReviewFrequency, actionplan.FieldRevision, actionplan.FieldOwnerID, actionplan.FieldPriority, actionplan.FieldSource:
			values[i] = new(sql.NullString)
		case actionplan.FieldCreatedAt, actionplan.FieldUpdatedAt, actionplan.FieldDeletedAt, actionplan.FieldReviewDue, actionplan.FieldDueDate:
			values[i] = new(sql.NullTime)
		case actionplan.ForeignKeys[0]: // action_plan_approver
			values[i] = new(sql.NullString)
		case actionplan.ForeignKeys[1]: // action_plan_delegate
			values[i] = new(sql.NullString)
		case actionplan.ForeignKeys[2]: // subcontrol_action_plans
			values[i] = new(sql.NullString)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the ActionPlan fields.
func (ap *ActionPlan) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case actionplan.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				ap.ID = value.String
			}
		case actionplan.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ap.CreatedAt = value.Time
			}
		case actionplan.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				ap.UpdatedAt = value.Time
			}
		case actionplan.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				ap.CreatedBy = value.String
			}
		case actionplan.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				ap.UpdatedBy = value.String
			}
		case actionplan.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				ap.DeletedAt = value.Time
			}
		case actionplan.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				ap.DeletedBy = value.String
			}
		case actionplan.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ap.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case actionplan.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				ap.Name = value.String
			}
		case actionplan.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				ap.Status = enums.DocumentStatus(value.String)
			}
		case actionplan.FieldActionPlanType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field action_plan_type", values[i])
			} else if value.Valid {
				ap.ActionPlanType = value.String
			}
		case actionplan.FieldDetails:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field details", values[i])
			} else if value.Valid {
				ap.Details = value.String
			}
		case actionplan.FieldApprovalRequired:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field approval_required", values[i])
			} else if value.Valid {
				ap.ApprovalRequired = value.Bool
			}
		case actionplan.FieldReviewDue:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field review_due", values[i])
			} else if value.Valid {
				ap.ReviewDue = value.Time
			}
		case actionplan.FieldReviewFrequency:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field review_frequency", values[i])
			} else if value.Valid {
				ap.ReviewFrequency = enums.Frequency(value.String)
			}
		case actionplan.FieldRevision:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field revision", values[i])
			} else if value.Valid {
				ap.Revision = value.String
			}
		case actionplan.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				ap.OwnerID = value.String
			}
		case actionplan.FieldDueDate:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field due_date", values[i])
			} else if value.Valid {
				ap.DueDate = value.Time
			}
		case actionplan.FieldPriority:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field priority", values[i])
			} else if value.Valid {
				ap.Priority = enums.Priority(value.String)
			}
		case actionplan.FieldSource:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field source", values[i])
			} else if value.Valid {
				ap.Source = value.String
			}
		case actionplan.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field action_plan_approver", values[i])
			} else if value.Valid {
				ap.action_plan_approver = new(string)
				*ap.action_plan_approver = value.String
			}
		case actionplan.ForeignKeys[1]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field action_plan_delegate", values[i])
			} else if value.Valid {
				ap.action_plan_delegate = new(string)
				*ap.action_plan_delegate = value.String
			}
		case actionplan.ForeignKeys[2]:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field subcontrol_action_plans", values[i])
			} else if value.Valid {
				ap.subcontrol_action_plans = new(string)
				*ap.subcontrol_action_plans = value.String
			}
		default:
			ap.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the ActionPlan.
// This includes values selected through modifiers, order, etc.
func (ap *ActionPlan) Value(name string) (ent.Value, error) {
	return ap.selectValues.Get(name)
}

// QueryApprover queries the "approver" edge of the ActionPlan entity.
func (ap *ActionPlan) QueryApprover() *GroupQuery {
	return NewActionPlanClient(ap.config).QueryApprover(ap)
}

// QueryDelegate queries the "delegate" edge of the ActionPlan entity.
func (ap *ActionPlan) QueryDelegate() *GroupQuery {
	return NewActionPlanClient(ap.config).QueryDelegate(ap)
}

// QueryOwner queries the "owner" edge of the ActionPlan entity.
func (ap *ActionPlan) QueryOwner() *OrganizationQuery {
	return NewActionPlanClient(ap.config).QueryOwner(ap)
}

// QueryRisk queries the "risk" edge of the ActionPlan entity.
func (ap *ActionPlan) QueryRisk() *RiskQuery {
	return NewActionPlanClient(ap.config).QueryRisk(ap)
}

// QueryControl queries the "control" edge of the ActionPlan entity.
func (ap *ActionPlan) QueryControl() *ControlQuery {
	return NewActionPlanClient(ap.config).QueryControl(ap)
}

// QueryUser queries the "user" edge of the ActionPlan entity.
func (ap *ActionPlan) QueryUser() *UserQuery {
	return NewActionPlanClient(ap.config).QueryUser(ap)
}

// QueryProgram queries the "program" edge of the ActionPlan entity.
func (ap *ActionPlan) QueryProgram() *ProgramQuery {
	return NewActionPlanClient(ap.config).QueryProgram(ap)
}

// Update returns a builder for updating this ActionPlan.
// Note that you need to call ActionPlan.Unwrap() before calling this method if this ActionPlan
// was returned from a transaction, and the transaction was committed or rolled back.
func (ap *ActionPlan) Update() *ActionPlanUpdateOne {
	return NewActionPlanClient(ap.config).UpdateOne(ap)
}

// Unwrap unwraps the ActionPlan entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ap *ActionPlan) Unwrap() *ActionPlan {
	_tx, ok := ap.config.driver.(*txDriver)
	if !ok {
		panic("generated: ActionPlan is not a transactional entity")
	}
	ap.config.driver = _tx.drv
	return ap
}

// String implements the fmt.Stringer.
func (ap *ActionPlan) String() string {
	var builder strings.Builder
	builder.WriteString("ActionPlan(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ap.ID))
	builder.WriteString("created_at=")
	builder.WriteString(ap.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(ap.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(ap.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(ap.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(ap.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(ap.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", ap.Tags))
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(ap.Name)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", ap.Status))
	builder.WriteString(", ")
	builder.WriteString("action_plan_type=")
	builder.WriteString(ap.ActionPlanType)
	builder.WriteString(", ")
	builder.WriteString("details=")
	builder.WriteString(ap.Details)
	builder.WriteString(", ")
	builder.WriteString("approval_required=")
	builder.WriteString(fmt.Sprintf("%v", ap.ApprovalRequired))
	builder.WriteString(", ")
	builder.WriteString("review_due=")
	builder.WriteString(ap.ReviewDue.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("review_frequency=")
	builder.WriteString(fmt.Sprintf("%v", ap.ReviewFrequency))
	builder.WriteString(", ")
	builder.WriteString("revision=")
	builder.WriteString(ap.Revision)
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(ap.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("due_date=")
	builder.WriteString(ap.DueDate.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("priority=")
	builder.WriteString(fmt.Sprintf("%v", ap.Priority))
	builder.WriteString(", ")
	builder.WriteString("source=")
	builder.WriteString(ap.Source)
	builder.WriteByte(')')
	return builder.String()
}

// NamedRisk returns the Risk named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ap *ActionPlan) NamedRisk(name string) ([]*Risk, error) {
	if ap.Edges.namedRisk == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ap.Edges.namedRisk[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ap *ActionPlan) appendNamedRisk(name string, edges ...*Risk) {
	if ap.Edges.namedRisk == nil {
		ap.Edges.namedRisk = make(map[string][]*Risk)
	}
	if len(edges) == 0 {
		ap.Edges.namedRisk[name] = []*Risk{}
	} else {
		ap.Edges.namedRisk[name] = append(ap.Edges.namedRisk[name], edges...)
	}
}

// NamedControl returns the Control named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ap *ActionPlan) NamedControl(name string) ([]*Control, error) {
	if ap.Edges.namedControl == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ap.Edges.namedControl[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ap *ActionPlan) appendNamedControl(name string, edges ...*Control) {
	if ap.Edges.namedControl == nil {
		ap.Edges.namedControl = make(map[string][]*Control)
	}
	if len(edges) == 0 {
		ap.Edges.namedControl[name] = []*Control{}
	} else {
		ap.Edges.namedControl[name] = append(ap.Edges.namedControl[name], edges...)
	}
}

// NamedUser returns the User named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ap *ActionPlan) NamedUser(name string) ([]*User, error) {
	if ap.Edges.namedUser == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ap.Edges.namedUser[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ap *ActionPlan) appendNamedUser(name string, edges ...*User) {
	if ap.Edges.namedUser == nil {
		ap.Edges.namedUser = make(map[string][]*User)
	}
	if len(edges) == 0 {
		ap.Edges.namedUser[name] = []*User{}
	} else {
		ap.Edges.namedUser[name] = append(ap.Edges.namedUser[name], edges...)
	}
}

// NamedProgram returns the Program named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ap *ActionPlan) NamedProgram(name string) ([]*Program, error) {
	if ap.Edges.namedProgram == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ap.Edges.namedProgram[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ap *ActionPlan) appendNamedProgram(name string, edges ...*Program) {
	if ap.Edges.namedProgram == nil {
		ap.Edges.namedProgram = make(map[string][]*Program)
	}
	if len(edges) == 0 {
		ap.Edges.namedProgram[name] = []*Program{}
	} else {
		ap.Edges.namedProgram[name] = append(ap.Edges.namedProgram[name], edges...)
	}
}

// ActionPlans is a parsable slice of ActionPlan.
type ActionPlans []*ActionPlan
