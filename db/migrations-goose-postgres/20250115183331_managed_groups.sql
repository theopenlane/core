-- +goose Up
-- modify "group_history" table
ALTER TABLE "group_history" ADD COLUMN "is_managed" boolean NULL DEFAULT false;
-- modify "groups" table
ALTER TABLE "groups" ADD COLUMN "is_managed" boolean NULL DEFAULT false;

-- +goose Down
-- reverse: modify "groups" table
ALTER TABLE "groups" DROP COLUMN "is_managed";
-- reverse: modify "group_history" table
ALTER TABLE "group_history" DROP COLUMN "is_managed";
