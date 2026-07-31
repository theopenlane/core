-- +goose Up
-- drop index "asset_name_owner_id" from table: "assets"
DROP INDEX "asset_name_owner_id";

-- +goose Down
-- reverse: drop index "asset_name_owner_id" from table: "assets"
CREATE UNIQUE INDEX "asset_name_owner_id" ON "assets" ("name", "owner_id") WHERE (deleted_at IS NULL);
