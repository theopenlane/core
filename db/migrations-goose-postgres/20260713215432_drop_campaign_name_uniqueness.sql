-- +goose Up
-- drop index "campaign_name_owner_id" from table: "campaigns"
DROP INDEX "campaign_name_owner_id";
-- create index "campaign_name_owner_id" to table: "campaigns"
CREATE INDEX "campaign_name_owner_id" ON "campaigns" ("name", "owner_id") WHERE (deleted_at IS NULL);

-- +goose Down
-- reverse: create index "campaign_name_owner_id" to table: "campaigns"
DROP INDEX "campaign_name_owner_id";
-- reverse: drop index "campaign_name_owner_id" from table: "campaigns"
CREATE UNIQUE INDEX "campaign_name_owner_id" ON "campaigns" ("name", "owner_id") WHERE (deleted_at IS NULL);
