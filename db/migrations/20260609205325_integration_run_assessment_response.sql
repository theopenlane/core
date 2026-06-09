-- Modify "integration_runs" table
ALTER TABLE "integration_runs" ADD COLUMN "assessment_response_id" character varying NULL, ADD CONSTRAINT "integration_runs_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "integrationrun_assessment_response_id_operation_name" to table: "integration_runs"
CREATE UNIQUE INDEX "integrationrun_assessment_response_id_operation_name" ON "integration_runs" ("assessment_response_id", "operation_name") WHERE ((deleted_at IS NULL) AND (assessment_response_id IS NOT NULL));
-- Create index "integrationrun_assessment_response_id_started_at" to table: "integration_runs"
CREATE INDEX "integrationrun_assessment_response_id_started_at" ON "integration_runs" ("assessment_response_id", "started_at") WHERE (deleted_at IS NULL);
