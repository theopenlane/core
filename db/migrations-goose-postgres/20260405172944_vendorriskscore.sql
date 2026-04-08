-- +goose Up
-- modify "directory_accounts" table
ALTER TABLE "directory_accounts" ADD COLUMN "primary_source" boolean NOT NULL DEFAULT false;
-- modify "integrations" table
ALTER TABLE "integrations" ADD COLUMN "primary_directory" boolean NOT NULL DEFAULT false;
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "risk_score_coverage" bigint NULL;
-- create "vendor_scoring_configs" table
CREATE TABLE "vendor_scoring_configs" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "questions" jsonb NOT NULL, "scoring_mode" character varying NOT NULL DEFAULT 'ANSWERED_ONLY', "risk_thresholds" jsonb NOT NULL, "owner_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "vendor_scoring_configs_organizations_vendor_scoring_configs" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "vendorscoringconfig_owner_id" to table: "vendor_scoring_configs"
CREATE INDEX "vendorscoringconfig_owner_id" ON "vendor_scoring_configs" ("owner_id") WHERE (deleted_at IS NULL);
-- create "vendor_risk_scores" table
CREATE TABLE "vendor_risk_scores" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "question_key" character varying NOT NULL, "question_name" character varying NOT NULL, "question_description" character varying NULL, "question_category" character varying NOT NULL, "answer_type" character varying NOT NULL, "impact" character varying NOT NULL, "likelihood" character varying NOT NULL, "score" double precision NOT NULL DEFAULT 0, "answer" character varying NULL, "notes" character varying NULL, "assessment_response_vendor_risk_scores" character varying NULL, "entity_vendor_risk_scores" character varying NULL, "owner_id" character varying NULL, "vendor_scoring_config_id" character varying NULL, "entity_id" character varying NOT NULL, "assessment_response_id" character varying NULL, "vendor_scoring_config_vendor_risk_scores" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "vendor_risk_scores_assessment_responses_assessment_response" FOREIGN KEY ("assessment_response_id") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "vendor_risk_scores_assessment_responses_vendor_risk_scores" FOREIGN KEY ("assessment_response_vendor_risk_scores") REFERENCES "assessment_responses" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "vendor_risk_scores_entities_entity" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "vendor_risk_scores_entities_vendor_risk_scores" FOREIGN KEY ("entity_vendor_risk_scores") REFERENCES "entities" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "vendor_risk_scores_organizations_vendor_risk_scores" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_risk_scores" FOREIGN KEY ("vendor_scoring_config_vendor_risk_scores") REFERENCES "vendor_scoring_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "vendor_risk_scores_vendor_scoring_configs_vendor_scoring_config" FOREIGN KEY ("vendor_scoring_config_id") REFERENCES "vendor_scoring_configs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "vendorriskscore_owner_id" to table: "vendor_risk_scores"
CREATE INDEX "vendorriskscore_owner_id" ON "vendor_risk_scores" ("owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "vendorriskscore_owner_id" to table: "vendor_risk_scores"
DROP INDEX "vendorriskscore_owner_id";
-- reverse: create "vendor_risk_scores" table
DROP TABLE "vendor_risk_scores";
-- reverse: create index "vendorscoringconfig_owner_id" to table: "vendor_scoring_configs"
DROP INDEX "vendorscoringconfig_owner_id";
-- reverse: create "vendor_scoring_configs" table
DROP TABLE "vendor_scoring_configs";
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "risk_score_coverage";
-- reverse: modify "integrations" table
ALTER TABLE "integrations" DROP COLUMN "primary_directory";
-- reverse: modify "directory_accounts" table
ALTER TABLE "directory_accounts" DROP COLUMN "primary_source";
