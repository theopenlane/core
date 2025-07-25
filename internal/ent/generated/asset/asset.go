// Code generated by ent, DO NOT EDIT.

package asset

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
	// Label holds the string label denoting the asset type in the database.
	Label = "asset"
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
	// FieldAssetType holds the string denoting the asset_type field in the database.
	FieldAssetType = "asset_type"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldIdentifier holds the string denoting the identifier field in the database.
	FieldIdentifier = "identifier"
	// FieldWebsite holds the string denoting the website field in the database.
	FieldWebsite = "website"
	// FieldCpe holds the string denoting the cpe field in the database.
	FieldCpe = "cpe"
	// FieldCategories holds the string denoting the categories field in the database.
	FieldCategories = "categories"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// EdgeBlockedGroups holds the string denoting the blocked_groups edge name in mutations.
	EdgeBlockedGroups = "blocked_groups"
	// EdgeEditors holds the string denoting the editors edge name in mutations.
	EdgeEditors = "editors"
	// EdgeViewers holds the string denoting the viewers edge name in mutations.
	EdgeViewers = "viewers"
	// EdgeScans holds the string denoting the scans edge name in mutations.
	EdgeScans = "scans"
	// EdgeEntities holds the string denoting the entities edge name in mutations.
	EdgeEntities = "entities"
	// EdgeControls holds the string denoting the controls edge name in mutations.
	EdgeControls = "controls"
	// Table holds the table name of the asset in the database.
	Table = "assets"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "assets"
	// OwnerInverseTable is the table name for the Organization entity.
	// It exists in this package in order to avoid circular dependency with the "organization" package.
	OwnerInverseTable = "organizations"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "owner_id"
	// BlockedGroupsTable is the table that holds the blocked_groups relation/edge.
	BlockedGroupsTable = "groups"
	// BlockedGroupsInverseTable is the table name for the Group entity.
	// It exists in this package in order to avoid circular dependency with the "group" package.
	BlockedGroupsInverseTable = "groups"
	// BlockedGroupsColumn is the table column denoting the blocked_groups relation/edge.
	BlockedGroupsColumn = "asset_blocked_groups"
	// EditorsTable is the table that holds the editors relation/edge.
	EditorsTable = "groups"
	// EditorsInverseTable is the table name for the Group entity.
	// It exists in this package in order to avoid circular dependency with the "group" package.
	EditorsInverseTable = "groups"
	// EditorsColumn is the table column denoting the editors relation/edge.
	EditorsColumn = "asset_editors"
	// ViewersTable is the table that holds the viewers relation/edge.
	ViewersTable = "groups"
	// ViewersInverseTable is the table name for the Group entity.
	// It exists in this package in order to avoid circular dependency with the "group" package.
	ViewersInverseTable = "groups"
	// ViewersColumn is the table column denoting the viewers relation/edge.
	ViewersColumn = "asset_viewers"
	// ScansTable is the table that holds the scans relation/edge. The primary key declared below.
	ScansTable = "scan_assets"
	// ScansInverseTable is the table name for the Scan entity.
	// It exists in this package in order to avoid circular dependency with the "scan" package.
	ScansInverseTable = "scans"
	// EntitiesTable is the table that holds the entities relation/edge. The primary key declared below.
	EntitiesTable = "entity_assets"
	// EntitiesInverseTable is the table name for the Entity entity.
	// It exists in this package in order to avoid circular dependency with the "entity" package.
	EntitiesInverseTable = "entities"
	// ControlsTable is the table that holds the controls relation/edge. The primary key declared below.
	ControlsTable = "control_assets"
	// ControlsInverseTable is the table name for the Control entity.
	// It exists in this package in order to avoid circular dependency with the "control" package.
	ControlsInverseTable = "controls"
)

// Columns holds all SQL columns for asset fields.
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
	FieldAssetType,
	FieldName,
	FieldDescription,
	FieldIdentifier,
	FieldWebsite,
	FieldCpe,
	FieldCategories,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "assets"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"risk_assets",
}

