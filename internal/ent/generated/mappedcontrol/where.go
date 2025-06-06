// Code generated by ent, DO NOT EDIT.

package mappedcontrol

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/theopenlane/core/internal/ent/generated/predicate"
	"github.com/theopenlane/core/pkg/enums"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContainsFold(FieldID, id))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldUpdatedAt, v))
}

// CreatedBy applies equality check predicate on the "created_by" field. It's identical to CreatedByEQ.
func CreatedBy(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldCreatedBy, v))
}

// UpdatedBy applies equality check predicate on the "updated_by" field. It's identical to UpdatedByEQ.
func UpdatedBy(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldUpdatedBy, v))
}

// DeletedAt applies equality check predicate on the "deleted_at" field. It's identical to DeletedAtEQ.
func DeletedAt(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldDeletedAt, v))
}

// DeletedBy applies equality check predicate on the "deleted_by" field. It's identical to DeletedByEQ.
func DeletedBy(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldDeletedBy, v))
}

// OwnerID applies equality check predicate on the "owner_id" field. It's identical to OwnerIDEQ.
func OwnerID(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldOwnerID, v))
}

// Relation applies equality check predicate on the "relation" field. It's identical to RelationEQ.
func Relation(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldRelation, v))
}

// Confidence applies equality check predicate on the "confidence" field. It's identical to ConfidenceEQ.
func Confidence(v int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldConfidence, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldCreatedAt, v))
}

// CreatedAtIsNil applies the IsNil predicate on the "created_at" field.
func CreatedAtIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldCreatedAt))
}

// CreatedAtNotNil applies the NotNil predicate on the "created_at" field.
func CreatedAtNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldCreatedAt))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldUpdatedAt, v))
}

// UpdatedAtIsNil applies the IsNil predicate on the "updated_at" field.
func UpdatedAtIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldUpdatedAt))
}

// UpdatedAtNotNil applies the NotNil predicate on the "updated_at" field.
func UpdatedAtNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldUpdatedAt))
}

// CreatedByEQ applies the EQ predicate on the "created_by" field.
func CreatedByEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldCreatedBy, v))
}

// CreatedByNEQ applies the NEQ predicate on the "created_by" field.
func CreatedByNEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldCreatedBy, v))
}

// CreatedByIn applies the In predicate on the "created_by" field.
func CreatedByIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldCreatedBy, vs...))
}

// CreatedByNotIn applies the NotIn predicate on the "created_by" field.
func CreatedByNotIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldCreatedBy, vs...))
}

// CreatedByGT applies the GT predicate on the "created_by" field.
func CreatedByGT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldCreatedBy, v))
}

// CreatedByGTE applies the GTE predicate on the "created_by" field.
func CreatedByGTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldCreatedBy, v))
}

// CreatedByLT applies the LT predicate on the "created_by" field.
func CreatedByLT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldCreatedBy, v))
}

// CreatedByLTE applies the LTE predicate on the "created_by" field.
func CreatedByLTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldCreatedBy, v))
}

// CreatedByContains applies the Contains predicate on the "created_by" field.
func CreatedByContains(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContains(FieldCreatedBy, v))
}

// CreatedByHasPrefix applies the HasPrefix predicate on the "created_by" field.
func CreatedByHasPrefix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasPrefix(FieldCreatedBy, v))
}

// CreatedByHasSuffix applies the HasSuffix predicate on the "created_by" field.
func CreatedByHasSuffix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasSuffix(FieldCreatedBy, v))
}

// CreatedByIsNil applies the IsNil predicate on the "created_by" field.
func CreatedByIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldCreatedBy))
}

// CreatedByNotNil applies the NotNil predicate on the "created_by" field.
func CreatedByNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldCreatedBy))
}

// CreatedByEqualFold applies the EqualFold predicate on the "created_by" field.
func CreatedByEqualFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEqualFold(FieldCreatedBy, v))
}

// CreatedByContainsFold applies the ContainsFold predicate on the "created_by" field.
func CreatedByContainsFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContainsFold(FieldCreatedBy, v))
}

// UpdatedByEQ applies the EQ predicate on the "updated_by" field.
func UpdatedByEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldUpdatedBy, v))
}

// UpdatedByNEQ applies the NEQ predicate on the "updated_by" field.
func UpdatedByNEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldUpdatedBy, v))
}

