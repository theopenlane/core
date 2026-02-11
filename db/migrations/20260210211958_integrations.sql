-- Drop index "integration_owner_id_kind" from table: "integrations"
DROP INDEX "integration_owner_id_kind";
-- Create index "integration_owner_id_kind" to table: "integrations"
CREATE INDEX "integration_owner_id_kind" ON "integrations" ("owner_id", "kind") WHERE (deleted_at IS NULL);
-- Modify "workflow_definitions" table
ALTER TABLE "workflow_definitions" ALTER COLUMN "approval_submission_mode" SET DEFAULT 'AUTO_SUBMIT';
