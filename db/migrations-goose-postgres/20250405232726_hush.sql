-- +goose Up
-- modify "hush_history" table
ALTER TABLE "hush_history" ADD COLUMN "owner_id" character varying NULL;
-- modify "hushes" table
ALTER TABLE "hushes" ADD COLUMN "owner_id" character varying NULL, ADD CONSTRAINT "hushes_organizations_secrets" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "hushes" table
ALTER TABLE "hushes" DROP CONSTRAINT "hushes_organizations_secrets", DROP COLUMN "owner_id";
-- reverse: modify "hush_history" table
ALTER TABLE "hush_history" DROP COLUMN "owner_id";
