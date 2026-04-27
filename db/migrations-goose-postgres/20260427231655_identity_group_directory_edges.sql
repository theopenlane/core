-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'LOW';
-- modify "directory_groups" table
ALTER TABLE "directory_groups" ADD COLUMN "identity_holder_directory_groups" character varying NULL, ADD CONSTRAINT "directory_groups_identity_holders_directory_groups" FOREIGN KEY ("identity_holder_directory_groups") REFERENCES "identity_holders" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "directory_groups" table
ALTER TABLE "directory_groups" DROP CONSTRAINT "directory_groups_identity_holders_directory_groups", DROP COLUMN "identity_holder_directory_groups";
-- reverse: modify "entities" table
ALTER TABLE "entities" ALTER COLUMN "tier" SET DEFAULT 'STANDARD';
