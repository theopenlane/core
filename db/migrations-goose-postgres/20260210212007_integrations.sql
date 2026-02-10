-- +goose Up
-- drop index "integration_owner_id_kind" from table: "integrations"
DROP INDEX "integration_owner_id_kind";
-- create index "integration_owner_id_kind" to table: "integrations"
CREATE INDEX "integration_owner_id_kind" ON "integrations" ("owner_id", "kind") WHERE (deleted_at IS NULL);
-- modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ALTER COLUMN "approval_submission_mode" SET DEFAULT 'AUTO_SUBMIT';

-- +goose Down
-- reverse: modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ALTER COLUMN "approval_submission_mode" SET DEFAULT 'MANUAL_SUBMIT';
-- reverse: create index "integration_owner_id_kind" to table: "integrations"
DROP INDEX "integration_owner_id_kind";
-- reverse: drop index "integration_owner_id_kind" from table: "integrations"
CREATE UNIQUE INDEX "integration_owner_id_kind" ON "integrations" ("owner_id", "kind") WHERE (deleted_at IS NULL);
