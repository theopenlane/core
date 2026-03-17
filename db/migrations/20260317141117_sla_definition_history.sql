-- Modify "finding_history" table
ALTER TABLE "finding_history" DROP COLUMN "status", ADD COLUMN "finding_status_name" character varying NULL, ADD COLUMN "finding_status_id" character varying NULL, ADD COLUMN "security_level" character varying NULL;
-- Modify "vulnerability_history" table
ALTER TABLE "vulnerability_history" DROP COLUMN "status", ADD COLUMN "vulnerability_status_name" character varying NULL, ADD COLUMN "vulnerability_status_id" character varying NULL, ADD COLUMN "security_level" character varying NULL;
-- Create "sla_definition_history" table
CREATE TABLE "sla_definition_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "sla_definition_severity_level_name" character varying NULL, "sla_definition_severity_level_id" character varying NULL, "sla_days" bigint NOT NULL, "security_level" character varying NOT NULL DEFAULT 'NONE', PRIMARY KEY ("id"));
-- Create index "sladefinitionhistory_history_time" to table: "sla_definition_history"
CREATE INDEX "sladefinitionhistory_history_time" ON "sla_definition_history" ("history_time");
