-- +goose Up
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ALTER COLUMN "allow_matching_domains_autojoin" SET DEFAULT true;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ALTER COLUMN "allow_matching_domains_autojoin" SET DEFAULT true;

-- +goose Down
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" ALTER COLUMN "allow_matching_domains_autojoin" SET DEFAULT false;
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ALTER COLUMN "allow_matching_domains_autojoin" SET DEFAULT false;
