// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/subcontrolhistory"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx/history"
)

// SubcontrolHistory is the model entity for the SubcontrolHistory schema.
type SubcontrolHistory struct {
	config `json:"-"`
	// ID of the ent.
	ID string `json:"id,omitempty"`
	// HistoryTime holds the value of the "history_time" field.
	HistoryTime time.Time `json:"history_time,omitempty"`
	// Ref holds the value of the "ref" field.
	Ref string `json:"ref,omitempty"`
	// Operation holds the value of the "operation" field.
	Operation history.OpType `json:"operation,omitempty"`
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
	// description of what the control is supposed to accomplish
	Description string `json:"description,omitempty"`
	// internal reference id of the control, can be used for internal tracking
	ReferenceID string `json:"reference_id,omitempty"`
	// external auditor id of the control, can be used to map to external audit partner mappings
	AuditorReferenceID string `json:"auditor_reference_id,omitempty"`
	// status of the control
	Status enums.ControlStatus `json:"status,omitempty"`
	// source of the control, e.g. framework, template, custom, etc.
	Source enums.ControlSource `json:"source,omitempty"`
	// the reference framework for the control if it came from a standard, empty if not associated with a standard
	ReferenceFramework *string `json:"reference_framework,omitempty"`
	// type of the control e.g. preventive, detective, corrective, or deterrent.
	ControlType enums.ControlType `json:"control_type,omitempty"`
	// category of the control
	Category string `json:"category,omitempty"`
	// category id of the control
	CategoryID string `json:"category_id,omitempty"`
	// subcategory of the control
	Subcategory string `json:"subcategory,omitempty"`
	// mapped categories of the control to other standards
	MappedCategories []string `json:"mapped_categories,omitempty"`
	// objectives of the audit assessment for the control
	AssessmentObjectives []models.AssessmentObjective `json:"assessment_objectives,omitempty"`
	// methods used to verify the control implementation during an audit
	AssessmentMethods []models.AssessmentMethod `json:"assessment_methods,omitempty"`
	// questions to ask to verify the control
	ControlQuestions []string `json:"control_questions,omitempty"`
	// implementation guidance for the control
	ImplementationGuidance []models.ImplementationGuidance `json:"implementation_guidance,omitempty"`
	// examples of evidence for the control
	ExampleEvidence []models.ExampleEvidence `json:"example_evidence,omitempty"`
	// references for the control
	References []models.Reference `json:"references,omitempty"`
	// the id of the group that owns the control
	ControlOwnerID *string `json:"control_owner_id,omitempty"`
	// the id of the group that is temporarily delegated to own the control
	DelegateID string `json:"delegate_id,omitempty"`
	// the ID of the organization owner of the object
	OwnerID string `json:"owner_id,omitempty"`
	// the unique reference code for the control
	RefCode string `json:"ref_code,omitempty"`
	// the id of the parent control
	ControlID    string `json:"control_id,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SubcontrolHistory) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case subcontrolhistory.FieldTags, subcontrolhistory.FieldMappedCategories, subcontrolhistory.FieldAssessmentObjectives, subcontrolhistory.FieldAssessmentMethods, subcontrolhistory.FieldControlQuestions, subcontrolhistory.FieldImplementationGuidance, subcontrolhistory.FieldExampleEvidence, subcontrolhistory.FieldReferences:
			values[i] = new([]byte)
		case subcontrolhistory.FieldOperation:
			values[i] = new(history.OpType)
		case subcontrolhistory.FieldID, subcontrolhistory.FieldRef, subcontrolhistory.FieldCreatedBy, subcontrolhistory.FieldUpdatedBy, subcontrolhistory.FieldDeletedBy, subcontrolhistory.FieldDisplayID, subcontrolhistory.FieldDescription, subcontrolhistory.FieldReferenceID, subcontrolhistory.FieldAuditorReferenceID, subcontrolhistory.FieldStatus, subcontrolhistory.FieldSource, subcontrolhistory.FieldReferenceFramework, subcontrolhistory.FieldControlType, subcontrolhistory.FieldCategory, subcontrolhistory.FieldCategoryID, subcontrolhistory.FieldSubcategory, subcontrolhistory.FieldControlOwnerID, subcontrolhistory.FieldDelegateID, subcontrolhistory.FieldOwnerID, subcontrolhistory.FieldRefCode, subcontrolhistory.FieldControlID:
			values[i] = new(sql.NullString)
		case subcontrolhistory.FieldHistoryTime, subcontrolhistory.FieldCreatedAt, subcontrolhistory.FieldUpdatedAt, subcontrolhistory.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SubcontrolHistory fields.
func (sh *SubcontrolHistory) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case subcontrolhistory.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				sh.ID = value.String
			}
		case subcontrolhistory.FieldHistoryTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field history_time", values[i])
			} else if value.Valid {
				sh.HistoryTime = value.Time
			}
		case subcontrolhistory.FieldRef:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field ref", values[i])
			} else if value.Valid {
				sh.Ref = value.String
			}
		case subcontrolhistory.FieldOperation:
			if value, ok := values[i].(*history.OpType); !ok {
				return fmt.Errorf("unexpected type %T for field operation", values[i])
			} else if value != nil {
				sh.Operation = *value
			}
		case subcontrolhistory.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				sh.CreatedAt = value.Time
			}
		case subcontrolhistory.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				sh.UpdatedAt = value.Time
			}
		case subcontrolhistory.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				sh.CreatedBy = value.String
			}
		case subcontrolhistory.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				sh.UpdatedBy = value.String
			}
		case subcontrolhistory.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				sh.DeletedAt = value.Time
			}
		case subcontrolhistory.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				sh.DeletedBy = value.String
			}
		case subcontrolhistory.FieldDisplayID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field display_id", values[i])
			} else if value.Valid {
				sh.DisplayID = value.String
			}
		case subcontrolhistory.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case subcontrolhistory.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				sh.Description = value.String
			}
		case subcontrolhistory.FieldReferenceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field reference_id", values[i])
			} else if value.Valid {
				sh.ReferenceID = value.String
			}
		case subcontrolhistory.FieldAuditorReferenceID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field auditor_reference_id", values[i])
			} else if value.Valid {
				sh.AuditorReferenceID = value.String
			}
		case subcontrolhistory.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				sh.Status = enums.ControlStatus(value.String)
			}
		case subcontrolhistory.FieldSource:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field source", values[i])
			} else if value.Valid {
				sh.Source = enums.ControlSource(value.String)
			}
		case subcontrolhistory.FieldReferenceFramework:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field reference_framework", values[i])
			} else if value.Valid {
				sh.ReferenceFramework = new(string)
				*sh.ReferenceFramework = value.String
			}
		case subcontrolhistory.FieldControlType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field control_type", values[i])
			} else if value.Valid {
				sh.ControlType = enums.ControlType(value.String)
			}
		case subcontrolhistory.FieldCategory:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field category", values[i])
			} else if value.Valid {
				sh.Category = value.String
			}
		case subcontrolhistory.FieldCategoryID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field category_id", values[i])
			} else if value.Valid {
				sh.CategoryID = value.String
			}
		case subcontrolhistory.FieldSubcategory:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field subcategory", values[i])
			} else if value.Valid {
				sh.Subcategory = value.String
			}
		case subcontrolhistory.FieldMappedCategories:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field mapped_categories", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.MappedCategories); err != nil {
					return fmt.Errorf("unmarshal field mapped_categories: %w", err)
				}
			}
		case subcontrolhistory.FieldAssessmentObjectives:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field assessment_objectives", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.AssessmentObjectives); err != nil {
					return fmt.Errorf("unmarshal field assessment_objectives: %w", err)
				}
			}
		case subcontrolhistory.FieldAssessmentMethods:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field assessment_methods", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.AssessmentMethods); err != nil {
					return fmt.Errorf("unmarshal field assessment_methods: %w", err)
				}
			}
		case subcontrolhistory.FieldControlQuestions:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field control_questions", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.ControlQuestions); err != nil {
					return fmt.Errorf("unmarshal field control_questions: %w", err)
				}
			}
		case subcontrolhistory.FieldImplementationGuidance:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field implementation_guidance", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.ImplementationGuidance); err != nil {
					return fmt.Errorf("unmarshal field implementation_guidance: %w", err)
				}
			}
		case subcontrolhistory.FieldExampleEvidence:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field example_evidence", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.ExampleEvidence); err != nil {
					return fmt.Errorf("unmarshal field example_evidence: %w", err)
				}
			}
		case subcontrolhistory.FieldReferences:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field references", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &sh.References); err != nil {
					return fmt.Errorf("unmarshal field references: %w", err)
				}
			}
		case subcontrolhistory.FieldControlOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field control_owner_id", values[i])
			} else if value.Valid {
				sh.ControlOwnerID = new(string)
				*sh.ControlOwnerID = value.String
			}
		case subcontrolhistory.FieldDelegateID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field delegate_id", values[i])
			} else if value.Valid {
				sh.DelegateID = value.String
			}
		case subcontrolhistory.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				sh.OwnerID = value.String
			}
		case subcontrolhistory.FieldRefCode:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field ref_code", values[i])
			} else if value.Valid {
				sh.RefCode = value.String
			}
		case subcontrolhistory.FieldControlID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field control_id", values[i])
			} else if value.Valid {
				sh.ControlID = value.String
			}
		default:
			sh.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SubcontrolHistory.
// This includes values selected through modifiers, order, etc.
func (sh *SubcontrolHistory) Value(name string) (ent.Value, error) {
	return sh.selectValues.Get(name)
}

// Update returns a builder for updating this SubcontrolHistory.
// Note that you need to call SubcontrolHistory.Unwrap() before calling this method if this SubcontrolHistory
// was returned from a transaction, and the transaction was committed or rolled back.
func (sh *SubcontrolHistory) Update() *SubcontrolHistoryUpdateOne {
	return NewSubcontrolHistoryClient(sh.config).UpdateOne(sh)
}

// Unwrap unwraps the SubcontrolHistory entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (sh *SubcontrolHistory) Unwrap() *SubcontrolHistory {
	_tx, ok := sh.config.driver.(*txDriver)
	if !ok {
		panic("generated: SubcontrolHistory is not a transactional entity")
	}
	sh.config.driver = _tx.drv
	return sh
}

// String implements the fmt.Stringer.
func (sh *SubcontrolHistory) String() string {
	var builder strings.Builder
	builder.WriteString("SubcontrolHistory(")
	builder.WriteString(fmt.Sprintf("id=%v, ", sh.ID))
	builder.WriteString("history_time=")
	builder.WriteString(sh.HistoryTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("ref=")
	builder.WriteString(sh.Ref)
	builder.WriteString(", ")
	builder.WriteString("operation=")
	builder.WriteString(fmt.Sprintf("%v", sh.Operation))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(sh.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(sh.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(sh.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(sh.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(sh.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(sh.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("display_id=")
	builder.WriteString(sh.DisplayID)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", sh.Tags))
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(sh.Description)
	builder.WriteString(", ")
	builder.WriteString("reference_id=")
	builder.WriteString(sh.ReferenceID)
	builder.WriteString(", ")
	builder.WriteString("auditor_reference_id=")
	builder.WriteString(sh.AuditorReferenceID)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", sh.Status))
	builder.WriteString(", ")
	builder.WriteString("source=")
	builder.WriteString(fmt.Sprintf("%v", sh.Source))
	builder.WriteString(", ")
	if v := sh.ReferenceFramework; v != nil {
		builder.WriteString("reference_framework=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("control_type=")
	builder.WriteString(fmt.Sprintf("%v", sh.ControlType))
	builder.WriteString(", ")
	builder.WriteString("category=")
	builder.WriteString(sh.Category)
	builder.WriteString(", ")
	builder.WriteString("category_id=")
	builder.WriteString(sh.CategoryID)
	builder.WriteString(", ")
	builder.WriteString("subcategory=")
	builder.WriteString(sh.Subcategory)
	builder.WriteString(", ")
	builder.WriteString("mapped_categories=")
	builder.WriteString(fmt.Sprintf("%v", sh.MappedCategories))
	builder.WriteString(", ")
	builder.WriteString("assessment_objectives=")
	builder.WriteString(fmt.Sprintf("%v", sh.AssessmentObjectives))
	builder.WriteString(", ")
	builder.WriteString("assessment_methods=")
	builder.WriteString(fmt.Sprintf("%v", sh.AssessmentMethods))
	builder.WriteString(", ")
	builder.WriteString("control_questions=")
	builder.WriteString(fmt.Sprintf("%v", sh.ControlQuestions))
	builder.WriteString(", ")
	builder.WriteString("implementation_guidance=")
	builder.WriteString(fmt.Sprintf("%v", sh.ImplementationGuidance))
	builder.WriteString(", ")
	builder.WriteString("example_evidence=")
	builder.WriteString(fmt.Sprintf("%v", sh.ExampleEvidence))
	builder.WriteString(", ")
	builder.WriteString("references=")
	builder.WriteString(fmt.Sprintf("%v", sh.References))
	builder.WriteString(", ")
	if v := sh.ControlOwnerID; v != nil {
		builder.WriteString("control_owner_id=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("delegate_id=")
	builder.WriteString(sh.DelegateID)
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(sh.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("ref_code=")
	builder.WriteString(sh.RefCode)
	builder.WriteString(", ")
	builder.WriteString("control_id=")
	builder.WriteString(sh.ControlID)
	builder.WriteByte(')')
	return builder.String()
}

// SubcontrolHistories is a parsable slice of SubcontrolHistory.
type SubcontrolHistories []*SubcontrolHistory