// UpdatedByIn applies the In predicate on the "updated_by" field.
func UpdatedByIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldUpdatedBy, vs...))
}

// UpdatedByNotIn applies the NotIn predicate on the "updated_by" field.
func UpdatedByNotIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldUpdatedBy, vs...))
}

// UpdatedByGT applies the GT predicate on the "updated_by" field.
func UpdatedByGT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldUpdatedBy, v))
}

// UpdatedByGTE applies the GTE predicate on the "updated_by" field.
func UpdatedByGTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldUpdatedBy, v))
}

// UpdatedByLT applies the LT predicate on the "updated_by" field.
func UpdatedByLT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldUpdatedBy, v))
}

// UpdatedByLTE applies the LTE predicate on the "updated_by" field.
func UpdatedByLTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldUpdatedBy, v))
}

// UpdatedByContains applies the Contains predicate on the "updated_by" field.
func UpdatedByContains(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContains(FieldUpdatedBy, v))
}

// UpdatedByHasPrefix applies the HasPrefix predicate on the "updated_by" field.
func UpdatedByHasPrefix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasPrefix(FieldUpdatedBy, v))
}

// UpdatedByHasSuffix applies the HasSuffix predicate on the "updated_by" field.
func UpdatedByHasSuffix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasSuffix(FieldUpdatedBy, v))
}

// UpdatedByIsNil applies the IsNil predicate on the "updated_by" field.
func UpdatedByIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldUpdatedBy))
}

// UpdatedByNotNil applies the NotNil predicate on the "updated_by" field.
func UpdatedByNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldUpdatedBy))
}

// UpdatedByEqualFold applies the EqualFold predicate on the "updated_by" field.
func UpdatedByEqualFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEqualFold(FieldUpdatedBy, v))
}

// UpdatedByContainsFold applies the ContainsFold predicate on the "updated_by" field.
func UpdatedByContainsFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContainsFold(FieldUpdatedBy, v))
}

// DeletedAtEQ applies the EQ predicate on the "deleted_at" field.
func DeletedAtEQ(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldDeletedAt, v))
}

// DeletedAtNEQ applies the NEQ predicate on the "deleted_at" field.
func DeletedAtNEQ(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldDeletedAt, v))
}

// DeletedAtIn applies the In predicate on the "deleted_at" field.
func DeletedAtIn(vs ...time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldDeletedAt, vs...))
}

// DeletedAtNotIn applies the NotIn predicate on the "deleted_at" field.
func DeletedAtNotIn(vs ...time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldDeletedAt, vs...))
}

// DeletedAtGT applies the GT predicate on the "deleted_at" field.
func DeletedAtGT(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldDeletedAt, v))
}

// DeletedAtGTE applies the GTE predicate on the "deleted_at" field.
func DeletedAtGTE(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldDeletedAt, v))
}

// DeletedAtLT applies the LT predicate on the "deleted_at" field.
func DeletedAtLT(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldDeletedAt, v))
}

// DeletedAtLTE applies the LTE predicate on the "deleted_at" field.
func DeletedAtLTE(v time.Time) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldDeletedAt, v))
}

// DeletedAtIsNil applies the IsNil predicate on the "deleted_at" field.
func DeletedAtIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldDeletedAt))
}

// DeletedAtNotNil applies the NotNil predicate on the "deleted_at" field.
func DeletedAtNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldDeletedAt))
}

// DeletedByEQ applies the EQ predicate on the "deleted_by" field.
func DeletedByEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldDeletedBy, v))
}

// DeletedByNEQ applies the NEQ predicate on the "deleted_by" field.
func DeletedByNEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldDeletedBy, v))
}

// DeletedByIn applies the In predicate on the "deleted_by" field.
func DeletedByIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldDeletedBy, vs...))
}

// DeletedByNotIn applies the NotIn predicate on the "deleted_by" field.
func DeletedByNotIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldDeletedBy, vs...))
}

// DeletedByGT applies the GT predicate on the "deleted_by" field.
func DeletedByGT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldDeletedBy, v))
}

// DeletedByGTE applies the GTE predicate on the "deleted_by" field.
func DeletedByGTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldDeletedBy, v))
}

// DeletedByLT applies the LT predicate on the "deleted_by" field.
func DeletedByLT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldDeletedBy, v))
}

// DeletedByLTE applies the LTE predicate on the "deleted_by" field.
func DeletedByLTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldDeletedBy, v))
}

