-- +goose Up
-- modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL;
-- modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL, ADD CONSTRAINT "action_plans_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "action_plans" table
ALTER TABLE "action_plans" DROP CONSTRAINT "action_plans_files_file", DROP COLUMN "file_id", DROP COLUMN "url";
-- reverse: modify "action_plan_history" table
ALTER TABLE "action_plan_history" DROP COLUMN "file_id", DROP COLUMN "url";
