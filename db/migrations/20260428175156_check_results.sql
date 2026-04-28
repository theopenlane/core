-- Modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "directory_name" character varying NULL;
-- Modify "vulnerabilities" table
ALTER TABLE "vulnerabilities" ADD COLUMN "fix_available" boolean NULL;
-- Create "check_results" table
CREATE TABLE "check_results" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "status" character varying NOT NULL DEFAULT 'UNKNOWN', "source" character varying NOT NULL, "last_observed_at" timestamptz NULL, "external_uri" character varying NULL, "details" text NULL, "parent_external_id" character varying NULL, "integration_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "check_results_integrations_check_results" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create "check_result_controls" table
CREATE TABLE "check_result_controls" ("check_result_id" character varying NOT NULL, "control_id" character varying NOT NULL, PRIMARY KEY ("check_result_id", "control_id"), CONSTRAINT "check_result_controls_check_result_id" FOREIGN KEY ("check_result_id") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "check_result_controls_control_id" FOREIGN KEY ("control_id") REFERENCES "controls" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Drop index "directorygroup_display_id_owner_id" from table: "directory_groups"
DROP INDEX "directorygroup_display_id_owner_id";
-- Drop index "directorygroup_integration_id_email" from table: "directory_groups"
DROP INDEX "directorygroup_integration_id_email";
-- Drop index "directorygroup_integration_id_external_id_directory_sync_run_id" from table: "directory_groups"
DROP INDEX "directorygroup_integration_id_external_id_directory_sync_run_id";
-- Drop index "directorygroup_owner_id_email" from table: "directory_groups"
DROP INDEX "directorygroup_owner_id_email";
-- Drop index "directorygroup_platform_id_email" from table: "directory_groups"
DROP INDEX "directorygroup_platform_id_email";
-- Drop index "directorygroup_platform_id_external_id" from table: "directory_groups"
DROP INDEX "directorygroup_platform_id_external_id";
-- Modify "directory_groups" table
ALTER TABLE "directory_groups" DROP CONSTRAINT "directory_groups_identity_holders_directory_groups", DROP CONSTRAINT "directory_groups_integrations_directory_groups", DROP CONSTRAINT "directory_groups_organizations_directory_groups", DROP CONSTRAINT "directory_groups_platforms_directory_groups", ADD COLUMN "directory_name" character varying NULL, ADD CONSTRAINT "directory_groups_identity_holders_directory_groups" FOREIGN KEY ("directory_sync_run_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_integrations_directory_groups" FOREIGN KEY ("identity_holder_directory_groups") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_organizations_directory_groups" FOREIGN KEY ("integration_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_platforms_directory_groups" FOREIGN KEY ("owner_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "directorygroup_display_id_owner_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_display_id_owner_id" ON "directory_groups" ("display_id", "integration_id");
-- Create index "directorygroup_integration_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_integration_id_email" ON "directory_groups" ("identity_holder_directory_groups", "email");
-- Create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_integration_id_external_id_directory_sync_run_id" ON "directory_groups" ("identity_holder_directory_groups", "external_id", "scope_id");
-- Create index "directorygroup_owner_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_owner_id_email" ON "directory_groups" ("integration_id", "email");
-- Create index "directorygroup_platform_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_email" ON "directory_groups" ("owner_id", "email");
-- Create index "directorygroup_platform_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_external_id" ON "directory_groups" ("owner_id", "external_id");
-- Create "finding_check_results" table
CREATE TABLE "finding_check_results" ("finding_id" character varying NOT NULL, "check_result_id" character varying NOT NULL, PRIMARY KEY ("finding_id", "check_result_id"), CONSTRAINT "finding_check_results_check_result_id" FOREIGN KEY ("check_result_id") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "finding_check_results_finding_id" FOREIGN KEY ("finding_id") REFERENCES "findings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "check_result_blocked_groups" character varying NULL, ADD COLUMN "check_result_editors" character varying NULL, ADD COLUMN "check_result_viewers" character varying NULL, ADD CONSTRAINT "groups_check_results_blocked_groups" FOREIGN KEY ("check_result_blocked_groups") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_editors" FOREIGN KEY ("check_result_editors") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_check_results_viewers" FOREIGN KEY ("check_result_viewers") REFERENCES "check_results" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
