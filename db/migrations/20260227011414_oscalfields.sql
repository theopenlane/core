-- Modify "controls" table
ALTER TABLE "controls" ADD COLUMN "external_uuid" character varying NULL, ADD COLUMN "implementation_status" character varying NULL DEFAULT 'PLANNED', ADD COLUMN "implementation_description" text NULL;
-- Create index "controls_external_uuid_key" to table: "controls"
CREATE UNIQUE INDEX "controls_external_uuid_key" ON "controls" ("external_uuid");
-- Modify "evidences" table
ALTER TABLE "evidences" ADD COLUMN "external_uuid" character varying NULL;
-- Create index "evidences_external_uuid_key" to table: "evidences"
CREATE UNIQUE INDEX "evidences_external_uuid_key" ON "evidences" ("external_uuid");
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "oscal_role" character varying NULL, ADD COLUMN "oscal_party_uuid" character varying NULL, ADD COLUMN "oscal_contact_uuids" jsonb NULL;
-- Modify "internal_policies" table
ALTER TABLE "internal_policies" ADD COLUMN "external_uuid" character varying NULL;
-- Create index "internal_policies_external_uuid_key" to table: "internal_policies"
CREATE UNIQUE INDEX "internal_policies_external_uuid_key" ON "internal_policies" ("external_uuid");
-- Modify "risks" table
ALTER TABLE "risks" ADD COLUMN "external_uuid" character varying NULL;
-- Create index "risks_external_uuid_key" to table: "risks"
CREATE UNIQUE INDEX "risks_external_uuid_key" ON "risks" ("external_uuid");
-- Modify "subcontrols" table
ALTER TABLE "subcontrols" ADD COLUMN "external_uuid" character varying NULL, ADD COLUMN "implementation_status" character varying NULL DEFAULT 'PLANNED', ADD COLUMN "implementation_description" text NULL;
-- Create index "subcontrols_external_uuid_key" to table: "subcontrols"
CREATE UNIQUE INDEX "subcontrols_external_uuid_key" ON "subcontrols" ("external_uuid");
-- Modify "tasks" table
ALTER TABLE "tasks" ADD COLUMN "external_uuid" character varying NULL;
-- Create index "tasks_external_uuid_key" to table: "tasks"
CREATE UNIQUE INDEX "tasks_external_uuid_key" ON "tasks" ("external_uuid");
-- Modify "platforms" table
ALTER TABLE "platforms" ADD COLUMN "external_uuid" character varying NULL;
-- Create index "platforms_external_uuid_key" to table: "platforms"
CREATE UNIQUE INDEX "platforms_external_uuid_key" ON "platforms" ("external_uuid");
-- Modify "programs" table
ALTER TABLE "programs" ADD COLUMN "external_uuid" character varying NULL;
-- Create index "programs_external_uuid_key" to table: "programs"
CREATE UNIQUE INDEX "programs_external_uuid_key" ON "programs" ("external_uuid");
-- Create "system_details" table
CREATE TABLE "system_details" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "system_name" character varying NOT NULL, "version" character varying NULL, "description" text NULL, "authorization_boundary" text NULL, "sensitivity_level" character varying NULL DEFAULT 'UNKNOWN', "last_reviewed" timestamptz NULL, "revision_history" jsonb NULL, "oscal_metadata_json" jsonb NULL, "owner_id" character varying NULL, "platform_id" character varying NULL, "program_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "system_details_organizations_system_details" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "system_details_platforms_system_detail" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "system_details_programs_system_detail" FOREIGN KEY ("program_id") REFERENCES "programs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "system_details_platform_id_key" to table: "system_details"
CREATE UNIQUE INDEX "system_details_platform_id_key" ON "system_details" ("platform_id");
-- Create index "system_details_program_id_key" to table: "system_details"
CREATE UNIQUE INDEX "system_details_program_id_key" ON "system_details" ("program_id");
-- Create index "systemdetail_display_id_owner_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_display_id_owner_id" ON "system_details" ("display_id", "owner_id");
-- Create index "systemdetail_owner_id" to table: "system_details"
CREATE INDEX "systemdetail_owner_id" ON "system_details" ("owner_id") WHERE (deleted_at IS NULL);
-- Create index "systemdetail_platform_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_platform_id" ON "system_details" ("platform_id") WHERE ((deleted_at IS NULL) AND (platform_id IS NOT NULL));
-- Create index "systemdetail_program_id" to table: "system_details"
CREATE UNIQUE INDEX "systemdetail_program_id" ON "system_details" ("program_id") WHERE ((deleted_at IS NULL) AND (program_id IS NOT NULL));