// DeletedByContains applies the Contains predicate on the "deleted_by" field.
func DeletedByContains(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContains(FieldDeletedBy, v))
}

// DeletedByHasPrefix applies the HasPrefix predicate on the "deleted_by" field.
func DeletedByHasPrefix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasPrefix(FieldDeletedBy, v))
}

// DeletedByHasSuffix applies the HasSuffix predicate on the "deleted_by" field.
func DeletedByHasSuffix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasSuffix(FieldDeletedBy, v))
}

// DeletedByIsNil applies the IsNil predicate on the "deleted_by" field.
func DeletedByIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldDeletedBy))
}

// DeletedByNotNil applies the NotNil predicate on the "deleted_by" field.
func DeletedByNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldDeletedBy))
}

// DeletedByEqualFold applies the EqualFold predicate on the "deleted_by" field.
func DeletedByEqualFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEqualFold(FieldDeletedBy, v))
}

// DeletedByContainsFold applies the ContainsFold predicate on the "deleted_by" field.
func DeletedByContainsFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContainsFold(FieldDeletedBy, v))
}

// TagsIsNil applies the IsNil predicate on the "tags" field.
func TagsIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldTags))
}

// TagsNotNil applies the NotNil predicate on the "tags" field.
func TagsNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldTags))
}

// OwnerIDEQ applies the EQ predicate on the "owner_id" field.
func OwnerIDEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldOwnerID, v))
}

// OwnerIDNEQ applies the NEQ predicate on the "owner_id" field.
func OwnerIDNEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldOwnerID, v))
}

// OwnerIDIn applies the In predicate on the "owner_id" field.
func OwnerIDIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldOwnerID, vs...))
}

// OwnerIDNotIn applies the NotIn predicate on the "owner_id" field.
func OwnerIDNotIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldOwnerID, vs...))
}

// OwnerIDGT applies the GT predicate on the "owner_id" field.
func OwnerIDGT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldOwnerID, v))
}

// OwnerIDGTE applies the GTE predicate on the "owner_id" field.
func OwnerIDGTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldOwnerID, v))
}

// OwnerIDLT applies the LT predicate on the "owner_id" field.
func OwnerIDLT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldOwnerID, v))
}

// OwnerIDLTE applies the LTE predicate on the "owner_id" field.
func OwnerIDLTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldOwnerID, v))
}

// OwnerIDContains applies the Contains predicate on the "owner_id" field.
func OwnerIDContains(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContains(FieldOwnerID, v))
}

// OwnerIDHasPrefix applies the HasPrefix predicate on the "owner_id" field.
func OwnerIDHasPrefix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasPrefix(FieldOwnerID, v))
}

// OwnerIDHasSuffix applies the HasSuffix predicate on the "owner_id" field.
func OwnerIDHasSuffix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasSuffix(FieldOwnerID, v))
}

// OwnerIDIsNil applies the IsNil predicate on the "owner_id" field.
func OwnerIDIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldOwnerID))
}

// OwnerIDNotNil applies the NotNil predicate on the "owner_id" field.
func OwnerIDNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldOwnerID))
}

// OwnerIDEqualFold applies the EqualFold predicate on the "owner_id" field.
func OwnerIDEqualFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEqualFold(FieldOwnerID, v))
}

// OwnerIDContainsFold applies the ContainsFold predicate on the "owner_id" field.
func OwnerIDContainsFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContainsFold(FieldOwnerID, v))
}

// MappingTypeEQ applies the EQ predicate on the "mapping_type" field.
func MappingTypeEQ(v enums.MappingType) predicate.MappedControl {
	vc := v
	return predicate.MappedControl(sql.FieldEQ(FieldMappingType, vc))
}

// MappingTypeNEQ applies the NEQ predicate on the "mapping_type" field.
func MappingTypeNEQ(v enums.MappingType) predicate.MappedControl {
	vc := v
	return predicate.MappedControl(sql.FieldNEQ(FieldMappingType, vc))
}

// MappingTypeIn applies the In predicate on the "mapping_type" field.
func MappingTypeIn(vs ...enums.MappingType) predicate.MappedControl {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.MappedControl(sql.FieldIn(FieldMappingType, v...))
}

// MappingTypeNotIn applies the NotIn predicate on the "mapping_type" field.
func MappingTypeNotIn(vs ...enums.MappingType) predicate.MappedControl {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.MappedControl(sql.FieldNotIn(FieldMappingType, v...))
}

