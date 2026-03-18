-- +goose Up
-- modify "assets" table
ALTER TABLE "assets" ADD COLUMN "display_name" character varying NULL;

-- +goose Down
-- reverse: modify "assets" table
ALTER TABLE "assets" DROP COLUMN "display_name";
