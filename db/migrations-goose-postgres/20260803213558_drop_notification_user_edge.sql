-- +goose Up
-- modify "notifications" table
ALTER TABLE "notifications" DROP CONSTRAINT "notifications_users_notifications";

-- +goose Down
-- reverse: modify "notifications" table
ALTER TABLE "notifications" ADD CONSTRAINT "notifications_users_notifications" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
