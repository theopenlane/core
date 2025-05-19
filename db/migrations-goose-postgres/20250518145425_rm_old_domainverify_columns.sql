-- +goose Up
-- modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" DROP COLUMN "txt_record_subdomain", DROP COLUMN "txt_record_value", DROP COLUMN "status";
-- modify "custom_domains" table
ALTER TABLE "custom_domains" DROP COLUMN "txt_record_subdomain", DROP COLUMN "txt_record_value", DROP COLUMN "status";

-- +goose Down
-- reverse: modify "custom_domains" table
ALTER TABLE "custom_domains" ADD COLUMN "status" character varying NOT NULL DEFAULT 'PENDING', ADD COLUMN "txt_record_value" character varying NOT NULL, ADD COLUMN "txt_record_subdomain" character varying NOT NULL DEFAULT '_olverify';
-- reverse: modify "custom_domain_history" table
ALTER TABLE "custom_domain_history" ADD COLUMN "status" character varying NOT NULL DEFAULT 'PENDING', ADD COLUMN "txt_record_value" character varying NOT NULL, ADD COLUMN "txt_record_subdomain" character varying NOT NULL DEFAULT '_olverify';
