-- +goose Up
-- modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "assessment_response_id" character varying NULL, ADD CONSTRAINT "integration_runs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "integrationrun_assessment_response_id_operation_name" to table: "integration_runs"
CREATE UNIQUE INDEX "integrationrun_assessment_response_id_operation_name" ON "integration_runs" ("assessment_response_id", "operation_name") WHERE ((deleted_at IS NULL) AND (assessment_response_id IS NOT NULL));
-- create index "integrationrun_assessment_response_id_started_at" to table: "integration_runs"
CREATE INDEX "integrationrun_assessment_response_id_started_at" ON "integration_runs" ("assessment_response_id", "started_at") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "integrationrun_assessment_response_id_started_at" to table: "integration_runs"
DROP INDEX "integrationrun_assessment_response_id_started_at";
-- reverse: create index "integrationrun_assessment_response_id_operation_name" to table: "integration_runs"
DROP INDEX "integrationrun_assessment_response_id_operation_name";
-- reverse: modify "integration_runs" table
ALTER TABLE "integration_runs" DROP CONSTRAINT "integration_runs_assessment_responses_assessment_response", DROP COLUMN "assessment_response_id";
