-- +goose Up
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "external_sharing_enabled" boolean NULL DEFAULT false;
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "organization_share_link_creators" character varying NULL, ADD COLUMN "organization_share_link_manager" character varying NULL, ADD CONSTRAINT "groups_organizations_share_link_creators" FOREIGN KEY ("organization_share_link_creators") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "groups_organizations_share_link_manager" FOREIGN KEY ("organization_share_link_manager") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create "share_links" table
CREATE TABLE "share_links" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "name" character varying NULL, "expires_at" timestamptz NULL, "max_uses" bigint NULL DEFAULT 0, "use_count" bigint NOT NULL DEFAULT 0, "recipient_emails" jsonb NULL, "landing_page" boolean NULL DEFAULT false, "require_session" boolean NULL DEFAULT false, "code_hash" bytea NULL, "code_expires_at" timestamptz NULL, "code_attempts" bigint NOT NULL DEFAULT 0, "revoked_at" timestamptz NULL, "token" character varying NOT NULL, "secret" bytea NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "share_links_organizations_share_links" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "share_link_owner_id_idx" to table: "share_links"
CREATE INDEX "share_link_owner_id_idx" ON "share_links" ("owner_id");
-- create "share_link_access_events" table
CREATE TABLE "share_link_access_events" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "file_id" character varying NOT NULL, "ip_address" character varying NULL, "user_agent" character varying NULL, "owner_id" character varying NULL, "share_link_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "share_link_access_events_organizations_share_link_access_events" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "share_link_access_events_share_links_share_link_access_events" FOREIGN KEY ("share_link_id") REFERENCES "share_links" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- create index "share_link_access_event_owner_id_idx" to table: "share_link_access_events"
CREATE INDEX "share_link_access_event_owner_id_idx" ON "share_link_access_events" ("owner_id");
-- create index "share_link_access_event_share_link_id_idx" to table: "share_link_access_events"
CREATE INDEX "share_link_access_event_share_link_id_idx" ON "share_link_access_events" ("share_link_id");
-- create "share_link_files" table
CREATE TABLE "share_link_files" ("share_link_id" character varying NOT NULL, "file_id" character varying NOT NULL, PRIMARY KEY ("share_link_id", "file_id"), CONSTRAINT "share_link_files_file_id" FOREIGN KEY ("file_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "share_link_files_share_link_id" FOREIGN KEY ("share_link_id") REFERENCES "share_links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create index "share_link_files_file_id_idx" to table: "share_link_files"
CREATE INDEX "share_link_files_file_id_idx" ON "share_link_files" ("file_id");

-- +goose Down
-- reverse: create index "share_link_files_file_id_idx" to table: "share_link_files"
DROP INDEX "share_link_files_file_id_idx";
-- reverse: create "share_link_files" table
DROP TABLE "share_link_files";
-- reverse: create index "share_link_access_event_share_link_id_idx" to table: "share_link_access_events"
DROP INDEX "share_link_access_event_share_link_id_idx";
-- reverse: create index "share_link_access_event_owner_id_idx" to table: "share_link_access_events"
DROP INDEX "share_link_access_event_owner_id_idx";
-- reverse: create "share_link_access_events" table
DROP TABLE "share_link_access_events";
-- reverse: create index "share_link_owner_id_idx" to table: "share_links"
DROP INDEX "share_link_owner_id_idx";
-- reverse: create "share_links" table
DROP TABLE "share_links";
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP CONSTRAINT "groups_organizations_share_link_manager", DROP CONSTRAINT "groups_organizations_share_link_creators", DROP COLUMN "organization_share_link_manager", DROP COLUMN "organization_share_link_creators";
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "external_sharing_enabled";
