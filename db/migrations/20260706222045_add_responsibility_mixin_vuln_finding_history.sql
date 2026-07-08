-- Modify "finding_history" table
ALTER TABLE "finding_history" ADD COLUMN "reviewed_by" character varying NULL, ADD COLUMN "reviewed_by_user_id" character varying NULL, ADD COLUMN "reviewed_by_group_id" character varying NULL, ADD COLUMN "assigned_to" character varying NULL, ADD COLUMN "assigned_to_user_id" character varying NULL, ADD COLUMN "assigned_to_group_id" character varying NULL;
-- Modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" ADD COLUMN "reviewed_by" character varying NULL, ADD COLUMN "reviewed_by_user_id" character varying NULL, ADD COLUMN "reviewed_by_group_id" character varying NULL, ADD COLUMN "assigned_to" character varying NULL, ADD COLUMN "assigned_to_user_id" character varying NULL, ADD COLUMN "assigned_to_group_id" character varying NULL;
