-- +goose Up
-- modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" ADD COLUMN "file_id" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_nda_request_history" table
ALTER TABLE "trust_center_nda_request_history" DROP COLUMN "file_id";
