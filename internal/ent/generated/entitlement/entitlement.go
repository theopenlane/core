// Code generated by ent, DO NOT EDIT.

package entitlement

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
)

const (
	// Label holds the string label denoting the entitlement type in the database.
	Label = "entitlement"
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
	// FieldMappingID holds the string denoting the mapping_id field in the database.
	FieldMappingID = "mapping_id"
	// FieldTags holds the string denoting the tags field in the database.
	FieldTags = "tags"
	// FieldDeletedAt holds the string denoting the deleted_at field in the database.
	FieldDeletedAt = "deleted_at"
	// FieldDeletedBy holds the string denoting the deleted_by field in the database.
	FieldDeletedBy = "deleted_by"
	// FieldOwnerID holds the string denoting the owner_id field in the database.
	FieldOwnerID = "owner_id"
	// FieldPlanID holds the string denoting the plan_id field in the database.
	FieldPlanID = "plan_id"
	// FieldOrganizationID holds the string denoting the organization_id field in the database.
	FieldOrganizationID = "organization_id"
	// FieldExternalCustomerID holds the string denoting the external_customer_id field in the database.
	FieldExternalCustomerID = "external_customer_id"
	// FieldExternalSubscriptionID holds the string denoting the external_subscription_id field in the database.
	FieldExternalSubscriptionID = "external_subscription_id"
	// FieldExpires holds the string denoting the expires field in the database.
	FieldExpires = "expires"
	// FieldExpiresAt holds the string denoting the expires_at field in the database.
	FieldExpiresAt = "expires_at"
	// FieldCancelled holds the string denoting the cancelled field in the database.
	FieldCancelled = "cancelled"
	// EdgeOwner holds the string denoting the owner edge name in mutations.
	EdgeOwner = "owner"
	// EdgePlan holds the string denoting the plan edge name in mutations.
	EdgePlan = "plan"
	// EdgeOrganization holds the string denoting the organization edge name in mutations.
	EdgeOrganization = "organization"
	// EdgeEvents holds the string denoting the events edge name in mutations.
	EdgeEvents = "events"
	// Table holds the table name of the entitlement in the database.
	Table = "entitlements"
	// OwnerTable is the table that holds the owner relation/edge.
	OwnerTable = "entitlements"
	// OwnerInverseTable is the table name for the Organization entity.
	// It exists in this package in order to avoid circular dependency with the "organization" package.
	OwnerInverseTable = "organizations"
	// OwnerColumn is the table column denoting the owner relation/edge.
	OwnerColumn = "owner_id"
	// PlanTable is the table that holds the plan relation/edge.
	PlanTable = "entitlements"
	// PlanInverseTable is the table name for the EntitlementPlan entity.
	// It exists in this package in order to avoid circular dependency with the "entitlementplan" package.
	PlanInverseTable = "entitlement_plans"
	// PlanColumn is the table column denoting the plan relation/edge.
	PlanColumn = "plan_id"
	// OrganizationTable is the table that holds the organization relation/edge.
	OrganizationTable = "entitlements"
	// OrganizationInverseTable is the table name for the Organization entity.
	// It exists in this package in order to avoid circular dependency with the "organization" package.
	OrganizationInverseTable = "organizations"
	// OrganizationColumn is the table column denoting the organization relation/edge.
	OrganizationColumn = "organization_id"
	// EventsTable is the table that holds the events relation/edge. The primary key declared below.
	EventsTable = "entitlement_events"
	// EventsInverseTable is the table name for the Event entity.
	// It exists in this package in order to avoid circular dependency with the "event" package.
	EventsInverseTable = "events"
)

// Columns holds all SQL columns for entitlement fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldCreatedBy,
	FieldUpdatedBy,
	FieldMappingID,
	FieldTags,
	FieldDeletedAt,
	FieldDeletedBy,
	FieldOwnerID,
	FieldPlanID,
	FieldOrganizationID,
	FieldExternalCustomerID,
	FieldExternalSubscriptionID,
	FieldExpires,
	FieldExpiresAt,
	FieldCancelled,
}

var (
	// EventsPrimaryKey and EventsColumn2 are the table columns denoting the
	// primary key for the events relation (M2M).
	EventsPrimaryKey = []string{"entitlement_id", "event_id"}
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
	Hooks        [5]ent.Hook
	Interceptors [2]ent.Interceptor
	Policy       ent.Policy
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultMappingID holds the default value on creation for the "mapping_id" field.
	DefaultMappingID func() string
	// DefaultTags holds the default value on creation for the "tags" field.
	DefaultTags []string
	// OwnerIDValidator is a validator for the "owner_id" field. It is called by the builders before save.
	OwnerIDValidator func(string) error
	// PlanIDValidator is a validator for the "plan_id" field. It is called by the builders before save.
	PlanIDValidator func(string) error
	// OrganizationIDValidator is a validator for the "organization_id" field. It is called by the builders before save.
	OrganizationIDValidator func(string) error
	// DefaultExpires holds the default value on creation for the "expires" field.
	DefaultExpires bool
	// DefaultCancelled holds the default value on creation for the "cancelled" field.
	DefaultCancelled bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

// OrderOption defines the ordering options for the Entitlement queries.
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

// ByMappingID orders the results by the mapping_id field.
func ByMappingID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldMappingID, opts...).ToFunc()
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

// ByPlanID orders the results by the plan_id field.
func ByPlanID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPlanID, opts...).ToFunc()
}

// ByOrganizationID orders the results by the organization_id field.
func ByOrganizationID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOrganizationID, opts...).ToFunc()
}

// ByExternalCustomerID orders the results by the external_customer_id field.
func ByExternalCustomerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldExternalCustomerID, opts...).ToFunc()
}

// ByExternalSubscriptionID orders the results by the external_subscription_id field.
func ByExternalSubscriptionID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldExternalSubscriptionID, opts...).ToFunc()
}

// ByExpires orders the results by the expires field.
func ByExpires(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldExpires, opts...).ToFunc()
}

// ByExpiresAt orders the results by the expires_at field.
func ByExpiresAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldExpiresAt, opts...).ToFunc()
}

// ByCancelled orders the results by the cancelled field.
func ByCancelled(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCancelled, opts...).ToFunc()
}

// ByOwnerField orders the results by owner field.
func ByOwnerField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOwnerStep(), sql.OrderByField(field, opts...))
	}
}

// ByPlanField orders the results by plan field.
func ByPlanField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newPlanStep(), sql.OrderByField(field, opts...))
	}
}

// ByOrganizationField orders the results by organization field.
func ByOrganizationField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOrganizationStep(), sql.OrderByField(field, opts...))
	}
}

// ByEventsCount orders the results by events count.
func ByEventsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEventsStep(), opts...)
	}
}

// ByEvents orders the results by events terms.
func ByEvents(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEventsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newOwnerStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OwnerInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
	)
}
func newPlanStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(PlanInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, PlanTable, PlanColumn),
	)
}
func newOrganizationStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OrganizationInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, OrganizationTable, OrganizationColumn),
	)
}
func newEventsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EventsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, EventsTable, EventsPrimaryKey...),
	)
}