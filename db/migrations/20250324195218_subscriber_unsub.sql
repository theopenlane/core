-- Drop index "subscriber_email_owner_id" from table: "subscribers"
DROP INDEX "subscriber_email_owner_id";
-- Modify "subscribers" table
ALTER TABLE "subscribers" ADD COLUMN "unsubscribed" boolean NOT NULL DEFAULT false;
-- Create index "subscriber_email_owner_id" to table: "subscribers"
CREATE UNIQUE INDEX "subscriber_email_owner_id" ON "subscribers" ("email", "owner_id") WHERE ((deleted_at IS NULL) AND (unsubscribed = false));
