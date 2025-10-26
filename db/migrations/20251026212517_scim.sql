-- Modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_display_name" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_group_mailing" character varying NULL;
-- Modify "groups" table
ALTER TABLE "groups" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_display_name" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_group_mailing" character varying NULL;
-- Modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ALTER COLUMN "font" SET DEFAULT 'HELVETICA';
-- Modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ALTER COLUMN "font" SET DEFAULT 'HELVETICA';
-- Modify "user_history" table
ALTER TABLE "user_history" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_username" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_preferred_language" character varying NULL, ADD COLUMN "scim_locale" character varying NULL;
-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_username" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_preferred_language" character varying NULL, ADD COLUMN "scim_locale" character varying NULL;
