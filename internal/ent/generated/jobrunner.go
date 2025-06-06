// Code generated by ent, DO NOT EDIT.

package generated

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/theopenlane/core/internal/ent/generated/jobrunner"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/pkg/enums"
)

// JobRunner is the model entity for the JobRunner schema.
type JobRunner struct {
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
	// the organization id that owns the object
	OwnerID string `json:"owner_id,omitempty"`
	// indicates if the record is owned by the the openlane system and not by an organization
	SystemOwned bool `json:"system_owned,omitempty"`
	// the name of the runner
	Name string `json:"name,omitempty"`
	// the status of this runner
	Status enums.JobRunnerStatus `json:"status,omitempty"`
	// the IP address of this runner
	IPAddress string `json:"ip_address,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the JobRunnerQuery when eager-loading is set.
	Edges        JobRunnerEdges `json:"edges"`
	selectValues sql.SelectValues
}

// JobRunnerEdges holds the relations/edges for other nodes in the graph.
type JobRunnerEdges struct {
	// Owner holds the value of the owner edge.
	Owner *Organization `json:"owner,omitempty"`
	// JobRunnerTokens holds the value of the job_runner_tokens edge.
	JobRunnerTokens []*JobRunnerToken `json:"job_runner_tokens,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
	// totalCount holds the count of the edges above.
	totalCount [2]map[string]int

	namedJobRunnerTokens map[string][]*JobRunnerToken
}

// OwnerOrErr returns the Owner value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e JobRunnerEdges) OwnerOrErr() (*Organization, error) {
	if e.Owner != nil {
		return e.Owner, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: organization.Label}
	}
	return nil, &NotLoadedError{edge: "owner"}
}

// JobRunnerTokensOrErr returns the JobRunnerTokens value or an error if the edge
// was not loaded in eager-loading.
func (e JobRunnerEdges) JobRunnerTokensOrErr() ([]*JobRunnerToken, error) {
	if e.loadedTypes[1] {
		return e.JobRunnerTokens, nil
	}
	return nil, &NotLoadedError{edge: "job_runner_tokens"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*JobRunner) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case jobrunner.FieldTags:
			values[i] = new([]byte)
		case jobrunner.FieldSystemOwned:
			values[i] = new(sql.NullBool)
		case jobrunner.FieldID, jobrunner.FieldCreatedBy, jobrunner.FieldUpdatedBy, jobrunner.FieldDeletedBy, jobrunner.FieldDisplayID, jobrunner.FieldOwnerID, jobrunner.FieldName, jobrunner.FieldStatus, jobrunner.FieldIPAddress:
			values[i] = new(sql.NullString)
		case jobrunner.FieldCreatedAt, jobrunner.FieldUpdatedAt, jobrunner.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the JobRunner fields.
func (jr *JobRunner) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case jobrunner.FieldID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value.Valid {
				jr.ID = value.String
			}
		case jobrunner.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				jr.CreatedAt = value.Time
			}
		case jobrunner.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				jr.UpdatedAt = value.Time
			}
		case jobrunner.FieldCreatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field created_by", values[i])
			} else if value.Valid {
				jr.CreatedBy = value.String
			}
		case jobrunner.FieldUpdatedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field updated_by", values[i])
			} else if value.Valid {
				jr.UpdatedBy = value.String
			}
		case jobrunner.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				jr.DeletedAt = value.Time
			}
		case jobrunner.FieldDeletedBy:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_by", values[i])
			} else if value.Valid {
				jr.DeletedBy = value.String
			}
		case jobrunner.FieldDisplayID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field display_id", values[i])
			} else if value.Valid {
				jr.DisplayID = value.String
			}
		case jobrunner.FieldTags:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field tags", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &jr.Tags); err != nil {
					return fmt.Errorf("unmarshal field tags: %w", err)
				}
			}
		case jobrunner.FieldOwnerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field owner_id", values[i])
			} else if value.Valid {
				jr.OwnerID = value.String
			}
		case jobrunner.FieldSystemOwned:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field system_owned", values[i])
			} else if value.Valid {
				jr.SystemOwned = value.Bool
			}
		case jobrunner.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				jr.Name = value.String
			}
		case jobrunner.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				jr.Status = enums.JobRunnerStatus(value.String)
			}
		case jobrunner.FieldIPAddress:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field ip_address", values[i])
			} else if value.Valid {
				jr.IPAddress = value.String
			}
		default:
			jr.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the JobRunner.
