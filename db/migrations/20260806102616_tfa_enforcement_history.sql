-- Modify "org_membership_history" table
ALTER TABLE "org_membership_history" ADD COLUMN "tfa_enforced" boolean NULL DEFAULT false, ADD COLUMN "tfa_enforced_reason" character varying NULL, ADD COLUMN "tfa_enforced_by" character varying NULL, ADD COLUMN "tfa_enforced_at" timestamptz NULL;
