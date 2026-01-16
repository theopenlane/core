-- Create index "trustcenterentityhistory_history_time" to table: "trust_center_entity_history"
CREATE INDEX "trustcenterentityhistory_history_time" ON "trust_center_entity_history" ("history_time");
-- Modify "workflow_assignment_history" table
ALTER TABLE "workflow_assignment_history" ADD COLUMN "approval_metadata" jsonb NULL, ADD COLUMN "rejection_metadata" jsonb NULL, ADD COLUMN "invalidation_metadata" jsonb NULL;
