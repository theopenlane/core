-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- Modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
