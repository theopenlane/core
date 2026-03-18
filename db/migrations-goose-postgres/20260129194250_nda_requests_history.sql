-- +goose Up
-- modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" ADD COLUMN "approved_at" timestamptz NULL, ADD COLUMN "approved_by_user_id" character varying NULL, ADD COLUMN "signed_at" timestamptz NULL, ADD COLUMN "document_data_id" character varying NULL;
-- modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ALTER COLUMN "overview" TYPE text;

-- +goose Down
-- reverse: modify "trust_center_setting_history" table
ALTER TABLE "trust_center_setting_history" ALTER COLUMN "overview" TYPE character varying;
-- reverse: modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" DROP COLUMN "document_data_id", DROP COLUMN "signed_at", DROP COLUMN "approved_by_user_id", DROP COLUMN "approved_at";
