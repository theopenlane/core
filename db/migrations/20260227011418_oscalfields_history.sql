-- Modify "control_history" table
ALTER TABLE "control_history" ADD COLUMN "external_uuid" character varying NULL, ADD COLUMN "implementation_status" character varying NULL DEFAULT 'PLANNED', ADD COLUMN "implementation_description" text NULL;
-- Modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "external_uuid" character varying NULL;
-- Modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "oscal_role" character varying NULL, ADD COLUMN "oscal_party_uuid" character varying NULL, ADD COLUMN "oscal_contact_uuids" jsonb NULL;
-- Modify "internal_policy_history" table
ALTER TABLE "internal_policy_history" ADD COLUMN "external_uuid" character varying NULL;
-- Modify "platform_history" table
ALTER TABLE "platform_history" ADD COLUMN "external_uuid" character varying NULL;
-- Modify "program_history" table
ALTER TABLE "program_history" ADD COLUMN "external_uuid" character varying NULL;
-- Modify "risk_history" table
ALTER TABLE "risk_history" ADD COLUMN "external_uuid" character varying NULL;
-- Modify "subcontrol_history" table
ALTER TABLE "subcontrol_history" ADD COLUMN "external_uuid" character varying NULL, ADD COLUMN "implementation_status" character varying NULL DEFAULT 'PLANNED', ADD COLUMN "implementation_description" text NULL;
-- Modify "task_history" table
ALTER TABLE "task_history" ADD COLUMN "external_uuid" character varying NULL;
-- Create "system_detail_history" table
CREATE TABLE "system_detail_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "program_id" character varying NULL, "platform_id" character varying NULL, "system_name" character varying NOT NULL, "version" character varying NULL, "description" text NULL, "authorization_boundary" text NULL, "sensitivity_level" character varying NULL DEFAULT 'UNKNOWN', "last_reviewed" timestamptz NULL, "revision_history" jsonb NULL, "oscal_metadata_json" jsonb NULL, PRIMARY KEY ("id"));
-- Create index "systemdetailhistory_history_time" to table: "system_detail_history"
CREATE INDEX "systemdetailhistory_history_time" ON "system_detail_history" ("history_time");
