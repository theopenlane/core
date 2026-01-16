-- +goose Up
-- create index "trustcenterentityhistory_history_time" to table: "trust_center_entity_history"
CREATE INDEX "trustcenterentityhistory_history_time" ON "trust_center_entity_history" ("history_time");
-- modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "approval_metadata" jsonb NULL, ADD COLUMN "rejection_metadata" jsonb NULL, ADD COLUMN "invalidation_metadata" jsonb NULL;

-- +goose Down
-- reverse: modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" DROP COLUMN "invalidation_metadata", DROP COLUMN "rejection_metadata", DROP COLUMN "approval_metadata";
-- reverse: create index "trustcenterentityhistory_history_time" to table: "trust_center_entity_history"
DROP INDEX "trustcenterentityhistory_history_time";
