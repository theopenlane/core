-- +goose Up
-- modify "integration_runs" table
ALTER TABLE "integration_runs" DROP COLUMN "remediation_integration_runs";
-- create "remediation_integration_runs" table
CREATE TABLE "remediation_integration_runs" ("remediation_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "integration_run_id"), CONSTRAINT "remediation_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_integration_runs_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create index "remediation_integration_runs_integration_run_id_idx" to table: "remediation_integration_runs"
CREATE INDEX "remediation_integration_runs_integration_run_id_idx" ON "remediation_integration_runs" ("integration_run_id");

-- +goose Down
-- reverse: create index "remediation_integration_runs_integration_run_id_idx" to table: "remediation_integration_runs"
DROP INDEX "remediation_integration_runs_integration_run_id_idx";
-- reverse: create "remediation_integration_runs" table
DROP TABLE "remediation_integration_runs";
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "remediation_integration_runs" character varying NULL;