// RelationEQ applies the EQ predicate on the "relation" field.
func RelationEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldRelation, v))
}

// RelationNEQ applies the NEQ predicate on the "relation" field.
func RelationNEQ(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldRelation, v))
}

// RelationIn applies the In predicate on the "relation" field.
func RelationIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldRelation, vs...))
}

// RelationNotIn applies the NotIn predicate on the "relation" field.
func RelationNotIn(vs ...string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldRelation, vs...))
}

// RelationGT applies the GT predicate on the "relation" field.
func RelationGT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldRelation, v))
}

// RelationGTE applies the GTE predicate on the "relation" field.
func RelationGTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldRelation, v))
}

// RelationLT applies the LT predicate on the "relation" field.
func RelationLT(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldRelation, v))
}

// RelationLTE applies the LTE predicate on the "relation" field.
func RelationLTE(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldRelation, v))
}

// RelationContains applies the Contains predicate on the "relation" field.
func RelationContains(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContains(FieldRelation, v))
}

// RelationHasPrefix applies the HasPrefix predicate on the "relation" field.
func RelationHasPrefix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasPrefix(FieldRelation, v))
}

// RelationHasSuffix applies the HasSuffix predicate on the "relation" field.
func RelationHasSuffix(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldHasSuffix(FieldRelation, v))
}

// RelationIsNil applies the IsNil predicate on the "relation" field.
func RelationIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldRelation))
}

// RelationNotNil applies the NotNil predicate on the "relation" field.
func RelationNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldRelation))
}

// RelationEqualFold applies the EqualFold predicate on the "relation" field.
func RelationEqualFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEqualFold(FieldRelation, v))
}

// RelationContainsFold applies the ContainsFold predicate on the "relation" field.
func RelationContainsFold(v string) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldContainsFold(FieldRelation, v))
}

// ConfidenceEQ applies the EQ predicate on the "confidence" field.
func ConfidenceEQ(v int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldEQ(FieldConfidence, v))
}

// ConfidenceNEQ applies the NEQ predicate on the "confidence" field.
func ConfidenceNEQ(v int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNEQ(FieldConfidence, v))
}

// ConfidenceIn applies the In predicate on the "confidence" field.
func ConfidenceIn(vs ...int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIn(FieldConfidence, vs...))
}

// ConfidenceNotIn applies the NotIn predicate on the "confidence" field.
func ConfidenceNotIn(vs ...int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotIn(FieldConfidence, vs...))
}

// ConfidenceGT applies the GT predicate on the "confidence" field.
func ConfidenceGT(v int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGT(FieldConfidence, v))
}

// ConfidenceGTE applies the GTE predicate on the "confidence" field.
func ConfidenceGTE(v int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldGTE(FieldConfidence, v))
}

// ConfidenceLT applies the LT predicate on the "confidence" field.
func ConfidenceLT(v int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLT(FieldConfidence, v))
}

// ConfidenceLTE applies the LTE predicate on the "confidence" field.
func ConfidenceLTE(v int) predicate.MappedControl {
	return predicate.MappedControl(sql.FieldLTE(FieldConfidence, v))
}

// ConfidenceIsNil applies the IsNil predicate on the "confidence" field.
func ConfidenceIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldConfidence))
}

// ConfidenceNotNil applies the NotNil predicate on the "confidence" field.
func ConfidenceNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldConfidence))
}

// SourceEQ applies the EQ predicate on the "source" field.
func SourceEQ(v enums.MappingSource) predicate.MappedControl {
	vc := v
	return predicate.MappedControl(sql.FieldEQ(FieldSource, vc))
}

// SourceNEQ applies the NEQ predicate on the "source" field.
func SourceNEQ(v enums.MappingSource) predicate.MappedControl {
	vc := v
	return predicate.MappedControl(sql.FieldNEQ(FieldSource, vc))
}

// SourceIn applies the In predicate on the "source" field.
func SourceIn(vs ...enums.MappingSource) predicate.MappedControl {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.MappedControl(sql.FieldIn(FieldSource, v...))
}

// SourceNotIn applies the NotIn predicate on the "source" field.
func SourceNotIn(vs ...enums.MappingSource) predicate.MappedControl {
	v := make([]any, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.MappedControl(sql.FieldNotIn(FieldSource, v...))
}

// SourceIsNil applies the IsNil predicate on the "source" field.
func SourceIsNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldIsNull(FieldSource))
}

