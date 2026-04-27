-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'LOW';
-- modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "identity_holder_id" character varying NULL, ADD CONSTRAINT "directory_groups_identity_holders_directory_groups" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "directorygroup_identity_holder_id" to table: "directory_groups"
CREATE INDEX "directorygroup_identity_holder_id" ON "directory_groups" ("identity_holder_id");

-- +goose Down
-- reverse: create index "directorygroup_identity_holder_id" to table: "directory_groups"
DROP INDEX "directorygroup_identity_holder_id";
-- reverse: modify "directory_groups" table
ALTER TABLE "directory_groups" DROP CONSTRAINT "directory_groups_identity_holders_directory_groups", DROP COLUMN "identity_holder_id";
-- reverse: modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';
