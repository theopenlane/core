-- +goose Up
-- modify "evidence_history" table
ALTER TABLE "evidence_history" ADD COLUMN "auditor_reference_id" character varying NULL;

-- +goose Down
-- reverse: modify "evidence_history" table
ALTER TABLE "evidence_history" DROP COLUMN "auditor_reference_id";
