// Code generated by ent, DO NOT EDIT.

package webauthn

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/theopenlane/core/internal/ent/generated/predicate"

	"github.com/theopenlane/core/internal/ent/generated/internal"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldID, id))
}

// IDEqualFold applies the EqualFold predicate on the ID field.
func IDEqualFold(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEqualFold(FieldID, id))
}

// IDContainsFold applies the ContainsFold predicate on the ID field.
func IDContainsFold(id string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContainsFold(FieldID, id))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUpdatedAt, v))
}

// CreatedBy applies equality check predicate on the "created_by" field. It's identical to CreatedByEQ.
func CreatedBy(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldCreatedBy, v))
}

// UpdatedBy applies equality check predicate on the "updated_by" field. It's identical to UpdatedByEQ.
func UpdatedBy(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUpdatedBy, v))
}

// MappingID applies equality check predicate on the "mapping_id" field. It's identical to MappingIDEQ.
func MappingID(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldMappingID, v))
}

// OwnerID applies equality check predicate on the "owner_id" field. It's identical to OwnerIDEQ.
func OwnerID(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldOwnerID, v))
}

// CredentialID applies equality check predicate on the "credential_id" field. It's identical to CredentialIDEQ.
func CredentialID(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldCredentialID, v))
}

// PublicKey applies equality check predicate on the "public_key" field. It's identical to PublicKeyEQ.
func PublicKey(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldPublicKey, v))
}

// AttestationType applies equality check predicate on the "attestation_type" field. It's identical to AttestationTypeEQ.
func AttestationType(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldAttestationType, v))
}

// Aaguid applies equality check predicate on the "aaguid" field. It's identical to AaguidEQ.
func Aaguid(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldAaguid, v))
}

// SignCount applies equality check predicate on the "sign_count" field. It's identical to SignCountEQ.
func SignCount(v int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldSignCount, v))
}

// BackupEligible applies equality check predicate on the "backup_eligible" field. It's identical to BackupEligibleEQ.
func BackupEligible(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldBackupEligible, v))
}

// BackupState applies equality check predicate on the "backup_state" field. It's identical to BackupStateEQ.
func BackupState(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldBackupState, v))
}

// UserPresent applies equality check predicate on the "user_present" field. It's identical to UserPresentEQ.
func UserPresent(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUserPresent, v))
}

// UserVerified applies equality check predicate on the "user_verified" field. It's identical to UserVerifiedEQ.
func UserVerified(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUserVerified, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldCreatedAt, v))
}

// CreatedAtIsNil applies the IsNil predicate on the "created_at" field.
func CreatedAtIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldCreatedAt))
}

// CreatedAtNotNil applies the NotNil predicate on the "created_at" field.
func CreatedAtNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldCreatedAt))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldUpdatedAt, v))
}

// UpdatedAtIsNil applies the IsNil predicate on the "updated_at" field.
func UpdatedAtIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldUpdatedAt))
}

// UpdatedAtNotNil applies the NotNil predicate on the "updated_at" field.
func UpdatedAtNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldUpdatedAt))
}

// CreatedByEQ applies the EQ predicate on the "created_by" field.
func CreatedByEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldCreatedBy, v))
}

// CreatedByNEQ applies the NEQ predicate on the "created_by" field.
func CreatedByNEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldCreatedBy, v))
}

// CreatedByIn applies the In predicate on the "created_by" field.
func CreatedByIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldCreatedBy, vs...))
}

// CreatedByNotIn applies the NotIn predicate on the "created_by" field.
func CreatedByNotIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldCreatedBy, vs...))
}

// CreatedByGT applies the GT predicate on the "created_by" field.
func CreatedByGT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldCreatedBy, v))
}

// CreatedByGTE applies the GTE predicate on the "created_by" field.
func CreatedByGTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldCreatedBy, v))
}

// CreatedByLT applies the LT predicate on the "created_by" field.
func CreatedByLT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldCreatedBy, v))
}

// CreatedByLTE applies the LTE predicate on the "created_by" field.
func CreatedByLTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldCreatedBy, v))
}

