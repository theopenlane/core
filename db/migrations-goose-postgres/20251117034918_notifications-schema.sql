-- +goose Up
-- drop index "notification_owner_id" from table: "notifications"
DROP INDEX "notification_owner_id";
-- drop index "notification_user_id_read_at_owner_id" from table: "notifications"
DROP INDEX "notification_user_id_read_at_owner_id";
-- modify "notifications" table
ALTER TABLE "notifications" DROP COLUMN "deleted_at", DROP COLUMN "deleted_by";
-- create index "notification_user_id_read_at_owner_id" to table: "notifications"
CREATE INDEX "notification_user_id_read_at_owner_id" ON "notifications" ("user_id", "read_at", "owner_id");

-- +goose Down
-- reverse: create index "notification_user_id_read_at_owner_id" to table: "notifications"
DROP INDEX "notification_user_id_read_at_owner_id";
-- reverse: modify "notifications" table
ALTER TABLE "notifications" ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "deleted_at" timestamptz NULL;
-- reverse: drop index "notification_user_id_read_at_owner_id" from table: "notifications"
CREATE INDEX "notification_user_id_read_at_owner_id" ON "notifications" ("user_id", "read_at", "owner_id") WHERE (deleted_at IS NULL);
-- reverse: drop index "notification_owner_id" from table: "notifications"
CREATE INDEX "notification_owner_id" ON "notifications" ("owner_id") WHERE (deleted_at IS NULL);
