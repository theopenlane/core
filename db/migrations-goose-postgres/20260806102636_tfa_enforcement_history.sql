-- +goose Up
-- modify "org_membership_history" table
ALTER TABLE "org_membership_history" ADD COLUMN "tfa_enforced" boolean NULL DEFAULT false, ADD COLUMN "tfa_enforced_reason" character varying NULL, ADD COLUMN "tfa_enforced_by" character varying NULL, ADD COLUMN "tfa_enforced_at" timestamptz NULL;

-- +goose Down
-- reverse: modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "tfa_enforced_at", DROP COLUMN "tfa_enforced_by", DROP COLUMN "tfa_enforced_reason", DROP COLUMN "tfa_enforced";
