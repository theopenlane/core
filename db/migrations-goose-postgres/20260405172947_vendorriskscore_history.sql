-- +goose Up
-- modify "directory_account_history" table
ALTER TABLE "directory_account_history" ADD COLUMN "primary_source" boolean NOT NULL DEFAULT false;
-- modify "entity_history" table
ALTER TABLE "entity_history" ADD COLUMN "risk_score_coverage" bigint NULL;
-- modify "integration_history" table
ALTER TABLE "integration_history" ADD COLUMN "primary_directory" boolean NOT NULL DEFAULT false;
-- create "vendor_risk_score_history" table
CREATE TABLE "vendor_risk_score_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "question_key" character varying NOT NULL, "question_name" character varying NOT NULL, "question_description" character varying NULL, "question_category" character varying NOT NULL, "answer_type" character varying NOT NULL, "impact" character varying NOT NULL, "likelihood" character varying NOT NULL, "score" double precision NOT NULL DEFAULT 0, "answer" character varying NULL, "notes" character varying NULL, "vendor_scoring_config_id" character varying NULL, "entity_id" character varying NOT NULL, "assessment_response_id" character varying NULL, PRIMARY KEY ("id"));
-- create index "vendorriskscorehistory_history_time" to table: "vendor_risk_score_history"
CREATE INDEX "vendorriskscorehistory_history_time" ON "vendor_risk_score_history" ("history_time");
-- create "vendor_scoring_config_history" table
CREATE TABLE "vendor_scoring_config_history" ("id" character varying NOT NULL, "history_time" timestamptz NOT NULL, "ref" character varying NULL, "operation" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "owner_id" character varying NULL, "questions" jsonb NOT NULL, "scoring_mode" character varying NOT NULL DEFAULT 'ANSWERED_ONLY', "risk_thresholds" jsonb NOT NULL, PRIMARY KEY ("id"));
-- create index "vendorscoringconfighistory_history_time" to table: "vendor_scoring_config_history"
CREATE INDEX "vendorscoringconfighistory_history_time" ON "vendor_scoring_config_history" ("history_time");

-- +goose Down
-- reverse: create index "vendorscoringconfighistory_history_time" to table: "vendor_scoring_config_history"
DROP INDEX "vendorscoringconfighistory_history_time";
-- reverse: create "vendor_scoring_config_history" table
DROP TABLE "vendor_scoring_config_history";
-- reverse: create index "vendorriskscorehistory_history_time" to table: "vendor_risk_score_history"
DROP INDEX "vendorriskscorehistory_history_time";
-- reverse: create "vendor_risk_score_history" table
DROP TABLE "vendor_risk_score_history";
-- reverse: modify "integration_history" table
ALTER TABLE "integration_history" DROP COLUMN "primary_directory";
-- reverse: modify "entity_history" table
ALTER TABLE "entity_history" DROP COLUMN "risk_score_coverage";
-- reverse: modify "directory_account_history" table
ALTER TABLE "directory_account_history" DROP COLUMN "primary_source";
