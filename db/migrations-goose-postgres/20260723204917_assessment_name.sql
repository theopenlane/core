-- +goose Up
-- drop index "assessment_name_owner_id" from table: "assessments"
DROP INDEX "assessment_name_owner_id";
-- create index "assessment_name_owner_id" to table: "assessments"
CREATE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "assessment_name_owner_id" to table: "assessments"
DROP INDEX "assessment_name_owner_id";
-- reverse: drop index "assessment_name_owner_id" from table: "assessments"
CREATE UNIQUE INDEX "assessment_name_owner_id" ON "assessments" ("name", "owner_id") WHERE (deleted_at IS NULL);
