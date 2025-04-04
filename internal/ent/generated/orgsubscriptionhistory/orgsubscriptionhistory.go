// Code generated by ent, DO NOT EDIT.

package orgsubscriptionhistory

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/entx/history"
)

const (
	// Label holds the string label denoting the orgsubscriptionhistory type in the database.
	Label = "org_subscription_history"
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
	// FieldTags holds the string denoting the tags field in the database.
	FieldTags = "tags"
	// FieldOwnerID holds the string denoting the owner_id field in the database.
	FieldOwnerID = "owner_id"
	// FieldStripeSubscriptionID holds the string denoting the stripe_subscription_id field in the database.
	FieldStripeSubscriptionID = "stripe_subscription_id"
	// FieldProductTier holds the string denoting the product_tier field in the database.
	FieldProductTier = "product_tier"
	// FieldProductPrice holds the string denoting the product_price field in the database.
	FieldProductPrice = "product_price"
	// FieldStripeProductTierID holds the string denoting the stripe_product_tier_id field in the database.
	FieldStripeProductTierID = "stripe_product_tier_id"
	// FieldStripeSubscriptionStatus holds the string denoting the stripe_subscription_status field in the database.
	FieldStripeSubscriptionStatus = "stripe_subscription_status"
	// FieldActive holds the string denoting the active field in the database.
	FieldActive = "active"
	// FieldStripeCustomerID holds the string denoting the stripe_customer_id field in the database.
	FieldStripeCustomerID = "stripe_customer_id"
	// FieldExpiresAt holds the string denoting the expires_at field in the database.
	FieldExpiresAt = "expires_at"
	// FieldTrialExpiresAt holds the string denoting the trial_expires_at field in the database.
	FieldTrialExpiresAt = "trial_expires_at"
	// FieldDaysUntilDue holds the string denoting the days_until_due field in the database.
	FieldDaysUntilDue = "days_until_due"
	// FieldPaymentMethodAdded holds the string denoting the payment_method_added field in the database.
	FieldPaymentMethodAdded = "payment_method_added"
	// FieldFeatures holds the string denoting the features field in the database.
	FieldFeatures = "features"
	// FieldFeatureLookupKeys holds the string denoting the feature_lookup_keys field in the database.
	FieldFeatureLookupKeys = "feature_lookup_keys"
	// Table holds the table name of the orgsubscriptionhistory in the database.
	Table = "org_subscription_history"
)

// Columns holds all SQL columns for orgsubscriptionhistory fields.
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
	FieldTags,
	FieldOwnerID,
	FieldStripeSubscriptionID,
	FieldProductTier,
	FieldProductPrice,
	FieldStripeProductTierID,
	FieldStripeSubscriptionStatus,
	FieldActive,
	FieldStripeCustomerID,
	FieldExpiresAt,
	FieldTrialExpiresAt,
	FieldDaysUntilDue,
	FieldPaymentMethodAdded,
	FieldFeatures,
	FieldFeatureLookupKeys,
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

var (
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
	// DefaultActive holds the default value on creation for the "active" field.
	DefaultActive bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() string
)

// OperationValidator is a validator for the "operation" field enum values. It is called by the builders before save.
func OperationValidator(o history.OpType) error {
	switch o.String() {
	case "INSERT", "UPDATE", "DELETE":
		return nil
	default:
		return fmt.Errorf("orgsubscriptionhistory: invalid enum value for operation field: %q", o)
	}
}

// OrderOption defines the ordering options for the OrgSubscriptionHistory queries.
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

// ByOwnerID orders the results by the owner_id field.
func ByOwnerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOwnerID, opts...).ToFunc()
}

// ByStripeSubscriptionID orders the results by the stripe_subscription_id field.
func ByStripeSubscriptionID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeSubscriptionID, opts...).ToFunc()
}

// ByProductTier orders the results by the product_tier field.
func ByProductTier(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProductTier, opts...).ToFunc()
}

// ByStripeProductTierID orders the results by the stripe_product_tier_id field.
func ByStripeProductTierID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeProductTierID, opts...).ToFunc()
}

// ByStripeSubscriptionStatus orders the results by the stripe_subscription_status field.
func ByStripeSubscriptionStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeSubscriptionStatus, opts...).ToFunc()
}

// ByActive orders the results by the active field.
func ByActive(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldActive, opts...).ToFunc()
}

// ByStripeCustomerID orders the results by the stripe_customer_id field.
func ByStripeCustomerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeCustomerID, opts...).ToFunc()
}

// ByExpiresAt orders the results by the expires_at field.
func ByExpiresAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldExpiresAt, opts...).ToFunc()
}

// ByTrialExpiresAt orders the results by the trial_expires_at field.
func ByTrialExpiresAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTrialExpiresAt, opts...).ToFunc()
}

// ByDaysUntilDue orders the results by the days_until_due field.
func ByDaysUntilDue(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDaysUntilDue, opts...).ToFunc()
}

// ByPaymentMethodAdded orders the results by the payment_method_added field.
func ByPaymentMethodAdded(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPaymentMethodAdded, opts...).ToFunc()
}

var (
	// history.OpType must implement graphql.Marshaler.
	_ graphql.Marshaler = (*history.OpType)(nil)
	// history.OpType must implement graphql.Unmarshaler.
	_ graphql.Unmarshaler = (*history.OpType)(nil)
)
