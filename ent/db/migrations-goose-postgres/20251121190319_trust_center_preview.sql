-- +goose Up
-- modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "preview_domain_id" character varying NULL;
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "environment" character varying NULL DEFAULT 'LIVE';
-- drop index "trust_center_settings_trust_center_id_key" from table: "trust_center_settings"
DROP INDEX "trust_center_settings_trust_center_id_key";
-- drop index "trustcentersetting_trust_center_id" from table: "trust_center_settings"
DROP INDEX "trustcentersetting_trust_center_id";
-- modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP CONSTRAINT "trust_center_settings_trust_centers_setting", ADD COLUMN "environment" character varying NULL DEFAULT 'LIVE';
-- create index "trustcentersetting_trust_center_id_environment" to table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_trust_center_id_environment" ON "trust_center_settings" ("trust_center_id", "environment") WHERE (deleted_at IS NULL);
-- modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "preview_domain_id" character varying NULL, ADD COLUMN "trust_center_setting" character varying NULL, ADD COLUMN "trust_center_preview_setting" character varying NULL, ADD CONSTRAINT "trust_centers_custom_domains_preview_domain" FOREIGN KEY ("preview_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_preview_setting" FOREIGN KEY ("trust_center_preview_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_setting" FOREIGN KEY ("trust_center_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "trust_centers" table
ALTER TABLE "trust_centers" DROP CONSTRAINT "trust_centers_trust_center_settings_setting", DROP CONSTRAINT "trust_centers_trust_center_settings_preview_setting", DROP CONSTRAINT "trust_centers_custom_domains_preview_domain", DROP COLUMN "trust_center_preview_setting", DROP COLUMN "trust_center_setting", DROP COLUMN "preview_domain_id";
-- reverse: create index "trustcentersetting_trust_center_id_environment" to table: "trust_center_settings"
DROP INDEX "trustcentersetting_trust_center_id_environment";
-- reverse: modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP COLUMN "environment", ADD CONSTRAINT "trust_center_settings_trust_centers_setting" FOREIGN KEY ("trust_center_id") REFERENCES "trust_centers" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: drop index "trustcentersetting_trust_center_id" from table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_trust_center_id" ON "trust_center_settings" ("trust_center_id") WHERE (deleted_at IS NULL);
-- reverse: drop index "trust_center_settings_trust_center_id_key" from table: "trust_center_settings"
CREATE UNIQUE INDEX "trust_center_settings_trust_center_id_key" ON "trust_center_settings" ("trust_center_id");
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" DROP COLUMN "environment";
-- reverse: modify "trust_center_history" table
ALTER TABLE "trust_center_history" DROP COLUMN "preview_domain_id";