var (
	// ScansPrimaryKey and ScansColumn2 are the table columns denoting the
	// primary key for the scans relation (M2M).
	ScansPrimaryKey = []string{"scan_id", "asset_id"}
	// EntitiesPrimaryKey and EntitiesColumn2 are the table columns denoting the
	// primary key for the entities relation (M2M).
	EntitiesPrimaryKey = []string{"entity_id", "asset_id"}
	// ControlsPrimaryKey and ControlsColumn2 are the table columns denoting the
	// primary key for the controls relation (M2M).
	ControlsPrimaryKey = []string{"control_id", "asset_id"}
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
	Hooks        [8]ent.Hook
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
	// OwnerIDValidator is a validator for the "owner_id" field. It is called by the builders before save.
	OwnerIDValidator func(string) error
	// NameValidator is a validator for the "name" field. It is called by the builders before save.
	NameValidator func(string) error
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

const DefaultAssetType enums.AssetType = "TECHNOLOGY"

// AssetTypeValidator is a validator for the "asset_type" field enum values. It is called by the builders before save.
func AssetTypeValidator(at enums.AssetType) error {
	switch at.String() {
	case "TECHNOLOGY", "DOMAIN", "DEVICE", "TELEPHONE":
		return nil
	default:
		return fmt.Errorf("asset: invalid enum value for asset_type field: %q", at)
	}
}

// OrderOption defines the ordering options for the Asset queries.
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

// ByAssetType orders the results by the asset_type field.
func ByAssetType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAssetType, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByIdentifier orders the results by the identifier field.
func ByIdentifier(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIdentifier, opts...).ToFunc()
}

// ByWebsite orders the results by the website field.
func ByWebsite(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldWebsite, opts...).ToFunc()
}

// ByCpe orders the results by the cpe field.
func ByCpe(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCpe, opts...).ToFunc()
}

// ByOwnerField orders the results by owner field.
func ByOwnerField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOwnerStep(), sql.OrderByField(field, opts...))
	}
}

// ByBlockedGroupsCount orders the results by blocked_groups count.
func ByBlockedGroupsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newBlockedGroupsStep(), opts...)
	}
}

// ByBlockedGroups orders the results by blocked_groups terms.
func ByBlockedGroups(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newBlockedGroupsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByEditorsCount orders the results by editors count.
func ByEditorsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEditorsStep(), opts...)
	}
}

// ByEditors orders the results by editors terms.
func ByEditors(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEditorsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByViewersCount orders the results by viewers count.
func ByViewersCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newViewersStep(), opts...)
	}
}

// ByViewers orders the results by viewers terms.
func ByViewers(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newViewersStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByScansCount orders the results by scans count.
func ByScansCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newScansStep(), opts...)
	}
}

// ByScans orders the results by scans terms.
func ByScans(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newScansStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByEntitiesCount orders the results by entities count.
func ByEntitiesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEntitiesStep(), opts...)
	}
}

// ByEntities orders the results by entities terms.
func ByEntities(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEntitiesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByControlsCount orders the results by controls count.
func ByControlsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newControlsStep(), opts...)
	}
}

// ByControls orders the results by controls terms.
func ByControls(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newOwnerStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OwnerInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
	)
}
func newBlockedGroupsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(BlockedGroupsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, BlockedGroupsTable, BlockedGroupsColumn),
	)
}
func newEditorsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EditorsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, EditorsTable, EditorsColumn),
	)
}
func newViewersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ViewersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, ViewersTable, ViewersColumn),
	)
}
func newScansStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ScansInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, ScansTable, ScansPrimaryKey...),
	)
}
func newEntitiesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EntitiesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, EntitiesTable, EntitiesPrimaryKey...),
	)
}
func newControlsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, ControlsTable, ControlsPrimaryKey...),
	)
}

var (
	// enums.AssetType must implement graphql.Marshaler.
	_ graphql.Marshaler = (*enums.AssetType)(nil)
	// enums.AssetType must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*enums.AssetType)(nil)
)