// CreatedByContains applies the Contains predicate on the "created_by" field.
func CreatedByContains(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContains(FieldCreatedBy, v))
}

// CreatedByHasPrefix applies the HasPrefix predicate on the "created_by" field.
func CreatedByHasPrefix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasPrefix(FieldCreatedBy, v))
}

// CreatedByHasSuffix applies the HasSuffix predicate on the "created_by" field.
func CreatedByHasSuffix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasSuffix(FieldCreatedBy, v))
}

// CreatedByIsNil applies the IsNil predicate on the "created_by" field.
func CreatedByIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldCreatedBy))
}

// CreatedByNotNil applies the NotNil predicate on the "created_by" field.
func CreatedByNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldCreatedBy))
}

// CreatedByEqualFold applies the EqualFold predicate on the "created_by" field.
func CreatedByEqualFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEqualFold(FieldCreatedBy, v))
}

// CreatedByContainsFold applies the ContainsFold predicate on the "created_by" field.
func CreatedByContainsFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContainsFold(FieldCreatedBy, v))
}

// UpdatedByEQ applies the EQ predicate on the "updated_by" field.
func UpdatedByEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUpdatedBy, v))
}

// UpdatedByNEQ applies the NEQ predicate on the "updated_by" field.
func UpdatedByNEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldUpdatedBy, v))
}

// UpdatedByIn applies the In predicate on the "updated_by" field.
func UpdatedByIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldUpdatedBy, vs...))
}

// UpdatedByNotIn applies the NotIn predicate on the "updated_by" field.
func UpdatedByNotIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldUpdatedBy, vs...))
}

// UpdatedByGT applies the GT predicate on the "updated_by" field.
func UpdatedByGT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldUpdatedBy, v))
}

// UpdatedByGTE applies the GTE predicate on the "updated_by" field.
func UpdatedByGTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldUpdatedBy, v))
}

// UpdatedByLT applies the LT predicate on the "updated_by" field.
func UpdatedByLT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldUpdatedBy, v))
}

// UpdatedByLTE applies the LTE predicate on the "updated_by" field.
func UpdatedByLTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldUpdatedBy, v))
}

// UpdatedByContains applies the Contains predicate on the "updated_by" field.
func UpdatedByContains(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContains(FieldUpdatedBy, v))
}

// UpdatedByHasPrefix applies the HasPrefix predicate on the "updated_by" field.
func UpdatedByHasPrefix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasPrefix(FieldUpdatedBy, v))
}

// UpdatedByHasSuffix applies the HasSuffix predicate on the "updated_by" field.
func UpdatedByHasSuffix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasSuffix(FieldUpdatedBy, v))
}

// UpdatedByIsNil applies the IsNil predicate on the "updated_by" field.
func UpdatedByIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldUpdatedBy))
}

// UpdatedByNotNil applies the NotNil predicate on the "updated_by" field.
func UpdatedByNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldUpdatedBy))
}

// UpdatedByEqualFold applies the EqualFold predicate on the "updated_by" field.
func UpdatedByEqualFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEqualFold(FieldUpdatedBy, v))
}

// UpdatedByContainsFold applies the ContainsFold predicate on the "updated_by" field.
func UpdatedByContainsFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContainsFold(FieldUpdatedBy, v))
}

// MappingIDEQ applies the EQ predicate on the "mapping_id" field.
func MappingIDEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldMappingID, v))
}

// MappingIDNEQ applies the NEQ predicate on the "mapping_id" field.
func MappingIDNEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldMappingID, v))
}

// MappingIDIn applies the In predicate on the "mapping_id" field.
func MappingIDIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldMappingID, vs...))
}

// MappingIDNotIn applies the NotIn predicate on the "mapping_id" field.
func MappingIDNotIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldMappingID, vs...))
}

// MappingIDGT applies the GT predicate on the "mapping_id" field.
func MappingIDGT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldMappingID, v))
}

// MappingIDGTE applies the GTE predicate on the "mapping_id" field.
func MappingIDGTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldMappingID, v))
}

// MappingIDLT applies the LT predicate on the "mapping_id" field.
func MappingIDLT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldMappingID, v))
}

