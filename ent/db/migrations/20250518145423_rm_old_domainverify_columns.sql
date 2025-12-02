-- Modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "txt_record_subdomain", DROP COLUMN "txt_record_value", DROP COLUMN "status";
-- Modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "txt_record_subdomain", DROP COLUMN "txt_record_value", DROP COLUMN "status";
