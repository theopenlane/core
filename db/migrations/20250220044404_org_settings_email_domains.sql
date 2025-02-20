-- Modify "document_data_history" table
ALTER TABLE "document_data_history" ALTER COLUMN "owner_id" SET NOT NULL;
-- Modify "organization_setting_history" table
ALTER TABLE "organization_setting_history" ADD COLUMN "allowed_email_domains" jsonb NULL;
-- Modify "organization_settings" table
ALTER TABLE "organization_settings" ADD COLUMN "allowed_email_domains" jsonb NULL;
-- Modify "document_data" table
ALTER TABLE "document_data" DROP CONSTRAINT "document_data_organizations_document_data", ALTER COLUMN "owner_id" SET NOT NULL, ADD CONSTRAINT "document_data_organizations_document_data" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
