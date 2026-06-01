-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;
-- modify "procedure_history" table
ALTER TABLE "procedure_history" ADD COLUMN "external_file_id" character varying NULL, ADD COLUMN "external_contents" character varying NULL;

-- +goose Down
-- reverse: modify "procedure_history" table
ALTER TABLE "procedure_history" DROP COLUMN "external_contents", DROP COLUMN "external_file_id";
-- reverse: modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" DROP COLUMN "external_contents", DROP COLUMN "external_file_id";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "external_contents", DROP COLUMN "external_file_id";
