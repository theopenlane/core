-- +goose Up
-- modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "reviewed_by" character varying NULL, ADD COLUMN "reviewed_by_user_id" character varying NULL, ADD COLUMN "reviewed_by_group_id" character varying NULL, ADD COLUMN "assigned_to" character varying NULL, ADD COLUMN "assigned_to_user_id" character varying NULL, ADD COLUMN "assigned_to_group_id" character varying NULL;
-- modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "reviewed_by" character varying NULL, ADD COLUMN "reviewed_by_user_id" character varying NULL, ADD COLUMN "reviewed_by_group_id" character varying NULL, ADD COLUMN "assigned_to" character varying NULL, ADD COLUMN "assigned_to_user_id" character varying NULL, ADD COLUMN "assigned_to_group_id" character varying NULL;

-- +goose Down
-- reverse: modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "assigned_to_group_id", DROP COLUMN "assigned_to_user_id", DROP COLUMN "assigned_to", DROP COLUMN "reviewed_by_group_id", DROP COLUMN "reviewed_by_user_id", DROP COLUMN "reviewed_by";
-- reverse: modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "assigned_to_group_id", DROP COLUMN "assigned_to_user_id", DROP COLUMN "assigned_to", DROP COLUMN "reviewed_by_group_id", DROP COLUMN "reviewed_by_user_id", DROP COLUMN "reviewed_by";
