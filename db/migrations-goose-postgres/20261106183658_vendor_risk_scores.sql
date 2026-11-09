-- +goose Up
-- modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "source_key" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- create index "vendorriskscore_integration_run_id" to table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_integration_run_id" ON "vendor_risk_scores" ("integration_run_id");
-- create index "vendorriskscore_source_instance_id" to table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_source_instance_id" ON "vendor_risk_scores" ("source_instance_id");
-- create index "vendorriskscore_source_key" to table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_source_key" ON "vendor_risk_scores" ("source_key");
-- modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "vendor_risk_score_integration_runs" character varying NULL, ADD CONSTRAINT "integration_runs_vendor_risk_scores_integration_runs" FOREIGN KEY ("vendor_risk_score_integration_runs") REFERENCES "vendor_risk_scores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" DROP CONSTRAINT "integration_runs_vendor_risk_scores_integration_runs", DROP COLUMN "vendor_risk_score_integration_runs";
-- reverse: create index "vendorriskscore_source_key" to table: "vendor_risk_scores"
DROP INDEX "vendorriskscore_source_key";
-- reverse: create index "vendorriskscore_source_instance_id" to table: "vendor_risk_scores"
DROP INDEX "vendorriskscore_source_instance_id";
-- reverse: create index "vendorriskscore_integration_run_id" to table: "vendor_risk_scores"
DROP INDEX "vendorriskscore_integration_run_id";
-- reverse: modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" DROP COLUMN "external_id", DROP COLUMN "source_key", DROP COLUMN "integration_run_id", DROP COLUMN "managed_by", DROP COLUMN "source_instance_id", DROP COLUMN "source_definition_version", DROP COLUMN "source_definition_id";
