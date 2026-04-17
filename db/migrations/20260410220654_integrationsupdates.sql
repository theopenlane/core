-- Modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "email_aliases" jsonb NULL;
-- Drop index "vulnerability_cve_id_owner_id" from table: "vulnerabilities"
DROP INDEX "vulnerability_cve_id_owner_id";
-- Create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
CREATE INDEX "vulnerability_cve_id_owner_id" ON "vulnerabilities" ("cve_id", "owner_id") WHERE (deleted_at IS NULL);
