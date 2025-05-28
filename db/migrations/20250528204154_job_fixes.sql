-- Modify "group_membership_history" table
ALTER TABLE "group_membership_history" DROP COLUMN "deleted_at", DROP COLUMN "deleted_by";
-- Drop index "groupmembership_user_id_group_id" from table: "group_memberships"
DROP INDEX "groupmembership_user_id_group_id";
-- Modify "group_memberships" table
ALTER TABLE "group_memberships" DROP COLUMN "deleted_at", DROP COLUMN "deleted_by";
-- Create index "groupmembership_user_id_group_id" to table: "group_memberships"
CREATE UNIQUE INDEX "groupmembership_user_id_group_id" ON "group_memberships" ("user_id", "group_id");
-- Modify "org_membership_history" table
ALTER TABLE "org_membership_history" DROP COLUMN "deleted_at", DROP COLUMN "deleted_by";
-- Drop index "orgmembership_user_id_organization_id" from table: "org_memberships"
DROP INDEX "orgmembership_user_id_organization_id";
-- Modify "org_memberships" table
ALTER TABLE "org_memberships" DROP COLUMN "deleted_at", DROP COLUMN "deleted_by";
-- Create index "orgmembership_user_id_organization_id" to table: "org_memberships"
CREATE UNIQUE INDEX "orgmembership_user_id_organization_id" ON "org_memberships" ("user_id", "organization_id");
-- Modify "program_membership_history" table
ALTER TABLE "program_membership_history" DROP COLUMN "deleted_at", DROP COLUMN "deleted_by";
-- Drop index "programmembership_user_id_program_id" from table: "program_memberships"
DROP INDEX "programmembership_user_id_program_id";
-- Modify "program_memberships" table
ALTER TABLE "program_memberships" DROP COLUMN "deleted_at", DROP COLUMN "deleted_by";
-- Create index "programmembership_user_id_program_id" to table: "program_memberships"
CREATE UNIQUE INDEX "programmembership_user_id_program_id" ON "program_memberships" ("user_id", "program_id");
-- Modify "scheduled_job_runs" table
ALTER TABLE "scheduled_job_runs" ALTER COLUMN "job_runner_id" SET NOT NULL, ADD COLUMN "expected_execution_time" timestamptz NOT NULL, ADD COLUMN "script" character varying NOT NULL, ADD CONSTRAINT "scheduled_job_runs_job_runners_job_runner" FOREIGN KEY ("job_runner_id") REFERENCES "job_runners" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
