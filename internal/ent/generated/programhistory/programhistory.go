// Code generated by ent, DO NOT EDIT.

package programhistory

import (
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx/history"
)

const (
	// Label holds the string label denoting the programhistory type in the database.
	Label = "program_history"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldHistoryTime holds the string denoting the history_time field in the database.
	FieldHistoryTime = "history_time"
	// FieldRef holds the string denoting the ref field in the database.
	FieldRef = "ref"
	// FieldOperation holds the string denoting the operation field in the database.
	FieldOperation = "operation"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldCreatedBy holds the string denoting the created_by field in the database.
	FieldCreatedBy = "created_by"
	// FieldUpdatedBy holds the string denoting the updated_by field in the database.
	FieldUpdatedBy = "updated_by"
	// FieldDeletedAt holds the string denoting the deleted_at field in the database.
	FieldDeletedAt = "deleted_at"
	// FieldDeletedBy holds the string denoting the deleted_by field in the database.
	FieldDeletedBy = "deleted_by"
	// FieldDisplayID holds the string denoting the display_id field in the database.
	FieldDisplayID = "display_id"
	// FieldTags holds the string denoting the tags field in the database.
	FieldTags = "tags"
	// FieldOwnerID holds the string denoting the owner_id field in the database.
	FieldOwnerID = "owner_id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldProgramType holds the string denoting the program_type field in the database.
	FieldProgramType = "program_type"
	// FieldFrameworkName holds the string denoting the framework_name field in the database.
	FieldFrameworkName = "framework_name"
	// FieldStartDate holds the string denoting the start_date field in the database.
	FieldStartDate = "start_date"
	// FieldEndDate holds the string denoting the end_date field in the database.
	FieldEndDate = "end_date"
	// FieldAuditorReady holds the string denoting the auditor_ready field in the database.
	FieldAuditorReady = "auditor_ready"
	// FieldAuditorWriteComments holds the string denoting the auditor_write_comments field in the database.
	FieldAuditorWriteComments = "auditor_write_comments"
	// FieldAuditorReadComments holds the string denoting the auditor_read_comments field in the database.
	FieldAuditorReadComments = "auditor_read_comments"
	// FieldAuditFirm holds the string denoting the audit_firm field in the database.
	FieldAuditFirm = "audit_firm"
	// FieldAuditor holds the string denoting the auditor field in the database.
	FieldAuditor = "auditor"
	// FieldAuditorEmail holds the string denoting the auditor_email field in the database.
	FieldAuditorEmail = "auditor_email"
	// Table holds the table name of the programhistory in the database.
	Table = "program_history"
)

// Columns holds all SQL columns for programhistory fields.
var Columns = []string{
	FieldID,
	FieldHistoryTime,
	FieldRef,
	FieldOperation,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldCreatedBy,
	FieldUpdatedBy,
	FieldDeletedAt,
	FieldDeletedBy,
	FieldDisplayID,
	FieldTags,
	FieldOwnerID,
	FieldName,
	FieldDescription,
	FieldStatus,
	FieldProgramType,
	FieldFrameworkName,
	FieldStartDate,
	FieldEndDate,
	FieldAuditorReady,
	FieldAuditorWriteComments,
	FieldAuditorReadComments,
	FieldAuditFirm,
	FieldAuditor,
	FieldAuditorEmail,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

// Note that the variables below are initialized by the runtime
// package on the initialization of the application. Therefore,
// it should be imported in the main as follows:
//
//	import _ "github.com/theopenlane/core/internal/ent/generated/runtime"
var (
	Hooks        [1]ent.Hook
	Interceptors [1]ent.Interceptor
	Policy       ent.Policy
	// DefaultHistoryTime holds the default value on creation for the "history_time" field.
	DefaultHistoryTime func() time.Time
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultTags holds the default value on creation for the "tags" field.
	DefaultTags []string
	// DefaultAuditorReady holds the default value on creation for the "auditor_ready" field.
	DefaultAuditorReady bool
	// DefaultAuditorWriteComments holds the default value on creation for the "auditor_write_comments" field.
	DefaultAuditorWriteComments bool
	// DefaultAuditorReadComments holds the default value on creation for the "auditor_read_comments" field.
	DefaultAuditorReadComments bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

// OperationValidator is a validator for the "operation" field enum values. It is called by the builders before save.
func OperationValidator(o history.OpType) error {
	switch o.String() {
	case "INSERT", "UPDATE", "DELETE":
		return nil
	default:
		return fmt.Errorf("programhistory: invalid enum value for operation field: %q", o)
	}
}

const DefaultStatus enums.ProgramStatus = "NOT_STARTED"

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s enums.ProgramStatus) error {
	switch s.String() {
	case "NOT_STARTED", "IN_PROGRESS", "READY_FOR_AUDITOR", "COMPLETED", "ACTION_REQUIRED":
		return nil
	default:
		return fmt.Errorf("programhistory: invalid enum value for status field: %q", s)
	}
}

const DefaultProgramType enums.ProgramType = "FRAMEWORK"

// ProgramTypeValidator is a validator for the "program_type" field enum values. It is called by the builders before save.
func ProgramTypeValidator(pt enums.ProgramType) error {
	switch pt.String() {
	case "FRAMEWORK", "GAP_ANALYSIS", "RISK_ASSESSMENT", "OTHER":
		return nil
	default:
		return fmt.Errorf("programhistory: invalid enum value for program_type field: %q", pt)
	}
}

// OrderOption defines the ordering options for the ProgramHistory queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByHistoryTime orders the results by the history_time field.
func ByHistoryTime(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHistoryTime, opts...).ToFunc()
}

