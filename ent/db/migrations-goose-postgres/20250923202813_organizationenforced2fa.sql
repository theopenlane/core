-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "allow_matching_domains_autojoin" boolean NULL DEFAULT false, ADD COLUMN "saml_signin_url" character varying NULL, ADD COLUMN "saml_issuer" character varying NULL, ADD COLUMN "saml_cert" text NULL, ADD COLUMN "multifactor_auth_enforced" boolean NULL DEFAULT false;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "allow_matching_domains_autojoin" boolean NULL DEFAULT false, ADD COLUMN "saml_signin_url" character varying NULL, ADD COLUMN "saml_issuer" character varying NULL, ADD COLUMN "saml_cert" text NULL, ADD COLUMN "multifactor_auth_enforced" boolean NULL DEFAULT false;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "multifactor_auth_enforced", DROP COLUMN "saml_cert", DROP COLUMN "saml_issuer", DROP COLUMN "saml_signin_url", DROP COLUMN "allow_matching_domains_autojoin";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "multifactor_auth_enforced", DROP COLUMN "saml_cert", DROP COLUMN "saml_issuer", DROP COLUMN "saml_signin_url", DROP COLUMN "allow_matching_domains_autojoin";
