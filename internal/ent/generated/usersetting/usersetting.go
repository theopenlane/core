// Code generated by ent, DO NOT EDIT.

package usersetting

import (
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/core/pkg/enums"
)

const (
	// Label holds the string label denoting the usersetting type in the database.
	Label = "user_setting"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
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
	// FieldTags holds the string denoting the tags field in the database.
	FieldTags = "tags"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldLocked holds the string denoting the locked field in the database.
	FieldLocked = "locked"
	// FieldSilencedAt holds the string denoting the silenced_at field in the database.
	FieldSilencedAt = "silenced_at"
	// FieldSuspendedAt holds the string denoting the suspended_at field in the database.
	FieldSuspendedAt = "suspended_at"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldEmailConfirmed holds the string denoting the email_confirmed field in the database.
	FieldEmailConfirmed = "email_confirmed"
	// FieldIsWebauthnAllowed holds the string denoting the is_webauthn_allowed field in the database.
	FieldIsWebauthnAllowed = "is_webauthn_allowed"
	// FieldIsTfaEnabled holds the string denoting the is_tfa_enabled field in the database.
	FieldIsTfaEnabled = "is_tfa_enabled"
	// FieldPhoneNumber holds the string denoting the phone_number field in the database.
	FieldPhoneNumber = "phone_number"
	// EdgeUser holds the string denoting the user edge name in mutations.
	EdgeUser = "user"
	// EdgeDefaultOrg holds the string denoting the default_org edge name in mutations.
	EdgeDefaultOrg = "default_org"
	// EdgeFiles holds the string denoting the files edge name in mutations.
	EdgeFiles = "files"
	// Table holds the table name of the usersetting in the database.
	Table = "user_settings"
	// UserTable is the table that holds the user relation/edge.
	UserTable = "user_settings"
	// UserInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UserInverseTable = "users"
	// UserColumn is the table column denoting the user relation/edge.
	UserColumn = "user_id"
	// DefaultOrgTable is the table that holds the default_org relation/edge.
	DefaultOrgTable = "user_settings"
	// DefaultOrgInverseTable is the table name for the Organization entity.
	// It exists in this package in order to avoid circular dependency with the "organization" package.
	DefaultOrgInverseTable = "organizations"
	// DefaultOrgColumn is the table column denoting the default_org relation/edge.
	DefaultOrgColumn = "user_setting_default_org"
	// FilesTable is the table that holds the files relation/edge. The primary key declared below.
	FilesTable = "user_setting_files"
	// FilesInverseTable is the table name for the File entity.
	// It exists in this package in order to avoid circular dependency with the "file" package.
	FilesInverseTable = "files"
)

// Columns holds all SQL columns for usersetting fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldCreatedBy,
	FieldUpdatedBy,
	FieldDeletedAt,
	FieldDeletedBy,
	FieldTags,
	FieldUserID,
	FieldLocked,
	FieldSilencedAt,
	FieldSuspendedAt,
	FieldStatus,
	FieldEmailConfirmed,
	FieldIsWebauthnAllowed,
	FieldIsTfaEnabled,
	FieldPhoneNumber,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "user_settings"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"user_setting_default_org",
}

var (
	// FilesPrimaryKey and FilesColumn2 are the table columns denoting the
	// primary key for the files relation (M2M).
	FilesPrimaryKey = []string{"user_setting_id", "file_id"}
)

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
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
	Hooks        [5]ent.Hook
	Interceptors [3]ent.Interceptor
	Policy       ent.Policy
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultTags holds the default value on creation for the "tags" field.
	DefaultTags []string
	// DefaultLocked holds the default value on creation for the "locked" field.
	DefaultLocked bool
	// DefaultEmailConfirmed holds the default value on creation for the "email_confirmed" field.
	DefaultEmailConfirmed bool
	// DefaultIsWebauthnAllowed holds the default value on creation for the "is_webauthn_allowed" field.
	DefaultIsWebauthnAllowed bool
	// DefaultIsTfaEnabled holds the default value on creation for the "is_tfa_enabled" field.
	DefaultIsTfaEnabled bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

const DefaultStatus enums.UserStatus = "ACTIVE"

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s enums.UserStatus) error {
	switch s.String() {
	case "ACTIVE", "INACTIVE", "DEACTIVATED", "SUSPENDED", "ONBOARDING":
		return nil
	default:
		return fmt.Errorf("usersetting: invalid enum value for status field: %q", s)
	}
}

// OrderOption defines the ordering options for the UserSetting queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
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

// ByUserID orders the results by the user_id field.
func ByUserID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUserID, opts...).ToFunc()
}

// ByLocked orders the results by the locked field.
func ByLocked(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLocked, opts...).ToFunc()
}

// BySilencedAt orders the results by the silenced_at field.
func BySilencedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSilencedAt, opts...).ToFunc()
}

// BySuspendedAt orders the results by the suspended_at field.
func BySuspendedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSuspendedAt, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByEmailConfirmed orders the results by the email_confirmed field.
func ByEmailConfirmed(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEmailConfirmed, opts...).ToFunc()
}

// ByIsWebauthnAllowed orders the results by the is_webauthn_allowed field.
func ByIsWebauthnAllowed(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsWebauthnAllowed, opts...).ToFunc()
}

// ByIsTfaEnabled orders the results by the is_tfa_enabled field.
func ByIsTfaEnabled(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsTfaEnabled, opts...).ToFunc()
}

// ByPhoneNumber orders the results by the phone_number field.
func ByPhoneNumber(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPhoneNumber, opts...).ToFunc()
}

// ByUserField orders the results by user field.
func ByUserField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUserStep(), sql.OrderByField(field, opts...))
	}
}

// ByDefaultOrgField orders the results by default_org field.
func ByDefaultOrgField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newDefaultOrgStep(), sql.OrderByField(field, opts...))
	}
}

// ByFilesCount orders the results by files count.
func ByFilesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newFilesStep(), opts...)
	}
}

// ByFiles orders the results by files terms.
func ByFiles(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFilesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newUserStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UserInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, true, UserTable, UserColumn),
	)
}
func newDefaultOrgStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(DefaultOrgInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, DefaultOrgTable, DefaultOrgColumn),
	)
}
func newFilesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FilesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, FilesTable, FilesPrimaryKey...),
	)
}

var (
	// enums.UserStatus must implement graphql.Marshaler.
	_ graphql.Marshaler = (*enums.UserStatus)(nil)
	// enums.UserStatus must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*enums.UserStatus)(nil)
)
