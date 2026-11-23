-- +goose Up
-- create "remediation_integration_runs" table
CREATE TABLE "remediation_integration_runs" ("remediation_id" character varying NOT NULL, "integration_run_id" character varying NOT NULL, PRIMARY KEY ("remediation_id", "integration_run_id"), CONSTRAINT "remediation_integration_runs_integration_run_id" FOREIGN KEY ("integration_run_id") REFERENCES "integration_runs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "remediation_integration_runs_remediation_id" FOREIGN KEY ("remediation_id") REFERENCES "remediations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- create index "remediation_integration_runs_integration_run_id_idx" to table: "remediation_integration_runs"
CREATE INDEX "remediation_integration_runs_integration_run_id_idx" ON "remediation_integration_runs" ("integration_run_id");
-- copy existing remediation links into "remediation_integration_runs"
INSERT INTO "remediation_integration_runs" ("remediation_id", "integration_run_id") SELECT "remediation_integration_runs", "id" FROM "integration_runs" WHERE "remediation_integration_runs" IS NOT NULL;
-- modify "integration_runs" table
ALTER TABLE "integration_runs" DROP COLUMN "remediation_integration_runs";

-- +goose Down
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "remediation_integration_runs" character varying NULL;
-- restore one remediation link per run
UPDATE "integration_runs" SET "remediation_integration_runs" = links."remediation_id" FROM (SELECT DISTINCT ON ("integration_run_id") "integration_run_id", "remediation_id" FROM "remediation_integration_runs" ORDER BY "integration_run_id", "remediation_id") AS links WHERE "integration_runs"."id" = links."integration_run_id";
-- reverse: create index "remediation_integration_runs_integration_run_id_idx" to table: "remediation_integration_runs"
DROP INDEX "remediation_integration_runs_integration_run_id_idx";
-- reverse: create "remediation_integration_runs" table
DROP TABLE "remediation_integration_runs";
