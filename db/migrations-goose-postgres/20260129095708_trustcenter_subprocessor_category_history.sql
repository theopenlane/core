-- +goose Up
-- modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" DROP COLUMN "category", ADD COLUMN "trust_center_subprocessor_kind_name" character varying NULL, ADD COLUMN "trust_center_subprocessor_kind_id" character varying NULL;

-- +goose Down
-- reverse: modify "trust_center_subprocessor_history" table
ALTER TABLE "trust_center_subprocessor_history" DROP COLUMN "trust_center_subprocessor_kind_id", DROP COLUMN "trust_center_subprocessor_kind_name", ADD COLUMN "category" character varying NOT NULL;
