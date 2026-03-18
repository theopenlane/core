-- Modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" ADD COLUMN "approved_at" timestamptz NULL, ADD COLUMN "approved_by_user_id" character varying NULL, ADD COLUMN "signed_at" timestamptz NULL, ADD COLUMN "document_data_id" character varying NULL;
-- Modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ALTER COLUMN "overview" TYPE text;
