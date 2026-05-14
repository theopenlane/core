-- +goose Up
-- modify "entities" table
ALTER TABLE "entities" ADD COLUMN "logo_remote_url" character varying NULL;

-- +goose Down
-- reverse: modify "entities" table
ALTER TABLE "entities" DROP COLUMN "logo_remote_url";
