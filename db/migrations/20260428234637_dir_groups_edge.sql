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
ALTER TABLE "directory_groups" DROP CONSTRAINT "directory_groups_identity_holders_directory_groups", DROP CONSTRAINT "directory_groups_integrations_directory_groups", DROP CONSTRAINT "directory_groups_organizations_directory_groups", DROP CONSTRAINT "directory_groups_platforms_directory_groups", DROP COLUMN "identity_holder_directory_groups", ADD CONSTRAINT "directory_groups_integrations_directory_groups" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_organizations_directory_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_platforms_directory_groups" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "directorygroup_display_id_owner_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_display_id_owner_id" ON "directory_groups" ("display_id", "owner_id");
-- Create index "directorygroup_integration_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_integration_id_email" ON "directory_groups" ("integration_id", "email");
-- Create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_integration_id_external_id_directory_sync_run_id" ON "directory_groups" ("integration_id", "external_id", "directory_sync_run_id");
-- Create index "directorygroup_owner_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_owner_id_email" ON "directory_groups" ("owner_id", "email");
-- Create index "directorygroup_platform_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_email" ON "directory_groups" ("platform_id", "email");
-- Create index "directorygroup_platform_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_external_id" ON "directory_groups" ("platform_id", "external_id");
