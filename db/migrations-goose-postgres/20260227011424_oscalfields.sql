-- +goose Up
-- modify "controls" table
ALTER TABLE "controls" ADD COLUMN "external_uuid" character varying NULL, ADD COLUMN "implementation_status" character varying NULL DEFAULT 'PLANNED', ADD COLUMN "implementation_description" text NULL;
-- create index "controls_external_uuid_key" to table: "controls"
CREATE UNIQUE INDEX "controls_external_uuid_key" ON "controls" ("external_uuid");
-- modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "external_uuid" character varying NULL;
-- create index "evidences_external_uuid_key" to table: "evidences"
CREATE UNIQUE INDEX "evidences_external_uuid_key" ON "evidences" ("external_uuid");
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "oscal_role" character varying NULL, ADD COLUMN "oscal_party_uuid" character varying NULL, ADD COLUMN "oscal_contact_uuids" jsonb NULL;
-- modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "external_uuid" character varying NULL;
-- create index "internal_policies_external_uuid_key" to table: "internal_policies"
CREATE UNIQUE INDEX "internal_policies_external_uuid_key" ON "internal_policies" ("external_uuid");
-- modify "risks" table
ALTER TABLE "risks" ADD COLUMN "external_uuid" character varying NULL;
-- create index "risks_external_uuid_key" to table: "risks"
CREATE UNIQUE INDEX "risks_external_uuid_key" ON "risks" ("external_uuid");
-- modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "external_uuid" character varying NULL, ADD COLUMN "implementation_status" character varying NULL DEFAULT 'PLANNED', ADD COLUMN "implementation_description" text NULL;
-- create index "subcontrols_external_uuid_key" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrols_external_uuid_key" ON "subcontrols" ("external_uuid");
-- modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "external_uuid" character varying NULL;
-- create index "tasks_external_uuid_key" to table: "tasks"
CREATE UNIQUE INDEX "tasks_external_uuid_key" ON "tasks" ("external_uuid");
-- modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "external_uuid" character varying NULL;
-- create index "platforms_external_uuid_key" to table: "platforms"
CREATE UNIQUE INDEX "platforms_external_uuid_key" ON "platforms" ("external_uuid");
-- modify "programs" table
ALTER TABLE "programs" ADD COLUMN "external_uuid" character varying NULL;
-- create index "programs_external_uuid_key" to table: "programs"
CREATE UNIQUE INDEX "programs_external_uuid_key" ON "programs" ("external_uuid");
-- create "system_details" table
CREATE TABLE "system_details" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_name" character varying NOT NULL, "version" character varying NULL, "description" text NULL, "authorization_boundary" text NULL, "sensitivity_level" character varying NULL DEFAULT 'UNKNOWN', "last_reviewed" timestamptz NULL, "revision_history" jsonb NULL, "oscal_metadata_json" jsonb NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, "program_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "system_details_organizations_system_details" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "system_details_platforms_system_detail" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "system_details_programs_system_detail" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "system_details_platform_id_key" to table: "system_details"
CREATE UNIQUE INDEX "system_details_platform_id_key" ON "system_details" ("platform_id");
-- create index "system_details_program_id_key" to table: "system_details"
CREATE UNIQUE INDEX "system_details_program_id_key" ON "system_details" ("program_id");
-- create index "systemdetail_display_id_owner_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_display_id_owner_id" ON "system_details" ("display_id", "owner_id");
-- create index "systemdetail_owner_id" to table: "system_details"
CREATE INDEX "systemdetail_owner_id" ON "system_details" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "systemdetail_platform_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_platform_id" ON "system_details" ("platform_id") WHERE ((deleted_at IS NULL) AND (platform_id IS NOT NULL));
-- create index "systemdetail_program_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_program_id" ON "system_details" ("program_id") WHERE ((deleted_at IS NULL) AND (program_id IS NOT NULL));

-- +goose Down
-- reverse: create index "systemdetail_program_id" to table: "system_details"
DROP INDEX "systemdetail_program_id";
-- reverse: create index "systemdetail_platform_id" to table: "system_details"
DROP INDEX "systemdetail_platform_id";
-- reverse: create index "systemdetail_owner_id" to table: "system_details"
DROP INDEX "systemdetail_owner_id";
-- reverse: create index "systemdetail_display_id_owner_id" to table: "system_details"
DROP INDEX "systemdetail_display_id_owner_id";
-- reverse: create index "system_details_program_id_key" to table: "system_details"
DROP INDEX "system_details_program_id_key";
-- reverse: create index "system_details_platform_id_key" to table: "system_details"
DROP INDEX "system_details_platform_id_key";
-- reverse: create "system_details" table
DROP TABLE "system_details";
-- reverse: create index "programs_external_uuid_key" to table: "programs"
DROP INDEX "programs_external_uuid_key";
-- reverse: modify "programs" table
ALTER TABLE "programs" DROP COLUMN "external_uuid";
-- reverse: create index "platforms_external_uuid_key" to table: "platforms"
DROP INDEX "platforms_external_uuid_key";
-- reverse: modify "platforms" table
ALTER TABLE "platforms" DROP COLUMN "external_uuid";
-- reverse: create index "tasks_external_uuid_key" to table: "tasks"
DROP INDEX "tasks_external_uuid_key";
-- reverse: modify "tasks" table
ALTER TABLE "tasks" DROP COLUMN "external_uuid";
-- reverse: create index "subcontrols_external_uuid_key" to table: "subcontrols"
DROP INDEX "subcontrols_external_uuid_key";
-- reverse: modify "subcontrols" table
ALTER TABLE "subcontrols" DROP COLUMN "implementation_description", DROP COLUMN "implementation_status", DROP COLUMN "external_uuid";
-- reverse: create index "risks_external_uuid_key" to table: "risks"
DROP INDEX "risks_external_uuid_key";
-- reverse: modify "risks" table
ALTER TABLE "risks" DROP COLUMN "external_uuid";
-- reverse: create index "internal_policies_external_uuid_key" to table: "internal_policies"
DROP INDEX "internal_policies_external_uuid_key";
-- reverse: modify "internal_policies" table
ALTER TABLE "internal_policies" DROP COLUMN "external_uuid";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "oscal_contact_uuids", DROP COLUMN "oscal_party_uuid", DROP COLUMN "oscal_role";
-- reverse: create index "evidences_external_uuid_key" to table: "evidences"
DROP INDEX "evidences_external_uuid_key";
-- reverse: modify "evidences" table
ALTER TABLE "evidences" DROP COLUMN "external_uuid";
-- reverse: create index "controls_external_uuid_key" to table: "controls"
DROP INDEX "controls_external_uuid_key";
-- reverse: modify "controls" table
ALTER TABLE "controls" DROP COLUMN "implementation_description", DROP COLUMN "implementation_status", DROP COLUMN "external_uuid";
