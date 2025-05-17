-- Modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ALTER COLUMN "txt_record_subdomain" DROP NOT NULL, ALTER COLUMN "txt_record_value" DROP NOT NULL, ALTER COLUMN "status" DROP NOT NULL;
-- Modify "custom_domains" table
ALTER TABLE "custom_domains" ALTER COLUMN "txt_record_subdomain" DROP NOT NULL, ALTER COLUMN "txt_record_value" DROP NOT NULL, ALTER COLUMN "status" DROP NOT NULL;
