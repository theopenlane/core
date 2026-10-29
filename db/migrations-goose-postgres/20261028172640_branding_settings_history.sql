-- +goose Up
-- create "organization_branding_setting_history" table
CREATE TABLE "organization_branding_setting_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "updated_by_impersonator" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "organization_id" character varying NULL, "remove_assessment_branding" boolean NULL DEFAULT false, "primary_color" character varying NULL, "secondary_color" character varying NULL, "background_color" character varying NULL, "text_color" character varying NULL, "font_family" character varying NULL, PRIMARY KEY ("id"));
-- create index "organizationbrandingsettinghistory_history_time" to table: "organization_branding_setting_history"
CREATE INDEX "organizationbrandingsettinghistory_history_time" ON "organization_branding_setting_history" ("history_time");

-- +goose Down
-- reverse: create index "organizationbrandingsettinghistory_history_time" to table: "organization_branding_setting_history"
DROP INDEX "organizationbrandingsettinghistory_history_time";
-- reverse: create "organization_branding_setting_history" table
DROP TABLE "organization_branding_setting_history";