// MappingIDLTE applies the LTE predicate on the "mapping_id" field.
func MappingIDLTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldMappingID, v))
}

// MappingIDContains applies the Contains predicate on the "mapping_id" field.
func MappingIDContains(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContains(FieldMappingID, v))
}

// MappingIDHasPrefix applies the HasPrefix predicate on the "mapping_id" field.
func MappingIDHasPrefix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasPrefix(FieldMappingID, v))
}

// MappingIDHasSuffix applies the HasSuffix predicate on the "mapping_id" field.
func MappingIDHasSuffix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasSuffix(FieldMappingID, v))
}

// MappingIDEqualFold applies the EqualFold predicate on the "mapping_id" field.
func MappingIDEqualFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEqualFold(FieldMappingID, v))
}

// MappingIDContainsFold applies the ContainsFold predicate on the "mapping_id" field.
func MappingIDContainsFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContainsFold(FieldMappingID, v))
}

// TagsIsNil applies the IsNil predicate on the "tags" field.
func TagsIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldTags))
}

// TagsNotNil applies the NotNil predicate on the "tags" field.
func TagsNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldTags))
}

// OwnerIDEQ applies the EQ predicate on the "owner_id" field.
func OwnerIDEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldOwnerID, v))
}

// OwnerIDNEQ applies the NEQ predicate on the "owner_id" field.
func OwnerIDNEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldOwnerID, v))
}

// OwnerIDIn applies the In predicate on the "owner_id" field.
func OwnerIDIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldOwnerID, vs...))
}

// OwnerIDNotIn applies the NotIn predicate on the "owner_id" field.
func OwnerIDNotIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldOwnerID, vs...))
}

// OwnerIDGT applies the GT predicate on the "owner_id" field.
func OwnerIDGT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldOwnerID, v))
}

// OwnerIDGTE applies the GTE predicate on the "owner_id" field.
func OwnerIDGTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldOwnerID, v))
}

// OwnerIDLT applies the LT predicate on the "owner_id" field.
func OwnerIDLT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldOwnerID, v))
}

// OwnerIDLTE applies the LTE predicate on the "owner_id" field.
func OwnerIDLTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldOwnerID, v))
}

// OwnerIDContains applies the Contains predicate on the "owner_id" field.
func OwnerIDContains(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContains(FieldOwnerID, v))
}

// OwnerIDHasPrefix applies the HasPrefix predicate on the "owner_id" field.
func OwnerIDHasPrefix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasPrefix(FieldOwnerID, v))
}

// OwnerIDHasSuffix applies the HasSuffix predicate on the "owner_id" field.
func OwnerIDHasSuffix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasSuffix(FieldOwnerID, v))
}

// OwnerIDEqualFold applies the EqualFold predicate on the "owner_id" field.
func OwnerIDEqualFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEqualFold(FieldOwnerID, v))
}

// OwnerIDContainsFold applies the ContainsFold predicate on the "owner_id" field.
func OwnerIDContainsFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContainsFold(FieldOwnerID, v))
}

// CredentialIDEQ applies the EQ predicate on the "credential_id" field.
func CredentialIDEQ(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldCredentialID, v))
}

// CredentialIDNEQ applies the NEQ predicate on the "credential_id" field.
func CredentialIDNEQ(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldCredentialID, v))
}

// CredentialIDIn applies the In predicate on the "credential_id" field.
func CredentialIDIn(vs ...[]byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldCredentialID, vs...))
}

// CredentialIDNotIn applies the NotIn predicate on the "credential_id" field.
func CredentialIDNotIn(vs ...[]byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldCredentialID, vs...))
}

// CredentialIDGT applies the GT predicate on the "credential_id" field.
func CredentialIDGT(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldCredentialID, v))
}

// CredentialIDGTE applies the GTE predicate on the "credential_id" field.
func CredentialIDGTE(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldCredentialID, v))
}

// CredentialIDLT applies the LT predicate on the "credential_id" field.
func CredentialIDLT(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldCredentialID, v))
}

// CredentialIDLTE applies the LTE predicate on the "credential_id" field.
func CredentialIDLTE(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldCredentialID, v))
}

