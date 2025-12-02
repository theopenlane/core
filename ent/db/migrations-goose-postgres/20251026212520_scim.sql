-- +goose Up
-- modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_display_name" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_group_mailing" character varying NULL;
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_display_name" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_group_mailing" character varying NULL;
-- modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ALTER COLUMN "font" SET DEFAULT 'HELVETICA';
-- modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ALTER COLUMN "font" SET DEFAULT 'HELVETICA';
-- modify "user_history" table
ALTER TABLE "user_history" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_username" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_preferred_language" character varying NULL, ADD COLUMN "scim_locale" character varying NULL;
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "scim_external_id" character varying NULL, ADD COLUMN "scim_username" character varying NULL, ADD COLUMN "scim_active" boolean NULL DEFAULT true, ADD COLUMN "scim_preferred_language" character varying NULL, ADD COLUMN "scim_locale" character varying NULL;

-- +goose Down
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "scim_locale", DROP COLUMN "scim_preferred_language", DROP COLUMN "scim_active", DROP COLUMN "scim_username", DROP COLUMN "scim_external_id";
-- reverse: modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "scim_locale", DROP COLUMN "scim_preferred_language", DROP COLUMN "scim_active", DROP COLUMN "scim_username", DROP COLUMN "scim_external_id";
-- reverse: modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ALTER COLUMN "font" SET DEFAULT 'arial';
-- reverse: modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ALTER COLUMN "font" SET DEFAULT 'arial';
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "scim_group_mailing", DROP COLUMN "scim_active", DROP COLUMN "scim_display_name", DROP COLUMN "scim_external_id";
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "scim_group_mailing", DROP COLUMN "scim_active", DROP COLUMN "scim_display_name", DROP COLUMN "scim_external_id";
