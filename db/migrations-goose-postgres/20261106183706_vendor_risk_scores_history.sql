-- +goose Up
-- modify "vendor_risk_score_history" table
ALTER TABLE "vendor_risk_score_history" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "source_key" character varying NULL, ADD COLUMN "external_id" character varying NULL;

-- +goose Down
-- reverse: modify "vendor_risk_score_history" table
ALTER TABLE "vendor_risk_score_history" DROP COLUMN "external_id", DROP COLUMN "source_key", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
