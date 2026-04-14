-- +goose Up
-- drop index "evidence_external_uuid_owner_id" from table: "evidences"
DROP INDEX "evidence_external_uuid_owner_id";
-- create index "evidence_external_uuid_owner_id" to table: "evidences"
CREATE INDEX "evidence_external_uuid_owner_id" ON "evidences" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "evidence_external_uuid_owner_id" to table: "evidences"
DROP INDEX "evidence_external_uuid_owner_id";
-- reverse: drop index "evidence_external_uuid_owner_id" from table: "evidences"
CREATE UNIQUE INDEX "evidence_external_uuid_owner_id" ON "evidences" ("external_uuid", "owner_id") WHERE (deleted_at IS NULL);
