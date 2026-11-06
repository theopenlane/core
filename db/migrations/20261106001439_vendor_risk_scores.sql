-- Modify "vendor_risk_scores" table
ALTER TABLE "vendor_risk_scores" ADD COLUMN "source_definition_id" character varying NULL, ADD COLUMN "source_definition_version" character varying NULL, ADD COLUMN "source_instance_id" character varying NULL, ADD COLUMN "managed_by" character varying NULL, ADD COLUMN "integration_run_id" character varying NULL, ADD COLUMN "source_key" character varying NULL, ADD COLUMN "external_id" character varying NULL;
-- Create index "vendorriskscore_external_id_owner_id" to table: "vendor_risk_scores"
CREATE UNIQUE INDEX "vendorriskscore_external_id_owner_id" ON "vendor_risk_scores" ("external_id", "owner_id") WHERE (deleted_at IS NULL);
-- Create index "vendorriskscore_integration_run_id" to table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_integration_run_id" ON "vendor_risk_scores" ("integration_run_id");
-- Create index "vendorriskscore_source_instance_id" to table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_source_instance_id" ON "vendor_risk_scores" ("source_instance_id");
-- Create index "vendorriskscore_source_key" to table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_source_key" ON "vendor_risk_scores" ("source_key");
-- Modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "vendor_risk_score_integration_runs" character varying NULL, ADD CONSTRAINT "integration_runs_vendor_risk_scores_integration_runs" FOREIGN KEY ("vendor_risk_score_integration_runs") REFERENCES "vendor_risk_scores" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
