-- Drop index "campaign_name_owner_id" from table: "campaigns"
DROP INDEX "campaign_name_owner_id";
-- Create index "campaign_name_owner_id" to table: "campaigns"
CREATE INDEX "campaign_name_owner_id" ON "campaigns" ("name", "owner_id") WHERE (deleted_at IS NULL);