// CredentialIDIsNil applies the IsNil predicate on the "credential_id" field.
func CredentialIDIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldCredentialID))
}

// CredentialIDNotNil applies the NotNil predicate on the "credential_id" field.
func CredentialIDNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldCredentialID))
}

// PublicKeyEQ applies the EQ predicate on the "public_key" field.
func PublicKeyEQ(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldPublicKey, v))
}

// PublicKeyNEQ applies the NEQ predicate on the "public_key" field.
func PublicKeyNEQ(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldPublicKey, v))
}

// PublicKeyIn applies the In predicate on the "public_key" field.
func PublicKeyIn(vs ...[]byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldPublicKey, vs...))
}

// PublicKeyNotIn applies the NotIn predicate on the "public_key" field.
func PublicKeyNotIn(vs ...[]byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldPublicKey, vs...))
}

// PublicKeyGT applies the GT predicate on the "public_key" field.
func PublicKeyGT(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldPublicKey, v))
}

// PublicKeyGTE applies the GTE predicate on the "public_key" field.
func PublicKeyGTE(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldPublicKey, v))
}

// PublicKeyLT applies the LT predicate on the "public_key" field.
func PublicKeyLT(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldPublicKey, v))
}

// PublicKeyLTE applies the LTE predicate on the "public_key" field.
func PublicKeyLTE(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldPublicKey, v))
}

// PublicKeyIsNil applies the IsNil predicate on the "public_key" field.
func PublicKeyIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldPublicKey))
}

// PublicKeyNotNil applies the NotNil predicate on the "public_key" field.
func PublicKeyNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldPublicKey))
}

// AttestationTypeEQ applies the EQ predicate on the "attestation_type" field.
func AttestationTypeEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldAttestationType, v))
}

// AttestationTypeNEQ applies the NEQ predicate on the "attestation_type" field.
func AttestationTypeNEQ(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldAttestationType, v))
}

// AttestationTypeIn applies the In predicate on the "attestation_type" field.
func AttestationTypeIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldAttestationType, vs...))
}

// AttestationTypeNotIn applies the NotIn predicate on the "attestation_type" field.
func AttestationTypeNotIn(vs ...string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldAttestationType, vs...))
}

// AttestationTypeGT applies the GT predicate on the "attestation_type" field.
func AttestationTypeGT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldAttestationType, v))
}

// AttestationTypeGTE applies the GTE predicate on the "attestation_type" field.
func AttestationTypeGTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldAttestationType, v))
}

// AttestationTypeLT applies the LT predicate on the "attestation_type" field.
func AttestationTypeLT(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldAttestationType, v))
}

// AttestationTypeLTE applies the LTE predicate on the "attestation_type" field.
func AttestationTypeLTE(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldAttestationType, v))
}

// AttestationTypeContains applies the Contains predicate on the "attestation_type" field.
func AttestationTypeContains(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContains(FieldAttestationType, v))
}

// AttestationTypeHasPrefix applies the HasPrefix predicate on the "attestation_type" field.
func AttestationTypeHasPrefix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasPrefix(FieldAttestationType, v))
}

// AttestationTypeHasSuffix applies the HasSuffix predicate on the "attestation_type" field.
func AttestationTypeHasSuffix(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldHasSuffix(FieldAttestationType, v))
}

// AttestationTypeIsNil applies the IsNil predicate on the "attestation_type" field.
func AttestationTypeIsNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIsNull(FieldAttestationType))
}

// AttestationTypeNotNil applies the NotNil predicate on the "attestation_type" field.
func AttestationTypeNotNil() predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotNull(FieldAttestationType))
}

// AttestationTypeEqualFold applies the EqualFold predicate on the "attestation_type" field.
func AttestationTypeEqualFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEqualFold(FieldAttestationType, v))
}

// AttestationTypeContainsFold applies the ContainsFold predicate on the "attestation_type" field.
func AttestationTypeContainsFold(v string) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldContainsFold(FieldAttestationType, v))
}

// AaguidEQ applies the EQ predicate on the "aaguid" field.
func AaguidEQ(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldAaguid, v))
}

// AaguidNEQ applies the NEQ predicate on the "aaguid" field.
func AaguidNEQ(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldAaguid, v))
}

