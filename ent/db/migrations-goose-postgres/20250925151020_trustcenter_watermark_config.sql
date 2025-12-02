-- +goose Up
-- create "trust_center_watermark_config_history" table
CREATE TABLE "trust_center_watermark_config_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "owner_id" character varying NULL, "trust_center_id" character varying NULL, "logo_id" character varying NULL, "text" character varying NULL, "font_size" double precision NULL DEFAULT 48, "opacity" double precision NULL DEFAULT 0.3, "rotation" double precision NULL DEFAULT 45, "color" character varying NULL DEFAULT '#808080', "font" character varying NULL DEFAULT 'arial', PRIMARY KEY ("id"));
-- create index "trustcenterwatermarkconfighistory_history_time" to table: "trust_center_watermark_config_history"
CREATE INDEX "trustcenterwatermarkconfighistory_history_time" ON "trust_center_watermark_config_history" ("history_time");
-- create "trust_center_watermark_configs" table
CREATE TABLE "trust_center_watermark_configs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "trust_center_id" character varying NULL, "text" character varying NULL, "font_size" double precision NULL DEFAULT 48, "opacity" double precision NULL DEFAULT 0.3, "rotation" double precision NULL DEFAULT 45, "color" character varying NULL DEFAULT '#808080', "font" character varying NULL DEFAULT 'arial', "owner_id" character varying NULL, "logo_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "trust_center_watermark_configs_e2f038ca8412a7e2b03e1fad46be2f7f" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "trust_center_watermark_configs_files_file" FOREIGN KEY ("logo_id") REFERENCES "files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "text_or_logo_id_not_null" CHECK ((text IS NOT NULL) OR (logo_id IS NOT NULL)));
-- create index "trustcenterwatermarkconfig_owner_id" to table: "trust_center_watermark_configs"
CREATE INDEX "trustcenterwatermarkconfig_owner_id" ON "trust_center_watermark_configs" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "trustcenterwatermarkconfig_trust_center_id" to table: "trust_center_watermark_configs"
CREATE UNIQUE INDEX "trustcenterwatermarkconfig_trust_center_id" ON "trust_center_watermark_configs" ("trust_center_id") WHERE (deleted_at IS NULL);
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "trust_center_watermark_config" character varying NULL, ADD CONSTRAINT "trust_centers_trust_center_watermark_configs_watermark_config" FOREIGN KEY ("trust_center_watermark_config") REFERENCES "trust_center_watermark_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP CONSTRAINT "trust_centers_trust_center_watermark_configs_watermark_config", DROP COLUMN "trust_center_watermark_config";
-- reverse: create index "trustcenterwatermarkconfig_trust_center_id" to table: "trust_center_watermark_configs"
DROP INDEX "trustcenterwatermarkconfig_trust_center_id";
-- reverse: create index "trustcenterwatermarkconfig_owner_id" to table: "trust_center_watermark_configs"
DROP INDEX "trustcenterwatermarkconfig_owner_id";
-- reverse: create "trust_center_watermark_configs" table
DROP TABLE "trust_center_watermark_configs";
-- reverse: create index "trustcenterwatermarkconfighistory_history_time" to table: "trust_center_watermark_config_history"
DROP INDEX "trustcenterwatermarkconfighistory_history_time";
-- reverse: create "trust_center_watermark_config_history" table
DROP TABLE "trust_center_watermark_config_history";
