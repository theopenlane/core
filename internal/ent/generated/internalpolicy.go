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
	"github.com/theopenlane/core/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/enums"
)

// InternalPolicy is the model entity for the InternalPolicy schema.
type InternalPolicy struct {
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
	// revision of the object as a semver (e.g. v1.0.0), by default any update will bump the patch version, unless the revision_bump field is set
	Revision string `json:"revision,omitempty"`
	// the organization id that owns the object
	OwnerID string `json:"owner_id,omitempty"`
	// the name of the policy
	Name string `json:"name,omitempty"`
	// status of the policy, e.g. draft, published, archived, etc.
	Status enums.DocumentStatus `json:"status,omitempty"`
	// type of the policy, e.g. compliance, operational, health and safety, etc.
	PolicyType string `json:"policy_type,omitempty"`
	// details of the policy
	Details string `json:"details,omitempty"`
	// whether approval is required for edits to the policy
	ApprovalRequired bool `json:"approval_required,omitempty"`
	// the date the policy should be reviewed, calculated based on the review_frequency if not directly set
	ReviewDue time.Time `json:"review_due,omitempty"`
	// the frequency at which the policy should be reviewed, used to calculate the review_due date
	ReviewFrequency enums.Frequency `json:"review_frequency,omitempty"`
	// the id of the group responsible for approving the policy
	ApproverID string `json:"approver_id,omitempty"`
	// the id of the group responsible for approving the policy
	DelegateID string `json:"delegate_id,omitempty"`
	// Summary holds the value of the "summary" field.
	Summary string `json:"summary,omitempty"`
	// auto-generated tag suggestions for the policy
	TagSuggestions []string `json:"tag_suggestions,omitempty"`
	// tag suggestions dismissed by the user for the policy
	DismissedTagSuggestions []string `json:"dismissed_tag_suggestions,omitempty"`
	// proposed controls referenced in the policy
	ControlSuggestions []string `json:"control_suggestions,omitempty"`
	// control suggestions dismissed by the user for the policy
	DismissedControlSuggestions []string `json:"dismissed_control_suggestions,omitempty"`
	// suggested improvements for the policy
	ImprovementSuggestions []string `json:"improvement_suggestions,omitempty"`
	// improvement suggestions dismissed by the user for the policy
	DismissedImprovementSuggestions []string `json:"dismissed_improvement_suggestions,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the InternalPolicyQuery when eager-loading is set.
	Edges        InternalPolicyEdges `json:"edges"`
	selectValues sql.SelectValues
}

// InternalPolicyEdges holds the relations/edges for other nodes in the graph.
type InternalPolicyEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Organization `json:"owner,omitempty"`
	// groups that are blocked from viewing or editing the risk
	BlockedGroups []*Group `json:"blocked_groups,omitempty"`
	// provides edit access to the risk to members of the group
	Editors []*Group `json:"editors,omitempty"`
	// the group of users who are responsible for approving the policy
	Approver *Group `json:"approver,omitempty"`
	// temporary delegates for the policy, used for temporary approval
	Delegate *Group `json:"delegate,omitempty"`
	// ControlObjectives holds the value of the control_objectives edge.
	ControlObjectives []*ControlObjective `json:"control_objectives,omitempty"`
	// Controls holds the value of the controls edge.
	Controls []*Control `json:"controls,omitempty"`
	// Subcontrols holds the value of the subcontrols edge.
	Subcontrols []*Subcontrol `json:"subcontrols,omitempty"`
	// Procedures holds the value of the procedures edge.
	Procedures []*Procedure `json:"procedures,omitempty"`
	// Narratives holds the value of the narratives edge.
	Narratives []*Narrative `json:"narratives,omitempty"`
	// Tasks holds the value of the tasks edge.
	Tasks []*Task `json:"tasks,omitempty"`
	// Risks holds the value of the risks edge.
	Risks []*Risk `json:"risks,omitempty"`
	// Programs holds the value of the programs edge.
	Programs []*Program `json:"programs,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [13]bool
	// totalCount holds the count of the edges above.
	totalCount [13]map[string]int

	namedBlockedGroups     map[string][]*Group
	namedEditors           map[string][]*Group
	namedControlObjectives map[string][]*ControlObjective
	namedControls          map[string][]*Control
	namedSubcontrols       map[string][]*Subcontrol
	namedProcedures        map[string][]*Procedure
	namedNarratives        map[string][]*Narrative
	namedTasks             map[string][]*Task
	namedRisks             map[string][]*Risk
	namedPrograms          map[string][]*Program
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e InternalPolicyEdges) OwnerOrErr() (*Organization, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// BlockedGroupsOrErr returns the BlockedGroups value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) BlockedGroupsOrErr() ([]*Group, error) {
	if e.loadedTypes[1] {
		return e.BlockedGroups, nil
	}
	return nil, &NotLoadedError{edge: "blocked_groups"}
}

