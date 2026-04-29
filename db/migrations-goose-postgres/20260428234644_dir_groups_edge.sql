-- +goose Up
-- drop index "directorygroup_display_id_owner_id" from table: "directory_groups"
DROP INDEX "directorygroup_display_id_owner_id";
-- drop index "directorygroup_integration_id_email" from table: "directory_groups"
DROP INDEX "directorygroup_integration_id_email";
-- drop index "directorygroup_integration_id_external_id_directory_sync_run_id" from table: "directory_groups"
DROP INDEX "directorygroup_integration_id_external_id_directory_sync_run_id";
-- drop index "directorygroup_owner_id_email" from table: "directory_groups"
DROP INDEX "directorygroup_owner_id_email";
-- drop index "directorygroup_platform_id_email" from table: "directory_groups"
DROP INDEX "directorygroup_platform_id_email";
-- drop index "directorygroup_platform_id_external_id" from table: "directory_groups"
DROP INDEX "directorygroup_platform_id_external_id";
-- modify "directory_groups" table
ALTER TABLE "directory_groups" DROP CONSTRAINT "directory_groups_identity_holders_directory_groups", DROP CONSTRAINT "directory_groups_integrations_directory_groups", DROP CONSTRAINT "directory_groups_organizations_directory_groups", DROP CONSTRAINT "directory_groups_platforms_directory_groups", DROP COLUMN "identity_holder_directory_groups", ADD CONSTRAINT "directory_groups_integrations_directory_groups" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_organizations_directory_groups" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_platforms_directory_groups" FOREIGN KEY ("platform_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "directorygroup_display_id_owner_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_display_id_owner_id" ON "directory_groups" ("display_id", "owner_id");
-- create index "directorygroup_integration_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_integration_id_email" ON "directory_groups" ("integration_id", "email");
-- create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_integration_id_external_id_directory_sync_run_id" ON "directory_groups" ("integration_id", "external_id", "directory_sync_run_id");
-- create index "directorygroup_owner_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_owner_id_email" ON "directory_groups" ("owner_id", "email");
-- create index "directorygroup_platform_id_email" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_email" ON "directory_groups" ("platform_id", "email");
-- create index "directorygroup_platform_id_external_id" to table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_external_id" ON "directory_groups" ("platform_id", "external_id");

-- +goose Down
-- reverse: create index "directorygroup_platform_id_external_id" to table: "directory_groups"
DROP INDEX "directorygroup_platform_id_external_id";
-- reverse: create index "directorygroup_platform_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_platform_id_email";
-- reverse: create index "directorygroup_owner_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_owner_id_email";
-- reverse: create index "directorygroup_integration_id_external_id_directory_sync_run_id" to table: "directory_groups"
DROP INDEX "directorygroup_integration_id_external_id_directory_sync_run_id";
-- reverse: create index "directorygroup_integration_id_email" to table: "directory_groups"
DROP INDEX "directorygroup_integration_id_email";
-- reverse: create index "directorygroup_display_id_owner_id" to table: "directory_groups"
DROP INDEX "directorygroup_display_id_owner_id";
-- reverse: modify "directory_groups" table
ALTER TABLE "directory_groups" DROP CONSTRAINT "directory_groups_platforms_directory_groups", DROP CONSTRAINT "directory_groups_organizations_directory_groups", DROP CONSTRAINT "directory_groups_integrations_directory_groups", ADD COLUMN "identity_holder_directory_groups" character varying NULL, ADD CONSTRAINT "directory_groups_platforms_directory_groups" FOREIGN KEY ("owner_id") REFERENCES "platforms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_organizations_directory_groups" FOREIGN KEY ("integration_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "directory_groups_integrations_directory_groups" FOREIGN KEY ("identity_holder_directory_groups") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "directory_groups_identity_holders_directory_groups" FOREIGN KEY ("directory_sync_run_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: drop index "directorygroup_platform_id_external_id" from table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_external_id" ON "directory_groups" ("owner_id", "external_id");
-- reverse: drop index "directorygroup_platform_id_email" from table: "directory_groups"
CREATE INDEX "directorygroup_platform_id_email" ON "directory_groups" ("owner_id", "email");
-- reverse: drop index "directorygroup_owner_id_email" from table: "directory_groups"
CREATE INDEX "directorygroup_owner_id_email" ON "directory_groups" ("integration_id", "email");
-- reverse: drop index "directorygroup_integration_id_external_id_directory_sync_run_id" from table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_integration_id_external_id_directory_sync_run_id" ON "directory_groups" ("identity_holder_directory_groups", "external_id", "scope_id");
-- reverse: drop index "directorygroup_integration_id_email" from table: "directory_groups"
CREATE INDEX "directorygroup_integration_id_email" ON "directory_groups" ("identity_holder_directory_groups", "email");
-- reverse: drop index "directorygroup_display_id_owner_id" from table: "directory_groups"
CREATE UNIQUE INDEX "directorygroup_display_id_owner_id" ON "directory_groups" ("display_id", "integration_id");
