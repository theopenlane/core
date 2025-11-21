-- Modify "trust_center_history" table
ALTER TABLE "trust_center_history" ADD COLUMN "preview_domain_id" character varying NULL;
-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ADD COLUMN "environment" character varying NULL DEFAULT 'LIVE';
-- Drop index "trust_center_settings_trust_center_id_key" from table: "trust_center_settings"
DROP INDEX "trust_center_settings_trust_center_id_key";
-- Drop index "trustcentersetting_trust_center_id" from table: "trust_center_settings"
DROP INDEX "trustcentersetting_trust_center_id";
-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" DROP CONSTRAINT "trust_center_settings_trust_centers_setting", ADD COLUMN "environment" character varying NULL DEFAULT 'LIVE';
-- Create index "trustcentersetting_trust_center_id_environment" to table: "trust_center_settings"
CREATE UNIQUE INDEX "trustcentersetting_trust_center_id_environment" ON "trust_center_settings" ("trust_center_id", "environment") WHERE (deleted_at IS NULL);
-- Modify "trust_centers" table
ALTER TABLE "trust_centers" ADD COLUMN "preview_domain_id" character varying NULL, ADD COLUMN "trust_center_setting" character varying NULL, ADD COLUMN "trust_center_preview_setting" character varying NULL, ADD CONSTRAINT "trust_centers_custom_domains_preview_domain" FOREIGN KEY ("preview_domain_id") REFERENCES "custom_domains" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_preview_setting" FOREIGN KEY ("trust_center_preview_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "trust_centers_trust_center_settings_setting" FOREIGN KEY ("trust_center_setting") REFERENCES "trust_center_settings" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