// EditorsOrErr returns the Editors value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) EditorsOrErr() ([]*Group, error) {
	if e.loadedTypes[2] {
		return e.Editors, nil
	}
	return nil, &NotLoadedError{edge: "editors"}
}

// ApproverOrErr returns the Approver value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e InternalPolicyEdges) ApproverOrErr() (*Group, error) {
	if e.Approver != nil {
		return e.Approver, nil
	} else if e.loadedTypes[3] {
		return nil, &NotFoundError{label: group.Label}
	}
	return nil, &NotLoadedError{edge: "approver"}
}

// DelegateOrErr returns the Delegate value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e InternalPolicyEdges) DelegateOrErr() (*Group, error) {
	if e.Delegate != nil {
		return e.Delegate, nil
	} else if e.loadedTypes[4] {
		return nil, &NotFoundError{label: group.Label}
	}
	return nil, &NotLoadedError{edge: "delegate"}
}

// ControlObjectivesOrErr returns the ControlObjectives value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) ControlObjectivesOrErr() ([]*ControlObjective, error) {
	if e.loadedTypes[5] {
		return e.ControlObjectives, nil
	}
	return nil, &NotLoadedError{edge: "control_objectives"}
}

// ControlsOrErr returns the Controls value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) ControlsOrErr() ([]*Control, error) {
	if e.loadedTypes[6] {
		return e.Controls, nil
	}
	return nil, &NotLoadedError{edge: "controls"}
}

// SubcontrolsOrErr returns the Subcontrols value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) SubcontrolsOrErr() ([]*Subcontrol, error) {
	if e.loadedTypes[7] {
		return e.Subcontrols, nil
	}
	return nil, &NotLoadedError{edge: "subcontrols"}
}

// ProceduresOrErr returns the Procedures value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) ProceduresOrErr() ([]*Procedure, error) {
	if e.loadedTypes[8] {
		return e.Procedures, nil
	}
	return nil, &NotLoadedError{edge: "procedures"}
}

// NarrativesOrErr returns the Narratives value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) NarrativesOrErr() ([]*Narrative, error) {
	if e.loadedTypes[9] {
		return e.Narratives, nil
	}
	return nil, &NotLoadedError{edge: "narratives"}
}

// TasksOrErr returns the Tasks value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) TasksOrErr() ([]*Task, error) {
	if e.loadedTypes[10] {
		return e.Tasks, nil
	}
	return nil, &NotLoadedError{edge: "tasks"}
}

// RisksOrErr returns the Risks value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) RisksOrErr() ([]*Risk, error) {
	if e.loadedTypes[11] {
		return e.Risks, nil
	}
	return nil, &NotLoadedError{edge: "risks"}
}

