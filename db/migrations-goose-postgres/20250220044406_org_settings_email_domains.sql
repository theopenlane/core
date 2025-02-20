-- +goose Up
-- modify "document_data_history" table
ALTER TABLE "document_data_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "allowed_email_domains" jsonb NULL;
-- modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "allowed_email_domains" jsonb NULL;
-- modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_organizations_document_data", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "document_data_organizations_document_data" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- reverse: modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_organizations_document_data", ALTER COLUMN "owner_id" DROP NOT NULL, ADD CONSTRAINT "document_data_organizations_document_data" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- reverse: modify "organization_settings" table
ALTER TABLE "organization_settings" DROP COLUMN "allowed_email_domains";
-- reverse: modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" DROP COLUMN "allowed_email_domains";
-- reverse: modify "document_data_history" table
ALTER TABLE "document_data_history" ALTER COLUMN "owner_id" DROP NOT NULL;
