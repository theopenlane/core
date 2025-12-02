-- Drop index "notification_owner_id" from table: "notifications"
DROP INDEX "notification_owner_id";
-- Drop index "notification_user_id_read_at_owner_id" from table: "notifications"
DROP INDEX "notification_user_id_read_at_owner_id";
-- Modify "notifications" table
ALTER TABLE "notifications" DROP CONSTRAINT "notifications_users_notifications", DROP COLUMN "deleted_at", DROP COLUMN "deleted_by", ALTER COLUMN "user_id" DROP NOT NULL, ADD CONSTRAINT "notifications_users_notifications" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Create index "notification_user_id_read_at_owner_id" to table: "notifications"
CREATE INDEX "notification_user_id_read_at_owner_id" ON "notifications" ("user_id", "read_at", "owner_id");
