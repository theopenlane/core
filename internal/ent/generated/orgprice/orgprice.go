// Code generated by ent, DO NOT EDIT.

package orgprice

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the orgprice type in the database.
	Label = "org_price"
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
	// FieldPrice holds the string denoting the price field in the database.
	FieldPrice = "price"
	// FieldStripePriceID holds the string denoting the stripe_price_id field in the database.
	FieldStripePriceID = "stripe_price_id"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldActive holds the string denoting the active field in the database.
	FieldActive = "active"
	// FieldProductID holds the string denoting the product_id field in the database.
	FieldProductID = "product_id"
	// FieldSubscriptionID holds the string denoting the subscription_id field in the database.
	FieldSubscriptionID = "subscription_id"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// EdgeOrgProducts holds the string denoting the org_products edge name in mutations.
	EdgeOrgProducts = "org_products"
	// EdgeOrgModules holds the string denoting the org_modules edge name in mutations.
	EdgeOrgModules = "org_modules"
	// EdgeOrgSubscription holds the string denoting the org_subscription edge name in mutations.
	EdgeOrgSubscription = "org_subscription"
	// Table holds the table name of the orgprice in the database.
	Table = "org_prices"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "org_prices"
	// OwnerInverseTable is the table name for the Organization entity.
	// It exists in this package in order to avoid circular dependency with the "organization" package.
	OwnerInverseTable = "organizations"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "owner_id"
	// OrgProductsTable is the table that holds the org_products relation/edge. The primary key declared below.
	OrgProductsTable = "org_product_org_prices"
	// OrgProductsInverseTable is the table name for the OrgProduct entity.
	// It exists in this package in order to avoid circular dependency with the "orgproduct" package.
	OrgProductsInverseTable = "org_products"
	// OrgModulesTable is the table that holds the org_modules relation/edge. The primary key declared below.
	OrgModulesTable = "org_module_org_prices"
	// OrgModulesInverseTable is the table name for the OrgModule entity.
	// It exists in this package in order to avoid circular dependency with the "orgmodule" package.
	OrgModulesInverseTable = "org_modules"
	// OrgSubscriptionTable is the table that holds the org_subscription relation/edge.
	OrgSubscriptionTable = "org_prices"
	// OrgSubscriptionInverseTable is the table name for the OrgSubscription entity.
	// It exists in this package in order to avoid circular dependency with the "orgsubscription" package.
	OrgSubscriptionInverseTable = "org_subscriptions"
	// OrgSubscriptionColumn is the table column denoting the org_subscription relation/edge.
	OrgSubscriptionColumn = "subscription_id"
)

// Columns holds all SQL columns for orgprice fields.
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
	FieldPrice,
	FieldStripePriceID,
	FieldStatus,
	FieldActive,
	FieldProductID,
	FieldSubscriptionID,
}

var (
	// OrgProductsPrimaryKey and OrgProductsColumn2 are the table columns denoting the
	// primary key for the org_products relation (M2M).
	OrgProductsPrimaryKey = []string{"org_product_id", "org_price_id"}
	// OrgModulesPrimaryKey and OrgModulesColumn2 are the table columns denoting the
	// primary key for the org_modules relation (M2M).
	OrgModulesPrimaryKey = []string{"org_module_id", "org_price_id"}
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
	Hooks        [4]ent.Hook
	Interceptors [3]ent.Interceptor
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
	// DefaultActive holds the default value on creation for the "active" field.
	DefaultActive bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

// OrderOption defines the ordering options for the OrgPrice queries.
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

// ByStripePriceID orders the results by the stripe_price_id field.
func ByStripePriceID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripePriceID, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByActive orders the results by the active field.
func ByActive(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldActive, opts...).ToFunc()
}

// ByProductID orders the results by the product_id field.
func ByProductID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProductID, opts...).ToFunc()
}

// BySubscriptionID orders the results by the subscription_id field.
func BySubscriptionID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSubscriptionID, opts...).ToFunc()
}

// ByOwnerField orders the results by owner field.
func ByOwnerField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOwnerStep(), sql.OrderByField(field, opts...))
	}
}

// ByOrgProductsCount orders the results by org_products count.
func ByOrgProductsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOrgProductsStep(), opts...)
	}
}

// ByOrgProducts orders the results by org_products terms.
func ByOrgProducts(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOrgProductsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOrgModulesCount orders the results by org_modules count.
func ByOrgModulesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newOrgModulesStep(), opts...)
	}
}

// ByOrgModules orders the results by org_modules terms.
func ByOrgModules(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOrgModulesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByOrgSubscriptionField orders the results by org_subscription field.
func ByOrgSubscriptionField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOrgSubscriptionStep(), sql.OrderByField(field, opts...))
	}
}
func newOwnerStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OwnerInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
	)
}
func newOrgProductsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OrgProductsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, OrgProductsTable, OrgProductsPrimaryKey...),
	)
}
func newOrgModulesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OrgModulesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, OrgModulesTable, OrgModulesPrimaryKey...),
	)
}
func newOrgSubscriptionStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OrgSubscriptionInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OrgSubscriptionTable, OrgSubscriptionColumn),
	)
}
