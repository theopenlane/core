-- +goose Up
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- create index "control_integration_run_id" to table: "controls"
CREATE INDEX "control_integration_run_id" ON "controls" ("integration_run_id");
-- create index "control_source_instance_id" to table: "controls"
CREATE INDEX "control_source_instance_id" ON "controls" ("source_instance_id");
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- create index "group_integration_run_id" to table: "groups"
CREATE INDEX "group_integration_run_id" ON "groups" ("integration_run_id");
-- create index "group_source_instance_id" to table: "groups"
CREATE INDEX "group_source_instance_id" ON "groups" ("source_instance_id");
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- create index "mappedcontrol_external_id_owner_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_external_id_owner_id" ON "mapped_controls" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "mappedcontrol_integration_run_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_integration_run_id" ON "mapped_controls" ("integration_run_id");
-- create index "mappedcontrol_source_instance_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_source_instance_id" ON "mapped_controls" ("source_instance_id");
-- modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- create index "platform_integration_run_id" to table: "platforms"
CREATE INDEX "platform_integration_run_id" ON "platforms" ("integration_run_id");
-- create index "platform_source_instance_id" to table: "platforms"
CREATE INDEX "platform_source_instance_id" ON "platforms" ("source_instance_id");
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- create index "program_integration_run_id" to table: "programs"
CREATE INDEX "program_integration_run_id" ON "programs" ("integration_run_id");
-- create index "program_source_instance_id" to table: "programs"
CREATE INDEX "program_source_instance_id" ON "programs" ("source_instance_id");
-- modify "reviews" table
ALTER TABLE "reviews" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- create index "review_integration_run_id" to table: "reviews"
CREATE INDEX "review_integration_run_id" ON "reviews" ("integration_run_id");
-- create index "review_source_instance_id" to table: "reviews"
CREATE INDEX "review_source_instance_id" ON "reviews" ("source_instance_id");
-- modify "system_details" table
ALTER TABLE "system_details" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL;
-- create index "systemdetail_integration_run_id" to table: "system_details"
CREATE INDEX "systemdetail_integration_run_id" ON "system_details" ("integration_run_id");
-- create index "systemdetail_source_instance_id" to table: "system_details"
CREATE INDEX "systemdetail_source_instance_id" ON "system_details" ("source_instance_id");
-- modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "control_integration_runs" character varying NULL, ADD COLUMN "group_integration_runs" character varying NULL, ADD COLUMN "mapped_control_integration_runs" character varying NULL, ADD COLUMN "platform_integration_runs" character varying NULL, ADD COLUMN "program_integration_runs" character varying NULL, ADD COLUMN "review_integration_runs" character varying NULL, ADD COLUMN "system_detail_integration_runs" character varying NULL, ADD CONSTRAINT "integration_runs_controls_integration_runs" FOREIGN KEY ("control_integration_runs") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_groups_integration_runs" FOREIGN KEY ("group_integration_runs") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_mapped_controls_integration_runs" FOREIGN KEY ("mapped_control_integration_runs") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_platforms_integration_runs" FOREIGN KEY ("platform_integration_runs") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_programs_integration_runs" FOREIGN KEY ("program_integration_runs") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_reviews_integration_runs" FOREIGN KEY ("review_integration_runs") REFERENCES "reviews" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "integration_runs_system_details_integration_runs" FOREIGN KEY ("system_detail_integration_runs") REFERENCES "system_details" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "scans" table
ALTER TABLE "scans" ADD COLUMN "document_kind_name" character varying NULL, ADD COLUMN "document_kind_id" character varying NULL, ADD CONSTRAINT "scans_custom_type_enums_document_kind" FOREIGN KEY ("document_kind_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "scans" table
ALTER TABLE "scans" DROP CONSTRAINT "scans_custom_type_enums_document_kind", DROP COLUMN "document_kind_id", DROP COLUMN "document_kind_name";
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" DROP CONSTRAINT "integration_runs_system_details_integration_runs", DROP CONSTRAINT "integration_runs_reviews_integration_runs", DROP CONSTRAINT "integration_runs_programs_integration_runs", DROP CONSTRAINT "integration_runs_platforms_integration_runs", DROP CONSTRAINT "integration_runs_mapped_controls_integration_runs", DROP CONSTRAINT "integration_runs_groups_integration_runs", DROP CONSTRAINT "integration_runs_controls_integration_runs", DROP COLUMN "system_detail_integration_runs", DROP COLUMN "review_integration_runs", DROP COLUMN "program_integration_runs", DROP COLUMN "platform_integration_runs", DROP COLUMN "mapped_control_integration_runs", DROP COLUMN "group_integration_runs", DROP COLUMN "control_integration_runs";
-- reverse: create index "systemdetail_source_instance_id" to table: "system_details"
DROP INDEX "systemdetail_source_instance_id";
-- reverse: create index "systemdetail_integration_run_id" to table: "system_details"
DROP INDEX "systemdetail_integration_run_id";
-- reverse: modify "system_details" table
ALTER TABLE "system_details" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: create index "review_source_instance_id" to table: "reviews"
DROP INDEX "review_source_instance_id";
-- reverse: create index "review_integration_run_id" to table: "reviews"
DROP INDEX "review_integration_run_id";
-- reverse: modify "reviews" table
ALTER TABLE "reviews" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: create index "program_source_instance_id" to table: "programs"
DROP INDEX "program_source_instance_id";
-- reverse: create index "program_integration_run_id" to table: "programs"
DROP INDEX "program_integration_run_id";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: create index "platform_source_instance_id" to table: "platforms"
DROP INDEX "platform_source_instance_id";
-- reverse: create index "platform_integration_run_id" to table: "platforms"
DROP INDEX "platform_integration_run_id";
-- reverse: modify "platforms" table
ALTER TABLE "platforms" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: create index "mappedcontrol_source_instance_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_source_instance_id";
-- reverse: create index "mappedcontrol_integration_run_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_integration_run_id";
-- reverse: create index "mappedcontrol_external_id_owner_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_external_id_owner_id";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: create index "group_source_instance_id" to table: "groups"
DROP INDEX "group_source_instance_id";
-- reverse: create index "group_integration_run_id" to table: "groups"
DROP INDEX "group_integration_run_id";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
-- reverse: create index "control_source_instance_id" to table: "controls"
DROP INDEX "control_source_instance_id";
-- reverse: create index "control_integration_run_id" to table: "controls"
DROP INDEX "control_integration_run_id";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