// ByRef orders the results by the ref field.
func ByRef(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRef, opts...).ToFunc()
}

// ByOperation orders the results by the operation field.
func ByOperation(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOperation, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByCreatedBy orders the results by the created_by field.
func ByCreatedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedBy, opts...).ToFunc()
}

// ByUpdatedBy orders the results by the updated_by field.
func ByUpdatedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedBy, opts...).ToFunc()
}

// ByDeletedAt orders the results by the deleted_at field.
func ByDeletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeletedAt, opts...).ToFunc()
}

// ByDeletedBy orders the results by the deleted_by field.
func ByDeletedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeletedBy, opts...).ToFunc()
}

// ByDisplayID orders the results by the display_id field.
func ByDisplayID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDisplayID, opts...).ToFunc()
}

// ByOwnerID orders the results by the owner_id field.
func ByOwnerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOwnerID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByProgramType orders the results by the program_type field.
func ByProgramType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProgramType, opts...).ToFunc()
}

// ByFrameworkName orders the results by the framework_name field.
func ByFrameworkName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFrameworkName, opts...).ToFunc()
}

// ByStartDate orders the results by the start_date field.
func ByStartDate(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartDate, opts...).ToFunc()
}

// ByEndDate orders the results by the end_date field.
func ByEndDate(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEndDate, opts...).ToFunc()
}

// ByAuditorReady orders the results by the auditor_ready field.
func ByAuditorReady(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuditorReady, opts...).ToFunc()
}

// ByAuditorWriteComments orders the results by the auditor_write_comments field.
func ByAuditorWriteComments(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuditorWriteComments, opts...).ToFunc()
}

// ByAuditorReadComments orders the results by the auditor_read_comments field.
func ByAuditorReadComments(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuditorReadComments, opts...).ToFunc()
}

// ByAuditFirm orders the results by the audit_firm field.
func ByAuditFirm(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuditFirm, opts...).ToFunc()
}

// ByAuditor orders the results by the auditor field.
func ByAuditor(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuditor, opts...).ToFunc()
}

// ByAuditorEmail orders the results by the auditor_email field.
func ByAuditorEmail(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuditorEmail, opts...).ToFunc()
}

var (
	// history.OpType must implement graphql.Marshaler.
	_ graphql.Marshaler = (*history.OpType)(nil)
	// history.OpType must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*history.OpType)(nil)
)

var (
	// enums.ProgramStatus must implement graphql.Marshaler.
	_ graphql.Marshaler = (*enums.ProgramStatus)(nil)
	// enums.ProgramStatus must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*enums.ProgramStatus)(nil)
)

var (
	// enums.ProgramType must implement graphql.Marshaler.
	_ graphql.Marshaler = (*enums.ProgramType)(nil)
	// enums.ProgramType must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*enums.ProgramType)(nil)
)
