-- Modify "action_plan_history" table
ALTER TABLE "action_plan_history" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL;
-- Modify "action_plans" table
ALTER TABLE "action_plans" ADD COLUMN "url" character varying NULL, ADD COLUMN "file_id" character varying NULL, ADD CONSTRAINT "action_plans_files_file" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
