-- +goose Up
-- drop index "subscriber_email_owner_id" from table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "unsubscribed" boolean NOT NULL DEFAULT false;
-- create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false));

-- +goose Down
-- reverse: create index "subscriber_email_owner_id" to table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- reverse: modify "subscribers" table
ALTER TABLE "subscribers" DROP COLUMN "unsubscribed";
-- reverse: drop index "subscriber_email_owner_id" from table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE (deleted_at IS NULL);
