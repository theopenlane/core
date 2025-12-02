-- +goose Up
-- modify "user_history" table
ALTER TABLE "user_history" ADD COLUMN "last_login_provider" character varying NULL;
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "last_login_provider" character varying NULL;

-- +goose Down
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "last_login_provider";
-- reverse: modify "user_history" table
ALTER TABLE "user_history" DROP COLUMN "last_login_provider";