// ProgramsOrErr returns the Programs value or an error if the edge
// was not loaded in eager-loading.
func (e InternalPolicyEdges) ProgramsOrErr() ([]*Program, error) {
	if e.loadedTypes[12] {
		return e.Programs, nil
	}
	return nil, &NotLoadedError{edge: "programs"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*InternalPolicy) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case internalpolicy.FieldTags, internalpolicy.FieldTagSuggestions, internalpolicy.FieldDismissedTagSuggestions, internalpolicy.FieldControlSuggestions, internalpolicy.FieldDismissedControlSuggestions, internalpolicy.FieldImprovementSuggestions, internalpolicy.FieldDismissedImprovementSuggestions:
			values[i] = new([]byte)
		case internalpolicy.FieldApprovalRequired:
			values[i] = new(sql.NullBool)
		case internalpolicy.FieldID, internalpolicy.FieldCreatedBy, internalpolicy.FieldUpdatedBy, internalpolicy.FieldDeletedBy, internalpolicy.FieldDisplayID, internalpolicy.FieldRevision, internalpolicy.FieldOwnerID, internalpolicy.FieldName, internalpolicy.FieldStatus, internalpolicy.FieldPolicyType, internalpolicy.FieldDetails, internalpolicy.FieldReviewFrequency, internalpolicy.FieldApproverID, internalpolicy.FieldDelegateID, internalpolicy.FieldSummary:
			values[i] = new(sql.NullString)
		case internalpolicy.FieldCreatedAt, internalpolicy.FieldUpdatedAt, internalpolicy.FieldDeletedAt, internalpolicy.FieldReviewDue:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the InternalPolicy fields.
func (ip *InternalPolicy) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case internalpolicy.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				ip.ID = value.String
			}
		case internalpolicy.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ip.CreatedAt = value.Time
			}
		case internalpolicy.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				ip.UpdatedAt = value.Time
			}
		case internalpolicy.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				ip.CreatedBy = value.String
			}
		case internalpolicy.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				ip.UpdatedBy = value.String
			}
		case internalpolicy.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				ip.DeletedAt = value.Time
			}
		case internalpolicy.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				ip.DeletedBy = value.String
			}
		case internalpolicy.FieldDisplayID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field display_id", values[i])
			} else if value.Valid {
				ip.DisplayID = value.String
			}
		case internalpolicy.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ip.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case internalpolicy.FieldRevision:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field revision", values[i])
			} else if value.Valid {
				ip.Revision = value.String
			}
		case internalpolicy.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				ip.OwnerID = value.String
			}
		case internalpolicy.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				ip.Name = value.String
			}
		case internalpolicy.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				ip.Status = enums.DocumentStatus(value.String)
			}
		case internalpolicy.FieldPolicyType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field policy_type", values[i])
			} else if value.Valid {
				ip.PolicyType = value.String
			}
		case internalpolicy.FieldDetails:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field details", values[i])
			} else if value.Valid {
				ip.Details = value.String
			}
		case internalpolicy.FieldApprovalRequired:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field approval_required", values[i])
			} else if value.Valid {
				ip.ApprovalRequired = value.Bool
			}
		case internalpolicy.FieldReviewDue:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field review_due", values[i])
			} else if value.Valid {
				ip.ReviewDue = value.Time
			}
		case internalpolicy.FieldReviewFrequency:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field review_frequency", values[i])
			} else if value.Valid {
				ip.ReviewFrequency = enums.Frequency(value.String)
			}
		case internalpolicy.FieldApproverID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field approver_id", values[i])
			} else if value.Valid {
				ip.ApproverID = value.String
			}
		case internalpolicy.FieldDelegateID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field delegate_id", values[i])
			} else if value.Valid {
				ip.DelegateID = value.String
			}
		case internalpolicy.FieldSummary:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field summary", values[i])
			} else if value.Valid {
				ip.Summary = value.String
			}
		case internalpolicy.FieldTagSuggestions:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tag_suggestions", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ip.TagSuggestions); err != nil {
					return fmt.Errorf("unmarshal field tag_suggestions: %w", err)
				}
			}
		case internalpolicy.FieldDismissedTagSuggestions:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field dismissed_tag_suggestions", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ip.DismissedTagSuggestions); err != nil {
					return fmt.Errorf("unmarshal field dismissed_tag_suggestions: %w", err)
				}
			}
		case internalpolicy.FieldControlSuggestions:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field control_suggestions", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ip.ControlSuggestions); err != nil {
					return fmt.Errorf("unmarshal field control_suggestions: %w", err)
				}
			}
		case internalpolicy.FieldDismissedControlSuggestions:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field dismissed_control_suggestions", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ip.DismissedControlSuggestions); err != nil {
					return fmt.Errorf("unmarshal field dismissed_control_suggestions: %w", err)
				}
			}
		case internalpolicy.FieldImprovementSuggestions:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field improvement_suggestions", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ip.ImprovementSuggestions); err != nil {
					return fmt.Errorf("unmarshal field improvement_suggestions: %w", err)
				}
			}
		case internalpolicy.FieldDismissedImprovementSuggestions:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field dismissed_improvement_suggestions", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ip.DismissedImprovementSuggestions); err != nil {
					return fmt.Errorf("unmarshal field dismissed_improvement_suggestions: %w", err)
				}
			}
		default:
			ip.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the InternalPolicy.