// This includes values selected through modifiers, order, etc.
func (jr *JobRunner) Value(name string) (ent.Value, error) {
	return jr.selectValues.Get(name)
}

// QueryOwner queries the "owner" edge of the JobRunner entity.
func (jr *JobRunner) QueryOwner() *OrganizationQuery {
	return NewJobRunnerClient(jr.config).QueryOwner(jr)
}

// QueryJobRunnerTokens queries the "job_runner_tokens" edge of the JobRunner entity.
func (jr *JobRunner) QueryJobRunnerTokens() *JobRunnerTokenQuery {
	return NewJobRunnerClient(jr.config).QueryJobRunnerTokens(jr)
}

// Update returns a builder for updating this JobRunner.
// Note that you need to call JobRunner.Unwrap() before calling this method if this JobRunner
// was returned from a transaction, and the transaction was committed or rolled back.
func (jr *JobRunner) Update() *JobRunnerUpdateOne {
	return NewJobRunnerClient(jr.config).UpdateOne(jr)
}

// Unwrap unwraps the JobRunner entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (jr *JobRunner) Unwrap() *JobRunner {
	_tx, ok := jr.config.driver.(*txDriver)
	if !ok {
		panic("generated: JobRunner is not a transactional entity")
	}
	jr.config.driver = _tx.drv
	return jr
}

// String implements the fmt.Stringer.
func (jr *JobRunner) String() string {
	var builder strings.Builder
	builder.WriteString("JobRunner(")
	builder.WriteString(fmt.Sprintf("id=%v, ", jr.ID))
	builder.WriteString("created_at=")
	builder.WriteString(jr.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(jr.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("created_by=")
	builder.WriteString(jr.CreatedBy)
	builder.WriteString(", ")
	builder.WriteString("updated_by=")
	builder.WriteString(jr.UpdatedBy)
	builder.WriteString(", ")
	builder.WriteString("deleted_at=")
	builder.WriteString(jr.DeletedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("deleted_by=")
	builder.WriteString(jr.DeletedBy)
	builder.WriteString(", ")
	builder.WriteString("display_id=")
	builder.WriteString(jr.DisplayID)
	builder.WriteString(", ")
	builder.WriteString("tags=")
	builder.WriteString(fmt.Sprintf("%v", jr.Tags))
	builder.WriteString(", ")
	builder.WriteString("owner_id=")
	builder.WriteString(jr.OwnerID)
	builder.WriteString(", ")
	builder.WriteString("system_owned=")
	builder.WriteString(fmt.Sprintf("%v", jr.SystemOwned))
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(jr.Name)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", jr.Status))
	builder.WriteString(", ")
	builder.WriteString("ip_address=")
	builder.WriteString(jr.IPAddress)
	builder.WriteByte(')')
	return builder.String()
}

// NamedJobRunnerTokens returns the JobRunnerTokens named value or an error if the edge was not
// loaded in eager-loading with this name.
func (jr *JobRunner) NamedJobRunnerTokens(name string) ([]*JobRunnerToken, error) {
	if jr.Edges.namedJobRunnerTokens == nil {
		return nil, &NotLoadedError{edge: name}
	}
	nodes, ok := jr.Edges.namedJobRunnerTokens[name]
	if !ok {
		return nil, &NotLoadedError{edge: name}
	}
	return nodes, nil
}

func (jr *JobRunner) appendNamedJobRunnerTokens(name string, edges ...*JobRunnerToken) {
	if jr.Edges.namedJobRunnerTokens == nil {
		jr.Edges.namedJobRunnerTokens = make(map[string][]*JobRunnerToken)
	}
	if len(edges) == 0 {
		jr.Edges.namedJobRunnerTokens[name] = []*JobRunnerToken{}
	} else {
		jr.Edges.namedJobRunnerTokens[name] = append(jr.Edges.namedJobRunnerTokens[name], edges...)
	}
}

// JobRunners is a parsable slice of JobRunner.
type JobRunners []*JobRunner
