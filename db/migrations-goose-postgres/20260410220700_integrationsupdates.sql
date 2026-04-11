-- +goose Up
-- modify "identity_holders" table
ALTER TABLE "identity_holders" ADD COLUMN "email_aliases" jsonb NULL;
-- drop index "vulnerability_cve_id_owner_id" from table: "vulnerabilities"
DROP INDEX "vulnerability_cve_id_owner_id";
-- create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
CREATE INDEX "vulnerability_cve_id_owner_id" ON "vulnerabilities" ("cve_id", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "vulnerability_cve_id_owner_id" to table: "vulnerabilities"
DROP INDEX "vulnerability_cve_id_owner_id";
-- reverse: drop index "vulnerability_cve_id_owner_id" from table: "vulnerabilities"
CREATE UNIQUE INDEX "vulnerability_cve_id_owner_id" ON "vulnerabilities" ("cve_id", "owner_id") WHERE (deleted_at IS NULL);
-- reverse: modify "identity_holders" table
ALTER TABLE "identity_holders" DROP COLUMN "email_aliases";
