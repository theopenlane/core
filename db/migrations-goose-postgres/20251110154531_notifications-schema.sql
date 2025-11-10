-- +goose Up
-- modify "invites" table
ALTER TABLE "invites" ADD COLUMN "ownership_transfer" boolean NULL DEFAULT false;
-- modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ALTER COLUMN "font" SET DEFAULT 'HELVETICA';
-- modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ALTER COLUMN "font" SET DEFAULT 'HELVETICA';
-- create "notifications" table
CREATE TABLE "notifications" ("id" character varying NOT NULL, "created_at" timestamptz NULL, "updated_at" timestamptz NULL, "created_by" character varying NULL, "updated_by" character varying NULL, "deleted_at" timestamptz NULL, "deleted_by" character varying NULL, "tags" jsonb NULL, "title" character varying NOT NULL, "body" character varying NOT NULL, "data" jsonb NULL, "notification_type" character varying NOT NULL, "object_type" character varying NOT NULL, "read_at" timestamptz NULL, "channels" jsonb NULL, "owner_id" character varying NULL, "user_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "notifications_organizations_notifications" FOREIGN KEY ("owner_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "notifications_users_notifications" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- create index "notification_owner_id" to table: "notifications"
CREATE INDEX "notification_owner_id" ON "notifications" ("owner_id") WHERE (deleted_at IS NULL);
-- create index "notification_user_id_read_at_owner_id" to table: "notifications"
CREATE INDEX "notification_user_id_read_at_owner_id" ON "notifications" ("user_id", "read_at", "owner_id");

-- +goose Down
-- reverse: create index "notification_user_id_read_at_owner_id" to table: "notifications"
DROP INDEX "notification_user_id_read_at_owner_id";
-- reverse: create index "notification_owner_id" to table: "notifications"
DROP INDEX "notification_owner_id";
-- reverse: create "notifications" table
DROP TABLE "notifications";
-- reverse: modify "trust_center_watermark_configs" table
ALTER TABLE "trust_center_watermark_configs" ALTER COLUMN "font" SET DEFAULT 'arial';
-- reverse: modify "trust_center_watermark_config_history" table
ALTER TABLE "trust_center_watermark_config_history" ALTER COLUMN "font" SET DEFAULT 'arial';
-- reverse: modify "invites" table
ALTER TABLE "invites" DROP COLUMN "ownership_transfer";