// This includes values selected through modifiers, order, etc.
func (ip *InternalPolicy) Value(name string) (ent.Value, error) {
	return ip.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryOwner() *OrganizationQuery {
	return NewInternalPolicyClient(ip.config).QueryOwner(ip)
}

// QueryBlockedGroups queries the "blocked_groups" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryBlockedGroups() *GroupQuery {
	return NewInternalPolicyClient(ip.config).QueryBlockedGroups(ip)
}

// QueryEditors queries the "editors" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryEditors() *GroupQuery {
	return NewInternalPolicyClient(ip.config).QueryEditors(ip)
}

// QueryApprover queries the "approver" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryApprover() *GroupQuery {
	return NewInternalPolicyClient(ip.config).QueryApprover(ip)
}

// QueryDelegate queries the "delegate" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryDelegate() *GroupQuery {
	return NewInternalPolicyClient(ip.config).QueryDelegate(ip)
}

// QueryControlObjectives queries the "control_objectives" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryControlObjectives() *ControlObjectiveQuery {
	return NewInternalPolicyClient(ip.config).QueryControlObjectives(ip)
}

// QueryControls queries the "controls" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryControls() *ControlQuery {
	return NewInternalPolicyClient(ip.config).QueryControls(ip)
}

// QuerySubcontrols queries the "subcontrols" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QuerySubcontrols() *SubcontrolQuery {
	return NewInternalPolicyClient(ip.config).QuerySubcontrols(ip)
}

// QueryProcedures queries the "procedures" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryProcedures() *ProcedureQuery {
	return NewInternalPolicyClient(ip.config).QueryProcedures(ip)
}

// QueryNarratives queries the "narratives" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryNarratives() *NarrativeQuery {
	return NewInternalPolicyClient(ip.config).QueryNarratives(ip)
}

// QueryTasks queries the "tasks" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryTasks() *TaskQuery {
	return NewInternalPolicyClient(ip.config).QueryTasks(ip)
}

// QueryRisks queries the "risks" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryRisks() *RiskQuery {
	return NewInternalPolicyClient(ip.config).QueryRisks(ip)
}

// QueryPrograms queries the "programs" edge of the InternalPolicy entity.
func (ip *InternalPolicy) QueryPrograms() *ProgramQuery {
	return NewInternalPolicyClient(ip.config).QueryPrograms(ip)
}

// Update returns a builder for updating this InternalPolicy.
// Note that you need to call InternalPolicy.Unwrap() before calling this method if this InternalPolicy
// was returned from a transaction, and the transaction was committed or rolled back.
func (ip *InternalPolicy) Update() *InternalPolicyUpdateOne {
	return NewInternalPolicyClient(ip.config).UpdateOne(ip)
}

// Unwrap unwraps the InternalPolicy entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ip *InternalPolicy) Unwrap() *InternalPolicy {
	_tx, ok := ip.config.driver.(*txDriver)
	if !ok {
		panic("generated: InternalPolicy is not a transactional entity")
	}
	ip.config.driver = _tx.drv
	return ip
}

