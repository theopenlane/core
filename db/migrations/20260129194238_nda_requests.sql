-- Modify "trust_center_settings" table
ALTER TABLE "trust_center_settings" ALTER COLUMN "overview" TYPE text;
-- Modify "trust_center_nda_requests" table
ALTER TABLE "trust_center_nda_requests" ADD COLUMN "approved_at" timestamptz NULL, ADD COLUMN "approved_by_user_id" character varying NULL, ADD COLUMN "signed_at" timestamptz NULL, ADD COLUMN "document_data_id" character varying NULL, ADD CONSTRAINT "trust_center_nda_requests_document_data_document" FOREIGN KEY ("document_data_id") REFERENCES "document_data" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
