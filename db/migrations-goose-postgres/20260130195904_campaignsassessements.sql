-- +goose Up
-- drop index "assessmentresponse_assessment_id_email" from table: "assessment_responses"
DROP INDEX "assessmentresponse_assessment_id_email";
-- create index "assessmentresponse_assessment_id_email_is_test" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_email_is_test" ON "assessment_responses" ("assessment_id", "email", "is_test") WHERE ((deleted_at IS NULL) AND (campaign_id IS NULL));
-- create index "assessmentresponse_campaign_id_assessment_id_email_is_test" to table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_campaign_id_assessment_id_email_is_test" ON "assessment_responses" ("campaign_id", "assessment_id", "email", "is_test") WHERE ((deleted_at IS NULL) AND (campaign_id IS NOT NULL));

-- +goose Down
-- reverse: create index "assessmentresponse_campaign_id_assessment_id_email_is_test" to table: "assessment_responses"
DROP INDEX "assessmentresponse_campaign_id_assessment_id_email_is_test";
-- reverse: create index "assessmentresponse_assessment_id_email_is_test" to table: "assessment_responses"
DROP INDEX "assessmentresponse_assessment_id_email_is_test";
-- reverse: drop index "assessmentresponse_assessment_id_email" from table: "assessment_responses"
CREATE UNIQUE INDEX "assessmentresponse_assessment_id_email" ON "assessment_responses" ("assessment_id", "email") WHERE (deleted_at IS NULL);
