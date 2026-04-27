-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'LOW';
-- modify "directory_memberships" table
ALTER TABLE "directory_memberships" ADD COLUMN "identity_holder_id" character varying NULL, ADD CONSTRAINT "directory_memberships_identity_holders_directory_memberships" FOREIGN KEY ("identity_holder_id") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- create index "directorymembership_identity_holder_id" to table: "directory_memberships"
CREATE INDEX "directorymembership_identity_holder_id" ON "directory_memberships" ("identity_holder_id");

-- +goose Down
-- reverse: create index "directorymembership_identity_holder_id" to table: "directory_memberships"
DROP INDEX "directorymembership_identity_holder_id";
-- reverse: modify "directory_memberships" table
ALTER TABLE "directory_memberships" DROP CONSTRAINT "directory_memberships_identity_holders_directory_memberships", DROP COLUMN "identity_holder_id";
-- reverse: modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';
