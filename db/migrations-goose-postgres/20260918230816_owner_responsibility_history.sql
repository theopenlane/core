-- +goose Up
-- modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "internal_owner" character varying NULL, ADD COLUMN "internal_owner_user_id" character varying NULL, ADD COLUMN "internal_owner_group_id" character varying NULL, ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "stakeholder_name" character varying NULL, ADD COLUMN "stakeholder_user_id" character varying NULL, ADD COLUMN "stakeholder_group_id" character varying NULL, ADD COLUMN "stakeholder_identity_holder_id" character varying NULL, ADD COLUMN "delegate_name" character varying NULL, ADD COLUMN "delegate_user_id" character varying NULL, ADD COLUMN "delegate_group_id" character varying NULL, ADD COLUMN "delegate_identity_holder_id" character varying NULL;
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "internal_owner" character varying NULL, ADD COLUMN "internal_owner_user_id" character varying NULL, ADD COLUMN "internal_owner_group_id" character varying NULL, ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;

-- +goose Down
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "internal_owner_identity_holder_id", DROP COLUMN "internal_owner_group_id", DROP COLUMN "internal_owner_user_id", DROP COLUMN "internal_owner";
-- reverse: modify "risk_history" table
ALTER TABLE "risk_history" DROP COLUMN "delegate_identity_holder_id", DROP COLUMN "delegate_group_id", DROP COLUMN "delegate_user_id", DROP COLUMN "delegate_name", DROP COLUMN "stakeholder_identity_holder_id", DROP COLUMN "stakeholder_group_id", DROP COLUMN "stakeholder_user_id", DROP COLUMN "stakeholder_name";
-- reverse: modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "internal_owner_identity_holder_id", DROP COLUMN "internal_owner_group_id", DROP COLUMN "internal_owner_user_id", DROP COLUMN "internal_owner";
