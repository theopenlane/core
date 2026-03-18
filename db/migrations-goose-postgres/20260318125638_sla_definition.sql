-- +goose Up
-- modify "findings" table
ALTER TABLE "findings" ADD COLUMN "finding_status_name" character varying NULL, ADD COLUMN "security_level" character varying NULL DEFAULT 'NONE', ADD COLUMN "finding_status_id" character varying NULL, ADD CONSTRAINT "findings_custom_type_enums_finding_status" FOREIGN KEY ("finding_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "sla_definitions" table
CREATE TABLE "sla_definitions" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "display_id" character varying NOT NULL, "tags" jsonb NULL, "sla_definition_severity_level_name" character varying NULL, "sla_days" bigint NOT NULL, "security_level" character varying NOT NULL DEFAULT 'NONE', "owner_id" character varying NULL, "sla_definition_severity_level_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "sla_definitions_custom_type_enums_sla_definition_severity_level" FOREIGN KEY ("sla_definition_severity_level_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "sla_definitions_organizations_sla_definitions" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "sladefinition_display_id_owner_id" to table: "sla_definitions"
CREATE UNIQUE INDEX "sladefinition_display_id_owner_id" ON "sla_definitions" ("display_id", "owner_id");
-- create index "sladefinition_owner_id" to table: "sla_definitions"
CREATE INDEX "sladefinition_owner_id" ON "sla_definitions" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "sladefinition_security_level_owner_id" to table: "sla_definitions"
CREATE UNIQUE INDEX "sladefinition_security_level_owner_id" ON "sla_definitions" ("security_level", "owner_id") WHERE (deleted_at IS NULL);
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "sla_definition_blocked_groups" character varying NULL, ADD COLUMN "sla_definition_editors" character varying NULL, ADD COLUMN "sla_definition_viewers" character varying NULL, ADD CONSTRAINT "groups_sla_definitions_blocked_groups" FOREIGN KEY ("sla_definition_blocked_groups") REFERENCES "sla_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_sla_definitions_editors" FOREIGN KEY ("sla_definition_editors") REFERENCES "sla_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_sla_definitions_viewers" FOREIGN KEY ("sla_definition_viewers") REFERENCES "sla_definitions" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "vulnerability_status_name" character varying NULL, ADD COLUMN "security_level" character varying NULL DEFAULT 'NONE', ADD COLUMN "vulnerability_status_id" character varying NULL, ADD CONSTRAINT "vulnerabilities_custom_type_enums_vulnerability_status" FOREIGN KEY ("vulnerability_status_id") REFERENCES "custom_type_enums" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" DROP CONSTRAINT "vulnerabilities_custom_type_enums_vulnerability_status", DROP COLUMN "vulnerability_status_id", DROP COLUMN "security_level", DROP COLUMN "vulnerability_status_name";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_sla_definitions_viewers", DROP CONSTRAINT "groups_sla_definitions_editors", DROP CONSTRAINT "groups_sla_definitions_blocked_groups", DROP COLUMN "sla_definition_viewers", DROP COLUMN "sla_definition_editors", DROP COLUMN "sla_definition_blocked_groups";
-- reverse: create index "sladefinition_security_level_owner_id" to table: "sla_definitions"
DROP INDEX "sladefinition_security_level_owner_id";
-- reverse: create index "sladefinition_owner_id" to table: "sla_definitions"
DROP INDEX "sladefinition_owner_id";
-- reverse: create index "sladefinition_display_id_owner_id" to table: "sla_definitions"
DROP INDEX "sladefinition_display_id_owner_id";
-- reverse: create "sla_definitions" table
DROP TABLE "sla_definitions";
-- reverse: modify "findings" table
ALTER TABLE "findings" DROP CONSTRAINT "findings_custom_type_enums_finding_status", DROP COLUMN "finding_status_id", DROP COLUMN "security_level", DROP COLUMN "finding_status_name";
