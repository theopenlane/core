-- +goose Up
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ADD COLUMN "nda_approver_group_id" character varying NULL, ADD CONSTRAINT "trust_center_settings_groups_nda_approver_group" FOREIGN KEY ("nda_approver_group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP CONSTRAINT "trust_center_settings_groups_nda_approver_group", DROP COLUMN "nda_approver_group_id";
