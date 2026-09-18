-- Modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "internal_owner" character varying NULL, ADD COLUMN "internal_owner_user_id" character varying NULL, ADD COLUMN "internal_owner_group_id" character varying NULL, ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "stakeholder_name" character varying NULL, ADD COLUMN "stakeholder_user_id" character varying NULL, ADD COLUMN "stakeholder_group_id" character varying NULL, ADD COLUMN "stakeholder_identity_holder_id" character varying NULL, ADD COLUMN "delegate_name" character varying NULL, ADD COLUMN "delegate_user_id" character varying NULL, ADD COLUMN "delegate_group_id" character varying NULL, ADD COLUMN "delegate_identity_holder_id" character varying NULL;
-- Modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "internal_owner" character varying NULL, ADD COLUMN "internal_owner_user_id" character varying NULL, ADD COLUMN "internal_owner_group_id" character varying NULL, ADD COLUMN "internal_owner_identity_holder_id" character varying NULL;
