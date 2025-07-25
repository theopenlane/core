// Code generated by ent, DO NOT EDIT.

package subprocessor

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the subprocessor type in the database.
	Label = "subprocessor"
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
	// FieldOwnerID holds the string denoting the owner_id field in the database.
	FieldOwnerID = "owner_id"
	// FieldSystemOwned holds the string denoting the system_owned field in the database.
	FieldSystemOwned = "system_owned"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldLogoRemoteURL holds the string denoting the logo_remote_url field in the database.
	FieldLogoRemoteURL = "logo_remote_url"
	// FieldLogoLocalFileID holds the string denoting the logo_local_file_id field in the database.
	FieldLogoLocalFileID = "logo_local_file_id"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// EdgeFiles holds the string denoting the files edge name in mutations.
	EdgeFiles = "files"
	// EdgeLogoFile holds the string denoting the logo_file edge name in mutations.
	EdgeLogoFile = "logo_file"
	// EdgeTrustCenterSubprocessors holds the string denoting the trust_center_subprocessors edge name in mutations.
	EdgeTrustCenterSubprocessors = "trust_center_subprocessors"
	// Table holds the table name of the subprocessor in the database.
	Table = "subprocessors"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "subprocessors"
	// OwnerInverseTable is the table name for the Organization entity.
	// It exists in this package in order to avoid circular dependency with the "organization" package.
	OwnerInverseTable = "organizations"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "owner_id"
	// FilesTable is the table that holds the files relation/edge. The primary key declared below.
	FilesTable = "subprocessor_files"
	// FilesInverseTable is the table name for the File entity.
	// It exists in this package in order to avoid circular dependency with the "file" package.
	FilesInverseTable = "files"
	// LogoFileTable is the table that holds the logo_file relation/edge.
	LogoFileTable = "subprocessors"
	// LogoFileInverseTable is the table name for the File entity.
	// It exists in this package in order to avoid circular dependency with the "file" package.
	LogoFileInverseTable = "files"
	// LogoFileColumn is the table column denoting the logo_file relation/edge.
	LogoFileColumn = "logo_local_file_id"
	// TrustCenterSubprocessorsTable is the table that holds the trust_center_subprocessors relation/edge.
	TrustCenterSubprocessorsTable = "trust_center_subprocessors"
	// TrustCenterSubprocessorsInverseTable is the table name for the TrustCenterSubprocessor entity.
	// It exists in this package in order to avoid circular dependency with the "trustcentersubprocessor" package.
	TrustCenterSubprocessorsInverseTable = "trust_center_subprocessors"
	// TrustCenterSubprocessorsColumn is the table column denoting the trust_center_subprocessors relation/edge.
	TrustCenterSubprocessorsColumn = "subprocessor_id"
)

// Columns holds all SQL columns for subprocessor fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldCreatedBy,
	FieldUpdatedBy,
	FieldDeletedAt,
	FieldDeletedBy,
	FieldTags,
	FieldOwnerID,
	FieldSystemOwned,
	FieldName,
	FieldDescription,
	FieldLogoRemoteURL,
	FieldLogoLocalFileID,
}

var (
	// FilesPrimaryKey and FilesColumn2 are the table columns denoting the
	// primary key for the files relation (M2M).
	FilesPrimaryKey = []string{"subprocessor_id", "file_id"}
)

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
	Hooks        [7]ent.Hook
	Interceptors [4]ent.Interceptor
	Policy       ent.Policy
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultTags holds the default value on creation for the "tags" field.
	DefaultTags []string
	// DefaultSystemOwned holds the default value on creation for the "system_owned" field.
	DefaultSystemOwned bool
	// NameValidator is a validator for the "name" field. It is called by the builders before save.
	NameValidator func(string) error
	// LogoRemoteURLValidator is a validator for the "logo_remote_url" field. It is called by the builders before save.
	LogoRemoteURLValidator func(string) error
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

// OrderOption defines the ordering options for the Subprocessor queries.
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

// ByOwnerID orders the results by the owner_id field.
func ByOwnerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOwnerID, opts...).ToFunc()
}

// BySystemOwned orders the results by the system_owned field.
func BySystemOwned(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSystemOwned, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByLogoRemoteURL orders the results by the logo_remote_url field.
func ByLogoRemoteURL(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLogoRemoteURL, opts...).ToFunc()
}

// ByLogoLocalFileID orders the results by the logo_local_file_id field.
func ByLogoLocalFileID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLogoLocalFileID, opts...).ToFunc()
}

// ByOwnerField orders the results by owner field.
func ByOwnerField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOwnerStep(), sql.OrderByField(field, opts...))
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

// ByLogoFileField orders the results by logo_file field.
func ByLogoFileField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newLogoFileStep(), sql.OrderByField(field, opts...))
	}
}

// ByTrustCenterSubprocessorsCount orders the results by trust_center_subprocessors count.
func ByTrustCenterSubprocessorsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newTrustCenterSubprocessorsStep(), opts...)
	}
}

// ByTrustCenterSubprocessors orders the results by trust_center_subprocessors terms.
func ByTrustCenterSubprocessors(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newTrustCenterSubprocessorsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newOwnerStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OwnerInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
	)
}
func newFilesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FilesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, FilesTable, FilesPrimaryKey...),
	)
}
func newLogoFileStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(LogoFileInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, LogoFileTable, LogoFileColumn),
	)
}
func newTrustCenterSubprocessorsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(TrustCenterSubprocessorsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, TrustCenterSubprocessorsTable, TrustCenterSubprocessorsColumn),
	)
}
