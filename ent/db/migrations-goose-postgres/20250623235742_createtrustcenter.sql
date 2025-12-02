-- +goose Up
-- create "trust_center_history" table
CREATE TABLE "trust_center_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "slug" character varying NULL, "custom_domain_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "trustcenterhistory_history_time" to table: "trust_center_history"
CREATE INDEX "trustcenterhistory_history_time" ON "trust_center_history" ("history_time");
-- create "trust_center_setting_history" table
CREATE TABLE "trust_center_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_id" character varying NULL, "title" character varying NULL, "overview" character varying NULL, "primary_color" character varying NULL, PRIMARY KEY ("id"));
-- create index "trustcentersettinghistory_history_time" to table: "trust_center_setting_history"
CREATE INDEX "trustcentersettinghistory_history_time" ON "trust_center_setting_history" ("history_time");
-- create "trust_centers" table
CREATE TABLE "trust_centers" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "slug" character varying NULL, "owner_id" character varying NULL, "custom_domain_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "trust_centers_custom_domains_custom_domain" FOREIGN KEY ("custom_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "trust_centers_organizations_trust_centers" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "trustcenter_id" to table: "trust_centers"
CREATE UNIQUE INDEX "trustcenter_id" ON "trust_centers" ("id");
-- create index "trustcenter_owner_id" to table: "trust_centers"
CREATE INDEX "trustcenter_owner_id" ON "trust_centers" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "trustcenter_slug" to table: "trust_centers"
CREATE UNIQUE INDEX "trustcenter_slug" ON "trust_centers" ("slug") WHERE (deleted_at IS NULL);
-- create "trust_center_settings" table
CREATE TABLE "trust_center_settings" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "title" character varying NULL, "overview" character varying NULL, "primary_color" character varying NULL, "trust_center_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "trust_center_settings_trust_centers_setting" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "trust_center_settings_trust_center_id_key" to table: "trust_center_settings"
CREATE UNIQUE INDEX "trust_center_settings_trust_center_id_key" ON "trust_center_settings" ("trust_center_id");
-- create index "trustcentersetting_id" to table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_id" ON "trust_center_settings" ("id");
-- create index "trustcentersetting_trust_center_id" to table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_trust_center_id" ON "trust_center_settings" ("trust_center_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "trustcentersetting_trust_center_id" to table: "trust_center_settings"
DROP INDEX "trustcentersetting_trust_center_id";
-- reverse: create index "trustcentersetting_id" to table: "trust_center_settings"
DROP INDEX "trustcentersetting_id";
-- reverse: create index "trust_center_settings_trust_center_id_key" to table: "trust_center_settings"
DROP INDEX "trust_center_settings_trust_center_id_key";
-- reverse: create "trust_center_settings" table
DROP TABLE "trust_center_settings";
-- reverse: create index "trustcenter_slug" to table: "trust_centers"
DROP INDEX "trustcenter_slug";
-- reverse: create index "trustcenter_owner_id" to table: "trust_centers"
DROP INDEX "trustcenter_owner_id";
-- reverse: create index "trustcenter_id" to table: "trust_centers"
DROP INDEX "trustcenter_id";
-- reverse: create "trust_centers" table
DROP TABLE "trust_centers";
-- reverse: create index "trustcentersettinghistory_history_time" to table: "trust_center_setting_history"
DROP INDEX "trustcentersettinghistory_history_time";
-- reverse: create "trust_center_setting_history" table
DROP TABLE "trust_center_setting_history";
-- reverse: create index "trustcenterhistory_history_time" to table: "trust_center_history"
DROP INDEX "trustcenterhistory_history_time";
-- reverse: create "trust_center_history" table
DROP TABLE "trust_center_history";
