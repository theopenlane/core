-- Drop index "evidence_external_uuid_owner_id" from table: "evidences"
DROP INDEX "evidence_external_uuid_owner_id";
-- Create index "evidence_external_uuid_owner_id" to table: "evidences"
CREATE INDEX "evidence_external_uuid_owner_id" ON "evidences" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