// String implements the fmt.Stringer.
func (ip *InternalPolicy) String() string {
	var builder strings.Builder
	builder.WriteString("InternalPolicy(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ip.ID))
	builder.WriteString("created_at=")
	builder.WriteString(ip.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(ip.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(ip.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(ip.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(ip.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(ip.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("display_id=")
	builder.WriteString(ip.DisplayID)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", ip.Tags))
	builder.WriteString(", ")
	builder.WriteString("revision=")
	builder.WriteString(ip.Revision)
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(ip.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(ip.Name)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", ip.Status))
	builder.WriteString(", ")
	builder.WriteString("policy_type=")
	builder.WriteString(ip.PolicyType)
	builder.WriteString(", ")
	builder.WriteString("details=")
	builder.WriteString(ip.Details)
	builder.WriteString(", ")
	builder.WriteString("approval_required=")
	builder.WriteString(fmt.Sprintf("%v", ip.ApprovalRequired))
	builder.WriteString(", ")
	builder.WriteString("review_due=")
	builder.WriteString(ip.ReviewDue.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("review_frequency=")
	builder.WriteString(fmt.Sprintf("%v", ip.ReviewFrequency))
	builder.WriteString(", ")
	builder.WriteString("approver_id=")
	builder.WriteString(ip.ApproverID)
	builder.WriteString(", ")
	builder.WriteString("delegate_id=")
	builder.WriteString(ip.DelegateID)
	builder.WriteString(", ")
	builder.WriteString("summary=")
	builder.WriteString(ip.Summary)
	builder.WriteString(", ")
	builder.WriteString("tag_suggestions=")
	builder.WriteString(fmt.Sprintf("%v", ip.TagSuggestions))
	builder.WriteString(", ")
	builder.WriteString("dismissed_tag_suggestions=")
	builder.WriteString(fmt.Sprintf("%v", ip.DismissedTagSuggestions))
	builder.WriteString(", ")
	builder.WriteString("control_suggestions=")
	builder.WriteString(fmt.Sprintf("%v", ip.ControlSuggestions))
	builder.WriteString(", ")
	builder.WriteString("dismissed_control_suggestions=")
	builder.WriteString(fmt.Sprintf("%v", ip.DismissedControlSuggestions))
	builder.WriteString(", ")
	builder.WriteString("improvement_suggestions=")
	builder.WriteString(fmt.Sprintf("%v", ip.ImprovementSuggestions))
	builder.WriteString(", ")
	builder.WriteString("dismissed_improvement_suggestions=")
	builder.WriteString(fmt.Sprintf("%v", ip.DismissedImprovementSuggestions))
	builder.WriteByte(')')
	return builder.String()
}

// NamedBlockedGroups returns the BlockedGroups named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedBlockedGroups(name string) ([]*Group, error) {
	if ip.Edges.namedBlockedGroups == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedBlockedGroups[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedBlockedGroups(name string, edges ...*Group) {
	if ip.Edges.namedBlockedGroups == nil {
		ip.Edges.namedBlockedGroups = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		ip.Edges.namedBlockedGroups[name] = []*Group{}
	} else {
		ip.Edges.namedBlockedGroups[name] = append(ip.Edges.namedBlockedGroups[name], edges...)
	}
}

// NamedEditors returns the Editors named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedEditors(name string) ([]*Group, error) {
	if ip.Edges.namedEditors == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedEditors[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedEditors(name string, edges ...*Group) {
	if ip.Edges.namedEditors == nil {
		ip.Edges.namedEditors = make(map[string][]*Group)
	}
	if len(edges) == 0 {
		ip.Edges.namedEditors[name] = []*Group{}
	} else {
		ip.Edges.namedEditors[name] = append(ip.Edges.namedEditors[name], edges...)
	}
}

// NamedControlObjectives returns the ControlObjectives named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedControlObjectives(name string) ([]*ControlObjective, error) {
	if ip.Edges.namedControlObjectives == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedControlObjectives[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedControlObjectives(name string, edges ...*ControlObjective) {
	if ip.Edges.namedControlObjectives == nil {
		ip.Edges.namedControlObjectives = make(map[string][]*ControlObjective)
	}
	if len(edges) == 0 {
		ip.Edges.namedControlObjectives[name] = []*ControlObjective{}
	} else {
		ip.Edges.namedControlObjectives[name] = append(ip.Edges.namedControlObjectives[name], edges...)
	}
}

// NamedControls returns the Controls named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedControls(name string) ([]*Control, error) {
	if ip.Edges.namedControls == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedControls[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedControls(name string, edges ...*Control) {
	if ip.Edges.namedControls == nil {
		ip.Edges.namedControls = make(map[string][]*Control)
	}
	if len(edges) == 0 {
		ip.Edges.namedControls[name] = []*Control{}
	} else {
		ip.Edges.namedControls[name] = append(ip.Edges.namedControls[name], edges...)
	}
}

// NamedSubcontrols returns the Subcontrols named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedSubcontrols(name string) ([]*Subcontrol, error) {
	if ip.Edges.namedSubcontrols == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedSubcontrols[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedSubcontrols(name string, edges ...*Subcontrol) {
	if ip.Edges.namedSubcontrols == nil {
		ip.Edges.namedSubcontrols = make(map[string][]*Subcontrol)
	}
	if len(edges) == 0 {
		ip.Edges.namedSubcontrols[name] = []*Subcontrol{}
	} else {
		ip.Edges.namedSubcontrols[name] = append(ip.Edges.namedSubcontrols[name], edges...)
	}
}

// NamedProcedures returns the Procedures named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedProcedures(name string) ([]*Procedure, error) {
	if ip.Edges.namedProcedures == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedProcedures[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedProcedures(name string, edges ...*Procedure) {
	if ip.Edges.namedProcedures == nil {
		ip.Edges.namedProcedures = make(map[string][]*Procedure)
	}
	if len(edges) == 0 {
		ip.Edges.namedProcedures[name] = []*Procedure{}
	} else {
		ip.Edges.namedProcedures[name] = append(ip.Edges.namedProcedures[name], edges...)
	}
}

// NamedNarratives returns the Narratives named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedNarratives(name string) ([]*Narrative, error) {
	if ip.Edges.namedNarratives == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedNarratives[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedNarratives(name string, edges ...*Narrative) {
	if ip.Edges.namedNarratives == nil {
		ip.Edges.namedNarratives = make(map[string][]*Narrative)
	}
	if len(edges) == 0 {
		ip.Edges.namedNarratives[name] = []*Narrative{}
	} else {
		ip.Edges.namedNarratives[name] = append(ip.Edges.namedNarratives[name], edges...)
	}
}

// NamedTasks returns the Tasks named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedTasks(name string) ([]*Task, error) {
	if ip.Edges.namedTasks == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedTasks[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedTasks(name string, edges ...*Task) {
	if ip.Edges.namedTasks == nil {
		ip.Edges.namedTasks = make(map[string][]*Task)
	}
	if len(edges) == 0 {
		ip.Edges.namedTasks[name] = []*Task{}
	} else {
		ip.Edges.namedTasks[name] = append(ip.Edges.namedTasks[name], edges...)
	}
}

// NamedRisks returns the Risks named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedRisks(name string) ([]*Risk, error) {
	if ip.Edges.namedRisks == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedRisks[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedRisks(name string, edges ...*Risk) {
	if ip.Edges.namedRisks == nil {
		ip.Edges.namedRisks = make(map[string][]*Risk)
	}
	if len(edges) == 0 {
		ip.Edges.namedRisks[name] = []*Risk{}
	} else {
		ip.Edges.namedRisks[name] = append(ip.Edges.namedRisks[name], edges...)
	}
}

// NamedPrograms returns the Programs named value or an error if the edge was not
// loaded in eager-loading with this name.
func (ip *InternalPolicy) NamedPrograms(name string) ([]*Program, error) {
	if ip.Edges.namedPrograms == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := ip.Edges.namedPrograms[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (ip *InternalPolicy) appendNamedPrograms(name string, edges ...*Program) {
	if ip.Edges.namedPrograms == nil {
		ip.Edges.namedPrograms = make(map[string][]*Program)
	}
	if len(edges) == 0 {
		ip.Edges.namedPrograms[name] = []*Program{}
	} else {
		ip.Edges.namedPrograms[name] = append(ip.Edges.namedPrograms[name], edges...)
	}
}

// InternalPolicies is a parsable slice of InternalPolicy.
type InternalPolicies []*InternalPolicy
