-- +goose Up
-- modify "mapped_controls" table
ALTER TABLE "mapped_controls" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- create index "mappedcontrol_external_id_owner_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_external_id_owner_id" ON "mapped_controls" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- create index "mappedcontrol_integration_run_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_integration_run_id" ON "mapped_controls" ("integration_run_id");
-- create index "mappedcontrol_source_instance_id" to table: "mapped_controls"
CREATE INDEX "mappedcontrol_source_instance_id" ON "mapped_controls" ("source_instance_id");
-- modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "mapped_control_integration_runs" character varying NULL, ADD CONSTRAINT "integration_runs_mapped_controls_integration_runs" FOREIGN KEY ("mapped_control_integration_runs") REFERENCES "mapped_controls" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" DROP CONSTRAINT "integration_runs_mapped_controls_integration_runs", DROP COLUMN "mapped_control_integration_runs";
-- reverse: create index "mappedcontrol_source_instance_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_source_instance_id";
-- reverse: create index "mappedcontrol_integration_run_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_integration_run_id";
-- reverse: create index "mappedcontrol_external_id_owner_id" to table: "mapped_controls"
DROP INDEX "mappedcontrol_external_id_owner_id";
-- reverse: modify "mapped_controls" table
ALTER TABLE "mapped_controls" DROP COLUMN "external_id", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
