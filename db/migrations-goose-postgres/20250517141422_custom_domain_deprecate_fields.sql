-- +goose Up
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ALTER COLUMN "txt_record_subdomain" DROP NOT NULL, ALTER COLUMN "txt_record_value" DROP NOT NULL, ALTER COLUMN "status" DROP NOT NULL;
-- modify "custom_domains" table
ALTER TABLE "custom_domains" ALTER COLUMN "txt_record_subdomain" DROP NOT NULL, ALTER COLUMN "txt_record_value" DROP NOT NULL, ALTER COLUMN "status" DROP NOT NULL;

-- +goose Down
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" ALTER COLUMN "status" SET NOT NULL, ALTER COLUMN "txt_record_value" SET NOT NULL, ALTER COLUMN "txt_record_subdomain" SET NOT NULL;
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ALTER COLUMN "status" SET NOT NULL, ALTER COLUMN "txt_record_value" SET NOT NULL, ALTER COLUMN "txt_record_subdomain" SET NOT NULL;
