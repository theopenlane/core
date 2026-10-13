-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "control_integration_run_id" to table: "controls"
CREATE INDEX "control_integration_run_id" ON "controls" ("integration_run_id");
-- Create index "control_source_instance_id" to table: "controls"
CREATE INDEX "control_source_instance_id" ON "controls" ("source_instance_id");
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "group_integration_run_id" to table: "groups"
CREATE INDEX "group_integration_run_id" ON "groups" ("integration_run_id");
-- Create index "group_source_instance_id" to table: "groups"
CREATE INDEX "group_source_instance_id" ON "groups" ("source_instance_id");
-- Modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "platform_integration_run_id" to table: "platforms"
CREATE INDEX "platform_integration_run_id" ON "platforms" ("integration_run_id");
-- Create index "platform_source_instance_id" to table: "platforms"
CREATE INDEX "platform_source_instance_id" ON "platforms" ("source_instance_id");
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "program_integration_run_id" to table: "programs"
CREATE INDEX "program_integration_run_id" ON "programs" ("integration_run_id");
-- Create index "program_source_instance_id" to table: "programs"
CREATE INDEX "program_source_instance_id" ON "programs" ("source_instance_id");
-- Modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "review_integration_run_id" to table: "reviews"
CREATE INDEX "review_integration_run_id" ON "reviews" ("integration_run_id");
-- Create index "review_source_instance_id" to table: "reviews"
CREATE INDEX "review_source_instance_id" ON "reviews" ("source_instance_id");
-- Modify "system_details" table
ALTER TABLE "system_details" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- Create index "systemdetail_integration_run_id" to table: "system_details"
CREATE INDEX "systemdetail_integration_run_id" ON "system_details" ("integration_run_id");
-- Create index "systemdetail_source_instance_id" to table: "system_details"
CREATE INDEX "systemdetail_source_instance_id" ON "system_details" ("source_instance_id");
-- Modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "control_integration_runs" character varying NULL, ADD COLUMN "group_integration_runs" character varying NULL, ADD COLUMN "platform_integration_runs" character varying NULL, ADD COLUMN "program_integration_runs" character varying NULL, ADD COLUMN "review_integration_runs" character varying NULL, ADD COLUMN "system_detail_integration_runs" character varying NULL, ADD CONSTRAINT "integration_runs_controls_integration_runs" FOREIGN KEY ("control_integration_runs") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_groups_integration_runs" FOREIGN KEY ("group_integration_runs") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_platforms_integration_runs" FOREIGN KEY ("platform_integration_runs") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_programs_integration_runs" FOREIGN KEY ("program_integration_runs") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_reviews_integration_runs" FOREIGN KEY ("review_integration_runs") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_system_details_integration_runs" FOREIGN KEY ("system_detail_integration_runs") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
