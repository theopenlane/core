-- +goose Up
-- modify "asset_history" table
ALTER TABLE "asset_history" ADD COLUMN "display_name" character varying NULL;

-- +goose Down
-- reverse: modify "asset_history" table
ALTER TABLE "asset_history" DROP COLUMN "display_name";