// SourceNotNil applies the NotNil predicate on the "source" field.
func SourceNotNil() predicate.MappedControl {
	return predicate.MappedControl(sql.FieldNotNull(FieldSource))
}

// HasOwner applies the HasEdge predicate on the "owner" edge.
func HasOwner() predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Organization
		step.Edge.Schema = schemaConfig.MappedControl
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasOwnerWith applies the HasEdge predicate on the "owner" edge with a given conditions (other predicates).
func HasOwnerWith(preds ...predicate.Organization) predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := newOwnerStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Organization
		step.Edge.Schema = schemaConfig.MappedControl
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasBlockedGroups applies the HasEdge predicate on the "blocked_groups" edge.
func HasBlockedGroups() predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, BlockedGroupsTable, BlockedGroupsPrimaryKey...),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Group
		step.Edge.Schema = schemaConfig.MappedControlBlockedGroups
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasBlockedGroupsWith applies the HasEdge predicate on the "blocked_groups" edge with a given conditions (other predicates).
func HasBlockedGroupsWith(preds ...predicate.Group) predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := newBlockedGroupsStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Group
		step.Edge.Schema = schemaConfig.MappedControlBlockedGroups
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasEditors applies the HasEdge predicate on the "editors" edge.
func HasEditors() predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, EditorsTable, EditorsPrimaryKey...),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Group
		step.Edge.Schema = schemaConfig.MappedControlEditors
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasEditorsWith applies the HasEdge predicate on the "editors" edge with a given conditions (other predicates).
func HasEditorsWith(preds ...predicate.Group) predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := newEditorsStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Group
		step.Edge.Schema = schemaConfig.MappedControlEditors
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasFromControls applies the HasEdge predicate on the "from_controls" edge.
func HasFromControls() predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, FromControlsTable, FromControlsPrimaryKey...),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Control
		step.Edge.Schema = schemaConfig.MappedControlFromControls
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasFromControlsWith applies the HasEdge predicate on the "from_controls" edge with a given conditions (other predicates).
func HasFromControlsWith(preds ...predicate.Control) predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := newFromControlsStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Control
		step.Edge.Schema = schemaConfig.MappedControlFromControls
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasToControls applies the HasEdge predicate on the "to_controls" edge.
func HasToControls() predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, ToControlsTable, ToControlsPrimaryKey...),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Control
		step.Edge.Schema = schemaConfig.MappedControlToControls
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasToControlsWith applies the HasEdge predicate on the "to_controls" edge with a given conditions (other predicates).
func HasToControlsWith(preds ...predicate.Control) predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := newToControlsStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Control
		step.Edge.Schema = schemaConfig.MappedControlToControls
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasFromSubcontrols applies the HasEdge predicate on the "from_subcontrols" edge.
func HasFromSubcontrols() predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, FromSubcontrolsTable, FromSubcontrolsPrimaryKey...),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Subcontrol
		step.Edge.Schema = schemaConfig.MappedControlFromSubcontrols
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasFromSubcontrolsWith applies the HasEdge predicate on the "from_subcontrols" edge with a given conditions (other predicates).
func HasFromSubcontrolsWith(preds ...predicate.Subcontrol) predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := newFromSubcontrolsStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Subcontrol
		step.Edge.Schema = schemaConfig.MappedControlFromSubcontrols
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasToSubcontrols applies the HasEdge predicate on the "to_subcontrols" edge.
func HasToSubcontrols() predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, false, ToSubcontrolsTable, ToSubcontrolsPrimaryKey...),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Subcontrol
		step.Edge.Schema = schemaConfig.MappedControlToSubcontrols
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasToSubcontrolsWith applies the HasEdge predicate on the "to_subcontrols" edge with a given conditions (other predicates).
func HasToSubcontrolsWith(preds ...predicate.Subcontrol) predicate.MappedControl {
	return predicate.MappedControl(func(s *sql.Selector) {
		step := newToSubcontrolsStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.Subcontrol
		step.Edge.Schema = schemaConfig.MappedControlToSubcontrols
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.MappedControl) predicate.MappedControl {
	return predicate.MappedControl(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.MappedControl) predicate.MappedControl {
	return predicate.MappedControl(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.MappedControl) predicate.MappedControl {
	return predicate.MappedControl(sql.NotPredicates(p))
}