// AaguidIn applies the In predicate on the "aaguid" field.
func AaguidIn(vs ...[]byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldAaguid, vs...))
}

// AaguidNotIn applies the NotIn predicate on the "aaguid" field.
func AaguidNotIn(vs ...[]byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldAaguid, vs...))
}

// AaguidGT applies the GT predicate on the "aaguid" field.
func AaguidGT(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldAaguid, v))
}

// AaguidGTE applies the GTE predicate on the "aaguid" field.
func AaguidGTE(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldAaguid, v))
}

// AaguidLT applies the LT predicate on the "aaguid" field.
func AaguidLT(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldAaguid, v))
}

// AaguidLTE applies the LTE predicate on the "aaguid" field.
func AaguidLTE(v []byte) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldAaguid, v))
}

// SignCountEQ applies the EQ predicate on the "sign_count" field.
func SignCountEQ(v int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldSignCount, v))
}

// SignCountNEQ applies the NEQ predicate on the "sign_count" field.
func SignCountNEQ(v int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldSignCount, v))
}

// SignCountIn applies the In predicate on the "sign_count" field.
func SignCountIn(vs ...int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldIn(FieldSignCount, vs...))
}

// SignCountNotIn applies the NotIn predicate on the "sign_count" field.
func SignCountNotIn(vs ...int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNotIn(FieldSignCount, vs...))
}

// SignCountGT applies the GT predicate on the "sign_count" field.
func SignCountGT(v int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGT(FieldSignCount, v))
}

// SignCountGTE applies the GTE predicate on the "sign_count" field.
func SignCountGTE(v int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldGTE(FieldSignCount, v))
}

// SignCountLT applies the LT predicate on the "sign_count" field.
func SignCountLT(v int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLT(FieldSignCount, v))
}

// SignCountLTE applies the LTE predicate on the "sign_count" field.
func SignCountLTE(v int32) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldLTE(FieldSignCount, v))
}

// BackupEligibleEQ applies the EQ predicate on the "backup_eligible" field.
func BackupEligibleEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldBackupEligible, v))
}

// BackupEligibleNEQ applies the NEQ predicate on the "backup_eligible" field.
func BackupEligibleNEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldBackupEligible, v))
}

// BackupStateEQ applies the EQ predicate on the "backup_state" field.
func BackupStateEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldBackupState, v))
}

// BackupStateNEQ applies the NEQ predicate on the "backup_state" field.
func BackupStateNEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldBackupState, v))
}

// UserPresentEQ applies the EQ predicate on the "user_present" field.
func UserPresentEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUserPresent, v))
}

// UserPresentNEQ applies the NEQ predicate on the "user_present" field.
func UserPresentNEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldUserPresent, v))
}

// UserVerifiedEQ applies the EQ predicate on the "user_verified" field.
func UserVerifiedEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldEQ(FieldUserVerified, v))
}

// UserVerifiedNEQ applies the NEQ predicate on the "user_verified" field.
func UserVerifiedNEQ(v bool) predicate.Webauthn {
	return predicate.Webauthn(sql.FieldNEQ(FieldUserVerified, v))
}

// HasOwner applies the HasEdge predicate on the "owner" edge.
func HasOwner() predicate.Webauthn {
	return predicate.Webauthn(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, OwnerTable, OwnerColumn),
		)
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.User
		step.Edge.Schema = schemaConfig.Webauthn
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasOwnerWith applies the HasEdge predicate on the "owner" edge with a given conditions (other predicates).
func HasOwnerWith(preds ...predicate.User) predicate.Webauthn {
	return predicate.Webauthn(func(s *sql.Selector) {
		step := newOwnerStep()
		schemaConfig := internal.SchemaConfigFromContext(s.Context())
		step.To.Schema = schemaConfig.User
		step.Edge.Schema = schemaConfig.Webauthn
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Webauthn) predicate.Webauthn {
	return predicate.Webauthn(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Webauthn) predicate.Webauthn {
	return predicate.Webauthn(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Webauthn) predicate.Webauthn {
	return predicate.Webauthn(sql.NotPredicates(p))
}