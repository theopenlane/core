-- Modify "user_history" table
ALTER TABLE "user_history" ADD COLUMN "last_login_provider" character varying NULL;
-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "last_login_provider" character varying NULL;
